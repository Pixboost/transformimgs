package service_test
import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/dooman87/transformimgs/service"
)

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	service.Health(w, req)

	expected := "OK"
	if w.Body.String() != expected {
		t.Fatalf("Expected %s but got %s", expected, w.Body.String())
	}
}
