package server

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/internal/pkg/middlerware"
)

func NewHttpServer(endpoint string, cors []string, vaildator *middlerware.APIVildator, mux *runtime.ServeMux) *http.Server {
	// cors := config.GetStringSlice(fmt.Sprintf("%s.gateway.cors", env))
	s := &http.Server{
		Addr:    endpoint,
		Handler: middlerware.AllowCORS(vaildator.Vildator(mux), cors),
	}
	return s
}
