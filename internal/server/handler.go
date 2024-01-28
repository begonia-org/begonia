package server

import (
	"net/http"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func NewHandlers(config *config.Config, vaildator *middleware.APIVildator, grpcServer *grpc.Server, mux *runtime.ServeMux) http.Handler {
	handler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})
	// cors := config.GetCorsConfig()
	// handler = middlerware.AllowCORS(handler, cors)
	return handler
}
