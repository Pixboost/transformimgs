package transformimgs_test

import (
	"fmt"
	"github.com/Pixboost/transformimgs/v8/img"
	"github.com/Pixboost/transformimgs/v8/img/loader"
	"github.com/Pixboost/transformimgs/v8/img/processor"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
)

func Example() {
	l := &loader.Http{}
	p, err := processor.NewImageMagick("/usr/local/bin/convert", "/usr/local/bin/identify")
	if err != nil {
		log.Fatal(err)
	}

	s, err := img.NewService(l, p, runtime.NumCPU())
	if err != nil {
		log.Fatal(err)
	}

	// s.GetRouter() is ready to use in your web server
	server := httptest.NewServer(s.GetRouter())
	defer server.Close()

	resizeApi := fmt.Sprintf("%s/img/https://raw.githubusercontent.com/Pixboost/transformimgs/main/quickstart/site/img/parrot-lossy.jpg/resize?size=600", server.URL)
	resp, err := http.Get(resizeApi)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)

	// Output:
	// 200 OK
}
