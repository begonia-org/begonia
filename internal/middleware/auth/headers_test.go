package auth_test

import (
	"net/http"
	"testing"

	"github.com/begonia-org/begonia/internal/middleware/auth"
	c "github.com/smartystreets/goconvey/convey"
)

type responseWriter struct {
}

func (r *responseWriter) Header() http.Header {
	return make(http.Header)
}
func (r *responseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (r *responseWriter) WriteHeader(int) {

}
func TestHeaders(t *testing.T) {
	c.Convey("TestHeaders", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost", nil)
		h := auth.NewHttpHeader(&responseWriter{}, req)
		c.So(h, c.ShouldNotBeNil)
		h.Set("key", "value")
		h.SendHeader("key", "value")
		h.Release()
	})
}
