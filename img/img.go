package img

import (
	"fmt"
	"net/http"
	"net/url"
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
}

type Service struct {
	Reader    ImgReader
	Processor ImgProcessor
}

//Transforms image that passed in url param and
//returns the result.
//Query params:
// * url - url of the original image. Required.
// * size - new size of the image. Should be in the width'x'height format.
//   Accepts only width, e.g. 300 or height e.g. x200
//
//Examples:
// */transform?url=www.site.com/img.png&size=300x200
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

	resp.Write(result)
}

func getQueryParam(url *url.URL, name string) string {
	if len(url.Query()[name]) == 1 {
		return url.Query()[name][0]
	}
	return ""
}
