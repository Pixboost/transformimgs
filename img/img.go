package img

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

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
	Reader    ImgReader
	Processor ImgProcessor
	cache     int
}

func NewService(r ImgReader, p ImgProcessor, cacheSec int) (*Service, error) {
	return &Service{
		Reader:    r,
		Processor: p,
		cache:     cacheSec,
	}, nil
}

func (r *Service) GetRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/img/resize", http.HandlerFunc(r.ResizeUrl))
	router.HandleFunc("/img/fit", http.HandlerFunc(r.FitToSizeUrl))
	router.HandleFunc("/img", http.HandlerFunc(r.OptimiseUrl))

	return router
}

// swagger:operation GET /img optimiseImage
//
// Optimises image that passed in url query param and returns the result.
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
//     description: Optimised image
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

	result, err := r.Processor.Optimise(input)

	if err != nil {
		http.Error(resp, fmt.Sprintf("Error transforming image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}

	r.addHeaders(resp, result)
	resp.Write(result)
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
//     description: Resized image
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

	result, err := r.Processor.Resize(input, size)

	if err != nil {
		http.Error(resp, fmt.Sprintf("Error transforming image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}

	r.addHeaders(resp, result)
	resp.Write(result)
}

// swagger:operation GET /img/fit fitImage
//
// Resizes image to the exact size and optimizes it.
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
//     description: Resized image
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

	result, err := r.Processor.FitToSize(input, size)

	if err != nil {
		http.Error(resp, fmt.Sprintf("Error transforming image: '%s'", err.Error()), http.StatusInternalServerError)
		return
	}

	r.addHeaders(resp, result)
	resp.Write(result)
}

//Adds Content-Length and Cache-Control headers
func (r *Service) addHeaders(resp http.ResponseWriter, body []byte) {
	resp.Header().Add("Content-Length", strconv.Itoa(len(body)))
	resp.Header().Add("Cache-Control", fmt.Sprintf("public, max-age=%d", r.cache))
}

func getQueryParam(url *url.URL, name string) string {
	if len(url.Query()[name]) == 1 {
		return url.Query()[name][0]
	}
	return ""
}
