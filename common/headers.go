package common

import (
	"context"
	"fmt"

	api "github.com/wetrycode/begonia/common/api/v1"
	"google.golang.org/grpc/metadata"
)

type RequestHeader interface {
	GetHeader(key string) string
}

type CommonHeaders struct {
	*api.Headers
}

func getHeader(md metadata.MD, key string) string {
	if v, ok := md[key]; ok {
		return v[0]
	}
	return ""
}
func NewHeadersFromContext(ctx context.Context) (RequestHeader, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("not found metadata")
	}
	return &CommonHeaders{

		Headers: &api.Headers{
			Uid:            getHeader(md, "x-uid"),
			Filename:       getHeader(md, "x-filename"),
			Authentication: getHeader(md, "authentication"),
		},
	}, nil
}

func (c *CommonHeaders) GetHeader(key string) string {
	switch key {
	case "x-uid":
		return c.Uid
	case "x-filename":
		return c.Filename
	case "authentication":
		return c.Authentication
	default:
		return ""
	}
}