package img

import (
	"context"
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

// Number of seconds that will be written to max-age HTTP header
var CacheTTL int

// Log writer that can be overrided. Should implement interface glogi.Logger.
// By default is using glogi.SimpleLogger.
var Log glogi.Logger = glogi.NewSimpleLogger()

// Loaders is responsible for loading an original image for transformation
type Loader interface {
	// Load loads an image from the given source.
	//
	// ctx is a context of the current transaction. Typically it's a context
	// of an incoming HTTP request, so we make it possible to pass values through middlewares.
	//
	// Returns an image.
	Load(src string, ctx context.Context) (*Image, error)
}

type Quality int

const (
	DEFAULT Quality = 1 + iota
	LOW
)

type ResizeConfig struct {
	// Size is a size of output images in the format WxH.
	Size string
}

// TransformationConfig is a configuration passed to Processor
// that used during transformations.
type TransformationConfig struct {
	// Src is a source image to transform.
	// This field is required for transformations.
	Src *Image
	// SupportedFormats is a list of output formats supported by client.
	// Processor will use one of those formats for result image. If list
	// is empty the format of the source image will be used.
	SupportedFormats []string
	// Quality defines quality of output image
	Quality Quality
	// Config is a configuration for the specific transformation
	Config interface{}
}

// Processor is an interface for transforming/optimising images.
//
// Each function accepts original image and a list of supported
// output format by client. Each format should be a MIME type, e.g.
// image/png, image/webp. The output image will be encoded in one
// of those formats.
type Processor interface {
	// Resize resizes given image preserving aspect ratio.
	// Format of the the size argument is width'x'height.
	// Any dimension could be skipped.
	// For example:
	//* 300x200
	//* 300 - only width
	//* x200 - only height
	Resize(input *TransformationConfig) (*Image, error)

	// FitToSize resizes given image cropping it to the given size and does not respect aspect ratio.
	// Format of the the size string is width'x'height, e.g. 300x400.
	FitToSize(input *TransformationConfig) (*Image, error)

	// Optimise optimises given image to reduce size of the served image.
	Optimise(input *TransformationConfig) (*Image, error)
}

type Service struct {
	Loader      Loader
	Processor   Processor
	Q           []*Queue
	currProc    int
	currProcMux sync.Mutex
}

type Cmd func(input *TransformationConfig) (*Image, error)

//type OptimiseCmd func(input *TransformationConfig) (*Image, error)
//type ResizeCmd func(input *TransformationConfig,size string) (*Image, error)

type Command struct {
	Transformation Cmd
	Config         *TransformationConfig
	Resp           http.ResponseWriter
	Result         *Image
	FinishedCond   *sync.Cond
	Finished       bool
	Err            error
}

func NewService(r Loader, p Processor, procNum int) (*Service, error) {
	if procNum <= 0 {
		return nil, fmt.Errorf("procNum must be positive, but got [%d]", procNum)
	}

	Log.Printf("Creating new service with [%d] number of processors\n", procNum)

	srv := &Service{
		Loader:    r,
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
	router.HandleFunc("/img/{imgUrl:.*}/resize", r.ResizeUrl)
	router.HandleFunc("/img/{imgUrl:.*}/fit", r.FitToSizeUrl)
	router.HandleFunc("/img/{imgUrl:.*}/asis", r.AsIs)
	router.HandleFunc("/img/{imgUrl:.*}/optimise", r.OptimiseUrl)

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

	srcImage, err := r.Loader.Load(imgUrl, req.Context())
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}
	resp.Header().Add("Vary", "Accept")

	r.execOp(&Command{
		Transformation: r.Processor.Optimise,
		Config: &TransformationConfig{
			Src:              srcImage,
			SupportedFormats: supportedFormats,
		},
		Resp: resp,
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
	if match, err := regexp.MatchString(`^\d*[x]?\d*$`, size); !match || err != nil {
		if err != nil {
			Log.Printf("Error while matching size: %s\n", err.Error())
		}
		http.Error(resp, "size param should be in format WxH", http.StatusBadRequest)
		return
	}

	supportedFormats := getSupportedFormats(req)

	Log.Printf("Resizing image %s to %s\n", imgUrl, size)

	srcImage, err := r.Loader.Load(imgUrl, req.Context())
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}
	resp.Header().Add("Vary", "Accept")

	r.execOp(&Command{
		Transformation: r.Processor.Resize,
		Config: &TransformationConfig{
			Src:              srcImage,
			SupportedFormats: supportedFormats,
			Config:           &ResizeConfig{Size: size},
		},
		Resp: resp,
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

	srcImage, err := r.Loader.Load(imgUrl, req.Context())
	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}
	resp.Header().Add("Vary", "Accept")

	r.execOp(&Command{
		Transformation: r.Processor.FitToSize,
		Config: &TransformationConfig{
			Src:              srcImage,
			SupportedFormats: supportedFormats,
			Config:           &ResizeConfig{Size: size},
		},
		Resp: resp,
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

	result, err := r.Loader.Load(imgUrl, req.Context())

	if err != nil {
		http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	} else {
		if len(result.MimeType) > 0 {
			resp.Header().Add("Content-Type", result.MimeType)
		}

		r.execOp(&Command{
			Config: &TransformationConfig{
				Src: &Image{
					Id: imgUrl,
				},
			},
			Result: result,
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
	// Get the next execution channel
	r.currProcMux.Lock()
	r.currProc++
	if r.currProc == len(r.Q) {
		r.currProc = 0
	}
	procIdx := r.currProc
	r.currProcMux.Unlock()

	return r.Q[procIdx]
}

// Adds Content-Length and Cache-Control headers
func addHeaders(resp http.ResponseWriter, image *Image) {
	if len(image.MimeType) != 0 {
		resp.Header().Add("Content-Type", image.MimeType)
	}
	resp.Header().Add("Content-Length", strconv.Itoa(len(image.Data)))
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
	op.Resp.Write(op.Result.Data)
}
