package img
import (
	"net/http"
	"strconv"
	"net/url"
)

type ImgReader interface {
	Read(url string) ([]byte, error)
}

type ImgProcessor interface {
	Resize(data []byte, width int, height int) ([]byte, error)
}

type Service struct {
	Reader ImgReader
	Processor ImgProcessor
}

//Transforms image that passed in url param and
//returns the result.
//Query params:
// * url - url of the original image
// * width - width to transform to
// * height - height to transform to
//
//Examples:
// */transform?url=www.site.com/img.png&width=200&height=300
func (r *Service) ResizeUrl(resp http.ResponseWriter, req *http.Request) {
	imgUrl := getQueryParam(req.URL, "url")
	if len(imgUrl) == 0 {
		http.Error(resp, "url param is required", 400)
		return
	}

	width, err := strconv.Atoi(req.URL.Query()["width"][0])
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}

	height, err := strconv.Atoi(req.URL.Query()["height"][0])
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}

	input, err := r.Reader.Read(imgUrl)
	if err != nil {
		http.Error(resp, err.Error(), 500)
		return
	}

	result, err := r.Processor.Resize(input, width, height)

	if err != nil {
		http.Error(resp, err.Error(), 500)
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

//func getQueryParamInt(url *url.URL, name string) (int, err) {
//	if len(url.Query()[name]) == 1 {
//		return url.Query()[name][0]
//	}
//	return ""
//}
