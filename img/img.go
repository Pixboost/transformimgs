package img

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"net/url"
	"strconv"
	"github.com/gorilla/mux"
	"regexp"
)

//Reads image from a given source
type ImgReader interface {
	//Reads image from the url.
	//Returns byte array of the image or
	//error.
	Read(url string) ([]byte, error)
}

//Processes image applying different transformation.
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
	FitToSize(data []byte, size string) ([]byte, error)
}

type Service struct {
	Reader    ImgReader
	Processor ImgProcessor
}

func (r *Service) GetRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/img/resize", http.HandlerFunc(r.ResizeUrl))
	router.HandleFunc("/img/fit", http.HandlerFunc(r.FitToSizeUrl))

	return router
}

//Transforms image that passed in url param and
//returns the result.
//Query params:
// * url - url of the original image. Required.
// * size - new size of the image. Should be in the width'x'height format.
//   Accepts only width, e.g. 300 or height e.g. x200
//
//Examples:
// */resize?url=www.site.com/img.png&size=300x200
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

	glog.Infof("Resizing image %s to %s", imgUrl, size)

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

	resp.Header().Add("Content-Length", strconv.Itoa(len(result)))
	resp.Header().Add("Cache-Control", "public, max-age=86400")
	resp.Write(result)
}

//Transforms image that passed in url param and
//returns the result.
//Query params:
// * url - url of the original image. Required.
// * size - new size of the image. Should be in the width'x'height format.
//
//Examples:
// */fit?url=www.site.com/img.png&size=300x200
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
			glog.Errorf("Error while matching size: %s", err.Error())
		}
		http.Error(resp, "size param should be in format WxH", http.StatusBadRequest)
		return
	}

	glog.Infof("Fit image %s to size %s", imgUrl, size)

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

	resp.Header().Add("Content-Length", strconv.Itoa(len(result)))
	resp.Header().Add("Cache-Control", "public, max-age=86400")
	resp.Write(result)
}

func getQueryParam(url *url.URL, name string) string {
	if len(url.Query()[name]) == 1 {
		return url.Query()[name][0]
	}
	return ""
}
