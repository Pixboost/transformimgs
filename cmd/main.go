package main
import (
	"net/http"
	"github.com/dooman87/transformimgs/health"
	"log"
)

func main() {
	http.HandleFunc("/health", health.Health)

	log.Fatal(http.ListenAndServe(":8080", nil))
}