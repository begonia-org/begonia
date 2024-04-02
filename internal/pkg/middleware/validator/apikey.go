package validator

import (
	"context"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ApiKeyAuth interface{
	gosdk.LocalPlugin
}

type ApiKeyAuthImpl struct {
	config   *config.Config
	priority int
	name     string
}

func (a *ApiKeyAuthImpl) SetPriority(priority int) {
	a.priority = priority
}
func (a *ApiKeyAuthImpl) Priority() int {
	return a.priority
}
func (a *ApiKeyAuthImpl) Name() string {
	return a.name
}
func (a *ApiKeyAuthImpl) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if !IfNeedValidate(ctx, info.FullMethod) {
		return handler(ctx, req)

	}
	if err := a.check(ctx); err != nil {
		return nil, err

	}
	return handler(ctx, req)
}

func NewApiKeyAuth(config *config.Config) ApiKeyAuth {
	return &ApiKeyAuthImpl{
		config: config,
		name:   "api_key_auth",
	}
}
func (a *ApiKeyAuthImpl) check(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New(status.Errorf(codes.Unauthenticated, "metadata not exists in context"), int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	// authorization := a.GetAuthorizationFromMetadata(md)
	apikeys := md.Get("x-api-key")
	if len(apikeys) == 0 {
		return errors.New(status.Errorf(codes.Unauthenticated, "apikey not exists in context"), int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	apikey := apikeys[0]
	if apikey != a.config.GetAdminAPIKey() {
		return errors.New(errors.ErrAPIKeyNotMatch, int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")

	}
	return nil
}
func (a *ApiKeyAuthImpl) ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	if err := a.check(ctx); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (a *ApiKeyAuthImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	grpcStream := NewGrpcStream(ss, info.FullMethod, ss.Context(), a)
	defer grpcStream.Release()
	return handler(srv, grpcStream)
}
