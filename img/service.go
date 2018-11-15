package img

import (
	"fmt"
	"github.com/dooman87/glogi"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//Number of seconds that will be written to max-age HTTP header
var CacheTTL int

//Log writer that can be overrided. Should implement interface glogi.Logger.
// By default is using glogi.SimpleLogger.
var Log glogi.Logger = glogi.NewSimpleLogger()

//Reads image from a given source
type ImgReader interface {
	//Reads image from the url.
	//Returns byte array of the image and content type or error.
	Read(url string) ([]byte, string, error)
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
	Resize(data []byte, size string, imageId string, supportedFormats []string) ([]byte, error)

	//Resize given image fitting it to a given size.
	//Form of the the size string is width'x'height.
	//For example, 300x400
	FitToSize(data []byte, size string, imageId string, supportedFormats []string) ([]byte, error)

	//Optimises given image to reduce size.
	Optimise(data []byte, imageId string, supportedFormats []string) ([]byte, error)
}

type Service struct {
	Reader      ImgReader
	Processor   ImgProcessor
	Q           []*Queue
	currProc    int
	currProcMux sync.Mutex
}

type OptimiseCmd func([]byte, string, []string) ([]byte, error)
type ResizeCmd func([]byte, string, string, []string) ([]byte, error)

type Command struct {
	Optimise         OptimiseCmd
	Resize           ResizeCmd
	Image            []byte
	ImgId            string
	Size             string
	Resp             http.ResponseWriter
	SupportedFormats []string
	Result           []byte
	FinishedCond     *sync.Cond
	Finished         bool
	Err              error
}

func NewService(r ImgReader, p ImgProcessor, procNum int) (*Service, error) {
	if procNum <= 0 {
		return nil, fmt.Errorf("procNum must be positive, but got [%d]", procNum)
	}

	Log.Printf("Creating new service with [%d] number of processors\n", procNum)

	srv := &Service{
		Reader:    r,
		Processor: p,
		Q:         make([]*Queue, procNum),
	}

	for i := 0; i < procNum; i++ {
		srv.Q[i] = NewQueue()
	}
	srv.currProc = 0

	return srv, nil
}

func (r *Service) GetRouter() *mux.Router {
	router := mux.NewRouter().SkipClean(true)
	router.HandleFunc("/img/{imgUrl:.*}/resize", http.HandlerFunc(r.ResizeUrl))
	router.HandleFunc("/img/{imgUrl:.*}/fit", http.HandlerFunc(r.FitToSizeUrl))
	router.HandleFunc("/img/{imgUrl:.*}/asis", http.HandlerFunc(r.AsIs))
	router.HandleFunc("/img/{imgUrl:.*}/optimise", http.HandlerFunc(r.OptimiseUrl))

	return router
}

// swagger:operation GET /img/{imgUrl}/optimise optimiseImage
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
// - name: imgUrl
//   required: true
//   in: path
//   type: string
//   description: url of the original image
// responses:
//   '200':
//     description: Optimised image in the same format as original.
func (r *Service) OptimiseUrl(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getImgUrl(req)
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}
	supportedFormats := getSupportedFormats(req)

	Log.Printf("Optimising image %s\n", imgUrl)

	input, _, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}
	resp.Header().Add("Vary", "Accept")

	r.execOp(&Command{
		Optimise:         r.Processor.Optimise,
		ImgId:            imgUrl,
		Image:            input,
		Resp:             resp,
		SupportedFormats: supportedFormats,
	})
}

// swagger:operation GET /img/{imgUrl}/resize resizeImage
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
// - name: imgUrl
//   required: true
//   in: path
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
	imgUrl := getImgUrl(req)
	size := getQueryParam(req.URL, "size")
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}
	if len(size) == 0 {
		http.Error(resp, "size param is required", http.StatusBadRequest)
		return
	}
	supportedFormats := getSupportedFormats(req)

	Log.Printf("Resizing image %s to %s\n", imgUrl, size)

	input, _, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}
	resp.Header().Add("Vary", "Accept")

	r.execOp(&Command{
		Resize:           r.Processor.Resize,
		Image:            input,
		ImgId:            imgUrl,
		Size:             size,
		Resp:             resp,
		SupportedFormats: supportedFormats,
	})
}

// swagger:operation GET /img/{imgUrl}/fit fitImage
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
// - name: imgUrl
//   required: true
//   in: path
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
	imgUrl := getImgUrl(req)
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
			Log.Printf("Error while matching size: %s\n", err.Error())
		}
		http.Error(resp, "size param should be in format WxH", http.StatusBadRequest)
		return
	}
	supportedFormats := getSupportedFormats(req)

	Log.Printf("Fit image %s to size %s\n", imgUrl, size)

	input, _, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}
	resp.Header().Add("Vary", "Accept")

	r.execOp(&Command{
		Resize:           r.Processor.FitToSize,
		Image:            input,
		ImgId:            imgUrl,
		Size:             size,
		Resp:             resp,
		SupportedFormats: supportedFormats,
	})
}

// swagger:operation GET /img/{imgUrl}/asis asisImage
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
// - name: imgUrl
//   required: true
//   in: path
//   type: string
//   description: url of the image
//
// responses:
//   '200':
//     description: Requested image.
func (r *Service) AsIs(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getImgUrl(req)
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}

	Log.Printf("Requested image %s as is\n", imgUrl)

	result, contentType, err := r.Reader.Read(imgUrl)

	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	} else {
		if len(contentType) > 0 {
			resp.Header().Add("Content-Type", contentType)
		}

		r.execOp(&Command{
			Result: result,
			ImgId:  imgUrl,
			Resp:   resp,
		})
	}
}

func (r *Service) execOp(op *Command) {
	op.FinishedCond = sync.NewCond(&sync.Mutex{})

	queue := r.getQueue()
	queue.AddAndWait(op, func() {
		writeResult(op)
	})
}

func (r *Service) getQueue() *Queue {
	//Get the next execution channel
	r.currProcMux.Lock()
	r.currProc++
	if r.currProc == len(r.Q) {
		r.currProc = 0
	}
	procIdx := r.currProc
	r.currProcMux.Unlock()

	return r.Q[procIdx]
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

func getImgUrl(req *http.Request) string {
	imgUrl := mux.Vars(req)["imgUrl"]
	if len(imgUrl) == 0 {
		return ""
	}

	if strings.HasPrefix(imgUrl, "//") && len(req.Header["X-Forwarded-Proto"]) == 1 {
		imgUrl = fmt.Sprintf("%s:%s", req.Header["X-Forwarded-Proto"][0], imgUrl)
	}

	return imgUrl
}

func getSupportedFormats(req *http.Request) []string {
	acceptHeader := req.Header["Accept"]
	if len(acceptHeader) > 0 {
		accepts := strings.Split(acceptHeader[0], ",")
		trimmedAccepts := make([]string, len(accepts))
		for i, a := range accepts {
			trimmedAccepts[i] = strings.TrimSpace(a)
		}
		return trimmedAccepts
	}

	return []string{}
}

func writeResult(op *Command) {
	if op.Err != nil {
		http.Error(op.Resp, fmt.Sprintf("Error transforming image: '%s'", op.Err.Error()), http.StatusInternalServerError)
		return
	}

	addHeaders(op.Resp, op.Result)
	op.Resp.Write(op.Result)
}
