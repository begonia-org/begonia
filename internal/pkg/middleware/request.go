package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	// headers := []string{"Content-Type", "Accept", "Authorization", "X-Token", "x-date", "x-access-key"}
	w.Header().Set("Access-Control-Allow-Headers", "*")
	// methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", "*")
}

type CorsMiddleware struct {
	Cors []string
}

func (cors *CorsMiddleware) Handle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clientOrigin := r.Header.Get("Origin"); clientOrigin != "" {
			var isAllowed bool

			for _, origin := range cors.Cors {
				if origin == "*" || strings.HasSuffix(clientOrigin, origin) {
					isAllowed = true
					break
				}
			}
			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", clientOrigin)
				if r.Method == "OPTIONS" {
					preflightHandler(w, r)
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}
func RequestIDMiddleware(ctx context.Context, r *http.Request) metadata.MD {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}
	if val := md.HeaderMD.Get("x-request-id"); len(val) > 0 {
		r.Header.Set("x-request-id", val[0])
	}
	return md.HeaderMD

}

func IncomingHeadersToMetadata(ctx context.Context, req *http.Request) metadata.MD {
	// 创建一个新的 metadata.MD 实例
	md := metadata.MD{}
	for k, v := range req.Header {
		if strings.ToLower(k) == "authorization" || strings.HasPrefix(strings.ToLower(k), "x-") {
			md.Set(strings.ToLower(k), v...)
		}
	}
	md.Set("x-request-id", uuid.New().String())
	md.Set("uri", req.RequestURI)
	md.Set("method", req.Method)
	md.Set("remote_addr", req.RemoteAddr)
	md.Set("protocol", req.Proto)
	md.Set("response-type", "application/json")
	_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", md.Get("x-request-id")[0]))
	if val := md.Get("x-uid"); len(val) > 0 {
		_ = grpc.SetHeader(ctx, metadata.Pairs("x-uid", val[0]))
	}
	uri := req.RequestURI
	method := req.Method
	remoteAddr := req.RemoteAddr
	_ = grpc.SetHeader(ctx, metadata.Pairs("uri", uri))
	_ = grpc.SetHeader(ctx, metadata.Pairs("method", method))
	_ = grpc.SetHeader(ctx, metadata.Pairs("remote_addr", remoteAddr))
	_ = grpc.SendHeader(ctx, md)
	return md
}
