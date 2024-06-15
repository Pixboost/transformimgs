package img

import (
	"context"
	"errors"
	"fmt"
	"github.com/dooman87/glogi"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CacheTTL is the number of seconds  that will be written to max-age HTTP header
var CacheTTL int

// SaveDataEnabled is the flag to enable/disable Save-Data client hint.
// Sometime CDN doesn't support Save-Data in Vary response header in which
// case you would need to set this to false
var SaveDataEnabled = true

// Log is the logger that could be overridden. Should implement interface glogi.Logger.
// By default is using glogi.SimpleLogger.
var Log glogi.Logger = glogi.NewSimpleLogger()

// Loader is responsible for loading an original image for transformation
type Loader interface {
	// Load loads an image from the given source.
	//
	// ctx is the context of the current transaction. Typically, it's a context
	// of an incoming HTTP request, so we make it possible to pass values through middlewares.
	//
	// Returns an image.
	Load(src string, ctx context.Context) (*Image, error)
}

type Quality int

const (
	DEFAULT Quality = 1 + iota
	LOW
	LOWER
)

type ResizeConfig struct {
	// Size is a size of output images in the format WxH.
	Size string
}

// TransformationConfig is a configuration passed to Processor
// that used during transformations.
type TransformationConfig struct {
	// Src is the source image to transform.
	// This field is required for transformations.
	Src *Image
	// SupportedFormats is the list of output formats supported by client.
	// Processor will use one of those formats for result image. If list
	// is empty the format of the source image will be used.
	SupportedFormats []string
	// Quality defines quality of output image
	Quality Quality
	// TrimBorder is a flag whether we need to remove border or not
	TrimBorder bool
	// Config is the configuration for the specific transformation
	Config interface{}
}

// Processor is the interface for transforming/optimising images.
//
// Each function accepts original image and a list of supported
// output format by client. Each format should be a MIME type, e.g.
// image/png, image/webp. The output image will be encoded in one
// of those formats.
type Processor interface {
	// Resize resizes given image preserving aspect ratio.
	// Format of the size argument is width'x'height.
	// Any dimension could be skipped.
	// For example:
	//* 300x200
	//* 300 - only width
	//* x200 - only height
	Resize(input *TransformationConfig) (*Image, error)

	// FitToSize resizes given image cropping it to the given size and does not respect aspect ratio.
	// Format of the size string is width'x'height, e.g. 300x400.
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

type Command struct {
	Transformation Cmd
	Config         *TransformationConfig
	Resp           http.ResponseWriter
	Result         *Image
	FinishedCond   *sync.Cond
	Finished       bool
	Err            error
}

var emptyGif = [...]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x1, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x21, 0xf9, 0x4, 0x1, 0xa, 0x0, 0x1, 0x0, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x0, 0x2, 0x2, 0x4c, 0x1, 0x0, 0x3b}

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

func (r *Service) OptimiseUrl(resp http.ResponseWriter, req *http.Request) {
	r.transformUrl(resp, req, r.Processor.Optimise, nil)
}

func (r *Service) ResizeUrl(resp http.ResponseWriter, req *http.Request) {
	size, _ := getQueryParam(req.URL, "size")
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

	r.transformUrl(resp, req, r.Processor.Resize, &ResizeConfig{Size: size})
}

func (r *Service) FitToSizeUrl(resp http.ResponseWriter, req *http.Request) {
	size, _ := getQueryParam(req.URL, "size")
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

	r.transformUrl(resp, req, r.Processor.FitToSize, &ResizeConfig{Size: size})
}

func (r *Service) AsIs(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getImgUrl(req)
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}

	Log.Printf("Requested image %s as is\n", imgUrl)

	result, err := r.Loader.Load(imgUrl, req.Context())

	if err != nil {
		sendError(resp, err)
		return
	}

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

func (r *Service) execOp(op *Command) {
	op.FinishedCond = sync.NewCond(&sync.Mutex{})

	queue := r.getQueue()
	queue.AddAndWait(op, func() {
		Log.Printf("Image [%s] transformed successfully, writing to the response", op.Config.Src.Id)
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

func getQueryParam(url *url.URL, name string) (string, bool) {
	if len(url.Query()[name]) == 1 {
		return url.Query()[name][0], true
	}
	return "", url.Query().Has(name)
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
	_, _ = op.Resp.Write(op.Result.Data)
}

func (r *Service) transformUrl(resp http.ResponseWriter, req *http.Request, transformation Cmd, config interface{}) {
	time.Sleep(5 * time.Second)
	imgUrl := getImgUrl(req)
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", http.StatusBadRequest)
		return
	}

	var dppx float64 = 0
	dppxParam, _ := getQueryParam(req.URL, "dppx")
	if len(dppxParam) != 0 {
		var err error
		dppx, err = strconv.ParseFloat(dppxParam, 32)
		if err != nil {
			http.Error(resp, "dppx query param must be a number", http.StatusBadRequest)
			return
		}
	}

	var saveDataParam = ""
	if SaveDataEnabled {
		saveDataParam, _ = getQueryParam(req.URL, "save-data")
		if len(saveDataParam) > 0 && saveDataParam != "off" && saveDataParam != "hide" {
			http.Error(resp, "save-data query param must be one of 'off', 'hide'", http.StatusBadRequest)
			return
		}
	}

	var trimBorder = false
	trimBorderParamValue, trimBorderParamExist := getQueryParam(req.URL, "trim-border")
	if trimBorderParamExist {
		if len(trimBorderParamValue) == 0 {
			trimBorder = true
		} else {
			var err error
			trimBorder, err = strconv.ParseBool(trimBorderParamValue)
			if err != nil {
				http.Error(resp, "can't parse trim-border param", http.StatusBadRequest)
			}
		}
	}

	saveDataHeader := req.Header.Get("Save-Data")

	Log.Printf("[%s]: Transforming image %s using config %+v\n", req.URL.String(), imgUrl, config)

	if SaveDataEnabled {
		resp.Header().Add("Vary", "Accept, Save-Data")

		if saveDataHeader == "on" && saveDataParam == "hide" {
			_, _ = resp.Write(emptyGif[:])
			return
		}
	} else {
		resp.Header().Add("Vary", "Accept")
	}

	supportedFormats := getSupportedFormats(req)

	srcImage, err := r.Loader.Load(imgUrl, req.Context())
	if err != nil {
		sendError(resp, err)
	}
	Log.Printf("Source image [%s] loaded successfully, adding to the queue\n", imgUrl)

	r.execOp(&Command{
		Transformation: transformation,
		Config: &TransformationConfig{
			Src:              srcImage,
			SupportedFormats: supportedFormats,
			Quality:          getQuality(saveDataHeader, saveDataParam, dppx),
			TrimBorder:       trimBorder,
			Config:           config,
		},
		Resp: resp,
	})
}

func getQuality(saveDataHeader string, saveDataParam string, dppx float64) Quality {
	if dppx >= 2.0 {
		return LOWER
	}

	if SaveDataEnabled && saveDataHeader == "on" && saveDataParam != "off" {
		return LOW
	}

	return DEFAULT
}

func sendError(resp http.ResponseWriter, err error) {
	if err != nil {
		var httpErr *HttpError
		if errors.As(err, &httpErr) {
			http.Error(resp, httpErr.Error(), httpErr.Code())
		} else {
			http.Error(resp, fmt.Sprintf("Error reading image: '%s'", err.Error()), http.StatusInternalServerError)
		}
	}
}
