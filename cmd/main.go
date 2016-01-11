package main
import (
	"net/http"
	"github.com/dooman87/transformimgs/service"
	"log"
)

func main() {
	http.HandleFunc("/health", service.Health)

	log.Fatal(http.ListenAndServe(":8080", nil))
}