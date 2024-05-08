package middleware

import (
	"context"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/middleware/auth"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Auth struct {
	ak       *auth.AccessKeyAuthMiddleware
	jwt      *auth.JWTAuth
	apikey   auth.ApiKeyAuth
	priority int
	name     string
}

func NewAuth(ak *auth.AccessKeyAuthMiddleware, jwt *auth.JWTAuth, apikey auth.ApiKeyAuth) gosdk.LocalPlugin {
	return &Auth{
		ak:     ak,
		jwt:    jwt,
		apikey: apikey,
		name:   "auth",
	}
}

func (a *Auth) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if !auth.IfNeedValidate(ctx, info.FullMethod) {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")
	}
	xApiKey:=md.Get("x-api-key")
	if len(xApiKey) != 0 {
		return a.apikey.UnaryInterceptor(ctx, req, info, handler)
	}
	authorization := a.jwt.GetAuthorizationFromMetadata(md)

	if authorization == "" {
		return nil, errors.New(errors.ErrTokenMissing, int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}

	if strings.Contains(authorization, "Bearer") {
		ctx, err = a.jwt.RequestBefore(ctx, info, req)
		if err != nil {
			return nil, err
		}
	} else {
		ctx, err = a.ak.RequestBefore(ctx, info, req)
		if err != nil {
			return nil, err
		}
	}
	return handler(ctx, req)
}

func (a *Auth) StreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !auth.IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata not exists in context")
	}
	xApiKey:=md.Get("x-api-key")
	if len(xApiKey) != 0 {
		return a.apikey.StreamInterceptor(srv, ss, info, handler)
	}
	authorization := a.jwt.GetAuthorizationFromMetadata(md)

	if authorization == "" {
		return errors.New(errors.ErrTokenMissing, int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	var err error
	if strings.Contains(authorization, "Bearer") {
		ss, err = a.jwt.StreamRequestBefore(ss.Context(), ss, info, nil)
		if err != nil {
			return err
		}
		return handler(srv, ss)

	} else {
		ss, err = a.ak.StreamRequestBefore(ss.Context(), ss, info, nil)
		if err != nil {
			return err
		}
	}
	return handler(srv, ss)
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
