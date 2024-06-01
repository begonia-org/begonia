package auth

import (
	"context"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Auth struct {
	ak       *AccessKeyAuthMiddleware
	jwt      *JWTAuth
	apikey   ApiKeyAuth
	priority int
	name     string
}

func NewAuth(ak *AccessKeyAuthMiddleware, jwt *JWTAuth, apikey ApiKeyAuth) gosdk.LocalPlugin {
	return &Auth{
		ak:     ak,
		jwt:    jwt,
		apikey: apikey,
		name:   "auth",
	}
}

func (a *Auth) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if !IfNeedValidate(ctx, info.FullMethod) {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")
	}
	xApiKey := md.Get("x-api-key")
	if len(xApiKey) != 0 {
		return a.apikey.UnaryInterceptor(ctx, req, info, handler)
	}
	authorization := a.jwt.GetAuthorizationFromMetadata(md)

	if authorization == "" {
		return nil, gosdk.NewError(pkg.ErrTokenMissing, int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}

	if strings.Contains(authorization, "Bearer") {
		return a.jwt.UnaryInterceptor(ctx, req, info, handler)
	}
	return a.ak.UnaryInterceptor(ctx, req, info, handler)

}

func (a *Auth) StreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata not exists in context")
	}
	xApiKey := md.Get("x-api-key")
	if len(xApiKey) != 0 {
		return a.apikey.StreamInterceptor(srv, ss, info, handler)
	}
	authorization := a.jwt.GetAuthorizationFromMetadata(md)

	if authorization == "" {
		return gosdk.NewError(pkg.ErrTokenMissing, int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	if strings.Contains(authorization, "Bearer") {
		return a.jwt.StreamInterceptor(srv, ss, info, handler)

	}
	return a.ak.StreamInterceptor(srv, ss, info, handler)
}

func (a *Auth) SetPriority(priority int) {
	a.priority = priority
}
func (a *Auth) Priority() int {
	return a.priority
}

func (a *Auth) Name() string {
	return a.name
}
