package middleware

import (
	"context"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/middleware/validator"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Auth struct {
	ak       *validator.AccessKeyAuth
	jwt      *validator.JWTAuth
	priority int
	name string
}

func NewAuth(ak *validator.AccessKeyAuth, jwt *validator.JWTAuth) gosdk.LocalPlugin {
	return &Auth{
		ak:  ak,
		jwt: jwt,
		name: "auth",
	}
}

func (a *Auth) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")
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
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata not exists in context")
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

func (a *Auth)Name() string {
	return a.name
}
