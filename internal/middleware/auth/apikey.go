package auth

import (
	"context"
	"fmt"

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
		md, _ := metadata.FromIncomingContext(ctx)

		md = metadata.Join(md, metadata.Pairs(gosdk.HeaderXIdentity, identity))
		ctx = metadata.NewIncomingContext(ctx, md)
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
	out, outOK := metadata.FromOutgoingContext(ctx)
	if !ok && !outOK {
		return "", gosdk.NewError(status.Errorf(codes.Unauthenticated, "metadata not exists in context"), int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	// authorization := a.GetAuthorizationFromMetadata(md)
	apikeys := md.Get(gosdk.HeaderXApiKey)
	outAPIKeys := out.Get(gosdk.HeaderXApiKey)
	if len(apikeys) == 0 && len(outAPIKeys) == 0 {
		return "", gosdk.NewError(status.Errorf(codes.Unauthenticated, "apikey not exists in context"), int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}

	apikey := ""
	if len(apikeys) != 0 {
		apikey = apikeys[0]
	} else if len(outAPIKeys) != 0 {
		apikey = outAPIKeys[0]
	}

	if apikey != a.config.GetAdminAPIKey() {
		return "", gosdk.NewError(pkg.ErrAPIKeyNotMatch, int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")

	}
	return apikey, nil
}
func (a *ApiKeyAuthImpl) ValidateStream(ctx context.Context, req interface{}, fullName string) (context.Context, error) {
	apikey := ""
	var err error
	if apikey, err = a.check(ctx); err == nil && apikey != "" {
		identity, err := a.authz.GetIdentity(ctx, gosdk.ApiKeyType, apikey)
		if err != nil {
			return ctx, gosdk.NewError(fmt.Errorf("query user id base on apikey err:%w", err), int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")
		}
		md, _ := metadata.FromIncomingContext(ctx)

		md = metadata.Join(md, metadata.Pairs(gosdk.HeaderXIdentity, identity))
		// headers.Set(strings.ToLower(gosdk.HeaderXIdentity), identity)
		return metadata.NewIncomingContext(ctx, md), err
	}
	return ctx, err
}

func (a *ApiKeyAuthImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	ctx := ss.Context()
	if apikey, err := a.check(ctx); err == nil && apikey != "" {
		identity, err := a.authz.GetIdentity(ctx, gosdk.ApiKeyType, apikey)
		if err != nil {
			return gosdk.NewError(fmt.Errorf("query user id base on apikey err:%w", err), int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")
		}
		md, _ := metadata.FromIncomingContext(ctx)
		md.Set(gosdk.HeaderXIdentity, identity)
		ctx = metadata.NewIncomingContext(ctx, md)
		ctx = metadata.AppendToOutgoingContext(ctx, gosdk.HeaderXIdentity, identity)
	} else {
		return gosdk.NewError(pkg.ErrAPIKeyNotMatch, int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")

	}
	grpcStream := NewGrpcStream(ss, info.FullMethod, ctx, a)
	defer grpcStream.Release()
	err := handler(srv, grpcStream)

	return err
}
func (a *ApiKeyAuthImpl) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if !IfNeedValidate(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}
	if apikey, err := a.check(ctx); err == nil && apikey != "" {
		identity, err := a.authz.GetIdentity(ctx, gosdk.ApiKeyType, apikey)
		if err != nil {
			return nil, gosdk.NewError(fmt.Errorf("query user id base on apikey err:%w", err), int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")
		}
		md, _ := metadata.FromOutgoingContext(ctx)
		md = metadata.Join(md, metadata.Pairs(gosdk.HeaderXIdentity, identity))
		ctx = metadata.NewOutgoingContext(ctx, md)
		in, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			in = metadata.New(make(map[string]string))
		}
		in.Set(gosdk.HeaderXIdentity, identity)
		// log.Printf("incoming identity:%s", identity)
		ctx = metadata.NewIncomingContext(ctx, in)
		return streamer(ctx, desc, cc, method, opts...)
	}
	return nil, gosdk.NewError(pkg.ErrAPIKeyNotMatch, int32(api.UserSvrCode_USER_APIKEY_NOT_MATCH_ERR), codes.Unauthenticated, "authorization_check")
}
