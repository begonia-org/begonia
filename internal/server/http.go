package server

import (
	"net/http"
)

func NewHttpServer(endpoint string, handle http.Handler) *http.Server {
	// cors := config.GetStringSlice(fmt.Sprintf("%s.gateway.cors", env))
	s := &http.Server{
		Addr:    endpoint,
		Handler: handle,
	}
	return s
}
