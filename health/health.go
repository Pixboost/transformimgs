//Provides health check
package health
import (
	"net/http"
	"fmt"
)

//Returns OK string.
//Shows only if service accessible or not.
func Health(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "OK")
}

