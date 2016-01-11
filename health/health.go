//Provides health check
package health

import (
	"fmt"
	"net/http"
)

//Returns OK string.
//Shows only if service accessible or not.
func Health(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "OK")
}
