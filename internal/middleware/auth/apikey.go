package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ApiKeyAuth interface {
	gosdk.LocalPlugin
}

type ApiKeyAuthImpl struct {
	config   *config.Config
	priority int
	name     string
	authz    *biz.AuthzUsecase
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
	apikey := ""
	if apikey, err = a.check(ctx); err == nil && apikey != "" {

		identity, err := a.authz.GetIdentity(ctx, gosdk.ApiKeyType, apikey)
		if err != nil {
			return nil, gosdk.NewError(fmt.Errorf("query uid base on apikey get error:%w", err), int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")
		}
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(gosdk.HeaderXIdentity, identity))
		return handler(ctx, req)
	}
	return nil, err
}

func NewApiKeyAuth(config *config.Config, authz *biz.AuthzUsecase) ApiKeyAuth {
	return &ApiKeyAuthImpl{
		config: config,
		authz:  authz,
		name:   "api_key_auth",
	}
}
func (a *ApiKeyAuthImpl) check(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", gosdk.NewError(status.Errorf(codes.Unauthenticated, "metadata not exists in context"), int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	// authorization := a.GetAuthorizationFromMetadata(md)
	apikeys := md.Get(gosdk.HeaderXApiKey)
	if len(apikeys) == 0 {
		return "", gosdk.NewError(status.Errorf(codes.Unauthenticated, "apikey not exists in context"), int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	apikey := apikeys[0]
	if apikey != a.config.GetAdminAPIKey() {
		return "", gosdk.NewError(pkg.ErrAPIKeyNotMatch, int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")

	}
	return apikey, nil
}
func (a *ApiKeyAuthImpl) ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	apikey := ""
	var err error
	if apikey, err = a.check(ctx); err == nil && apikey != "" {
		identity, err := a.authz.GetIdentity(ctx, gosdk.ApiKeyType, apikey)
		if err != nil {
			return ctx, gosdk.NewError(fmt.Errorf("query user id base on apikey err:%w", err), int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")
		}
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(gosdk.HeaderXIdentity, identity))
		headers.Set(strings.ToLower(gosdk.HeaderXIdentity), identity)
		return ctx, err
	}
	return ctx, err
}

func (a *ApiKeyAuthImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	grpcStream := NewGrpcStream(ss, info.FullMethod, ss.Context(), a)
	defer grpcStream.Release()
	return handler(srv, grpcStream)
}
