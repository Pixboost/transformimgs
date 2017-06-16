package img

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
)

//Number of seconds that will be written to max-age HTTP header
var CacheTTL int

//Reads image from a given source
type ImgReader interface {
	//Reads image from the url.
	//Returns byte array of the image or
	//error.
	Read(url string) ([]byte, error)
}

//Processes images applying different transformations.
type ImgProcessor interface {
	//Resize given image.
	//Form of the the size string is
	//width'x'height. Any dimension could be skipped.
	//For example:
	//* 300x200
	//* 300 - only width
	//* x200 - only height
	Resize(data []byte, size string) ([]byte, error)

	//Resize given image fitting it to a given size.
	//Form of the the size string is width'x'height.
	//For example, 300x400
	FitToSize(data []byte, size string) ([]byte, error)

	//Optimises given image to reduce size.
	Optimise(data []byte) ([]byte, error)
}

type Service struct {
	Reader      ImgReader
	Processor   ImgProcessor
	OpChans     []chan *Operation
	currProc    int
	currProcMux sync.Mutex
}

type ImgOp func([]byte) ([]byte, error)
type ImgResizeOp func([]byte, string) ([]byte, error)

type Operation struct {
	ImgOp        ImgOp
	ImgResizeOp  ImgResizeOp
	In           []byte
	Size         string
	Resp         http.ResponseWriter
	Result       []byte
	FinishedCond *sync.Cond
	Finished     bool
	Err          error
}

func NewService(r ImgReader, p ImgProcessor, procNum int) (*Service, error) {
	if procNum <= 0 {
		return nil, fmt.Errorf("procNum must be positive, but got [%d]", procNum)
	}

	fmt.Printf("Creating new service with [%d] number of processors\n", procNum)

	srv := &Service{
		Reader:    r,
		Processor: p,
		OpChans:   make([]chan *Operation, procNum),
	}

	for i := 0; i < procNum; i++ {
		c := make(chan *Operation)
		go proc(c)
		srv.OpChans[i] = c
	}
	srv.currProc = 0

	return srv, nil
}

func (r *Service) GetRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/img/resize", http.HandlerFunc(r.ResizeUrl))
	router.HandleFunc("/img/fit", http.HandlerFunc(r.FitToSizeUrl))
	router.HandleFunc("/img/asis", http.HandlerFunc(r.AsIs))
	router.HandleFunc("/img", http.HandlerFunc(r.OptimiseUrl))

	return router
}

// swagger:operation GET /img optimiseImage
//
// Optimises image from the given url.
//
// ---
// tags:
// - images
// produces:
// - image/png
// - image/jpeg
// parameters:
// - name: url
//   required: true
//   in: query
//   type: string
//   description: url of the original image
// responses:
//   '200':
//     description: Optimised image in the same format as original.
func (r *Service) OptimiseUrl(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getQueryParam(req.URL, "url")
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}

	log.Printf("Optimising image %s\n", imgUrl)

	input, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}

	r.execOp(&Operation{
		ImgOp: r.Processor.Optimise,
		In:    input,
		Resp:  resp,
	})
}

// swagger:operation GET /img/resize resizeImage
//
// Resize image with preserving aspect ratio and optimizes it.
// If you need the exact size then use /fit operation.
//
// ---
// tags:
// - images
// produces:
// - image/png
// - image/jpeg
// parameters:
// - name: url
//   required: true
//   in: query
//   type: string
//   description: url of the original image
// - name: size
//   required: true
//   in: query
//   type: string
//   description: |
//    size of the image in the response. Should be in format 'width'x'height', e.g. 200x300
//    Only width or height could be passed, e.g 200, x300.
//
// responses:
//   '200':
//     description: Resized image in the same format as original.
func (r *Service) ResizeUrl(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getQueryParam(req.URL, "url")
	size := getQueryParam(req.URL, "size")
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}
	if len(size) == 0 {
		http.Error(resp, "size param is required", http.StatusBadRequest)
		return
	}

	log.Printf("Resizing image %s to %s\n", imgUrl, size)

	input, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}

	r.execOp(&Operation{
		ImgResizeOp: r.Processor.Resize,
		In:          input,
		Size:        size,
		Resp:        resp,
	})
}

// swagger:operation GET /img/fit fitImage
//
// Resize image to the exact size and optimizes it.
// Will resize image and crop it to the size.
// If you need to resize image with preserved aspect ratio then use /img/resize endpoint.
//
// ---
// tags:
// - images
// produces:
// - image/png
// - image/jpeg
// parameters:
// - name: url
//   required: true
//   in: query
//   type: string
//   description: url of the original image
// - name: size
//   required: true
//   in: query
//   type: string
//   pattern: \d{1,4}x\d{1,4}
//   description: |
//    size of the image in the response. Should be in the format 'width'x'height', e.g. 200x300
//
// responses:
//   '200':
//     description: Resized image in the same format as original.
func (r *Service) FitToSizeUrl(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getQueryParam(req.URL, "url")
	size := getQueryParam(req.URL, "size")
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}
	if len(size) == 0 {
		http.Error(resp, "size param is required", http.StatusBadRequest)
		return
	}
	if match, err := regexp.MatchString(`^\d*[x]\d*$`, size); !match || err != nil {
		if err != nil {
			log.Printf("Error while matching size: %s\n", err.Error())
		}
		http.Error(resp, "size param should be in format WxH", http.StatusBadRequest)
		return
	}

	log.Printf("Fit image %s to size %s\n", imgUrl, size)

	input, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}

	r.execOp(&Operation{
		ImgResizeOp: r.Processor.FitToSize,
		In:          input,
		Size:        size,
		Resp:        resp,
	})
}

// swagger:operation GET /img/asis asisImage
//
// Respond with original image without any modifications
//
// ---
// tags:
// - images
// produces:
// - image/png
// - image/jpeg
// parameters:
// - name: url
//   required: true
//   in: query
//   type: string
//   description: url of the image
//
// responses:
//   '200':
//     description: Requested image.
func (r *Service) AsIs(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getQueryParam(req.URL, "url")
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}

	log.Printf("Requested image %s as is\n", imgUrl)

	result, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	} else {
		r.execOp(&Operation{
			Result: result,
			Resp:   resp,
		})
	}
}

func (r *Service) execOp(op *Operation) {
	op.FinishedCond = sync.NewCond(&sync.Mutex{})

	//Adding operation to the next channel
	r.currProcMux.Lock()
	r.currProc++
	if r.currProc == len(r.OpChans) {
		r.currProc = 0
	}
	procIdx := r.currProc
	r.currProcMux.Unlock()

	r.OpChans[procIdx] <- op

	//Waiting for operation to finish
	op.FinishedCond.L.Lock()
	for !op.Finished {
		op.FinishedCond.Wait()
	}
	op.FinishedCond.L.Unlock()

	writeResult(op)
}

//Adds Content-Length and Cache-Control headers
func addHeaders(resp http.ResponseWriter, body []byte) {
	resp.Header().Add("Content-Length", strconv.Itoa(len(body)))
	resp.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", CacheTTL))
}

func getQueryParam(url *url.URL, name string) string {
	if len(url.Query()[name]) == 1 {
		return url.Query()[name][0]
	}
	return ""
}

func proc(opChan chan *Operation) {
	for op := range opChan {
		if op.Result == nil {
			if op.ImgResizeOp != nil {
				op.Result, op.Err = op.ImgResizeOp(op.In, op.Size)
			} else if op.ImgOp != nil {
				op.Result, op.Err = op.ImgOp(op.In)
			}
		}
		op.Finished = true
		op.FinishedCond.Signal()
	}
}

func writeResult(op *Operation) {
	if op.Err != nil {
		http.Error(op.Resp, fmt.Sprintf("Error transforming image: '%s'", op.Err.Error()), http.StatusInternalServerError)
		return
	}

	addHeaders(op.Resp, op.Result)
	op.Resp.Write(op.Result)
}
