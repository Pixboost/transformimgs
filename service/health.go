//Provides health check
package service
import (
	"net/http"
	"fmt"
)

func Health(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "OK")
}

