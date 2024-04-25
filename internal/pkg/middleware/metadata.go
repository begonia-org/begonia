package middleware

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func HTTPAnnotator(ctx context.Context, r *http.Request) metadata.MD{
	md := metadata.Pairs("x-request-id", r.Header.Get("x-request-id"))
	return md
}
