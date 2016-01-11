package health_test
import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/dooman87/transformimgs/health"
)

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	health.Health(w, req)

	expected := "OK"
	if w.Body.String() != expected {
		t.Fatalf("Expected %s but got %s", expected, w.Body.String())
	}
}
