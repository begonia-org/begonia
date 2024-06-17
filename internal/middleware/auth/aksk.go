package auth

import (
	"context"
	"strings"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AccessKeyAuthMiddleware struct {
	app      *biz.AccessKeyAuth
	config   *config.Config
	log      logger.Logger
	priority int
	name     string
}

func NewAccessKeyAuth(app *biz.AccessKeyAuth, config *config.Config, log logger.Logger) *AccessKeyAuthMiddleware {
	return &AccessKeyAuthMiddleware{
		app:    app,
		config: config,
		// localCache: local,
		log:  log,
		name: "ak_auth",
	}
}

func IfNeedValidate(ctx context.Context, fullMethod string) bool {
	routersList := routers.Get()
	router := routersList.GetRouteByGrpcMethod(strings.ToUpper(fullMethod))
	if router == nil {
		return false
	}
	return router.AuthRequired

}

func (a *AccessKeyAuthMiddleware) RequestBefore(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) (context.Context, error) {
	gwRequest, err := gosdk.NewGatewayRequestFromGrpc(ctx, req, info.FullMethod)
	if err != nil {
		return ctx, status.Errorf(codes.InvalidArgument, "parse request error,%v", err)
	}
	accessKey, err := a.app.AppValidator(ctx, gwRequest)
	if err != nil {
		return ctx, err

	}

	owner, err := a.app.GetAppOwner(ctx, accessKey)
	if err != nil {
		return ctx, err

	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	// md.Set(gosdk.HeaderXIdentity, owner)
	md = metadata.Join(md, metadata.Pairs(gosdk.HeaderXIdentity, owner))

	ctx = metadata.NewIncomingContext(ctx, md)
	// md2, _ := metadata.FromIncomingContext(ctx)


	return ctx, nil

}

func (a *AccessKeyAuthMiddleware) ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	ctx,err:= a.RequestBefore(ctx, &grpc.UnaryServerInfo{FullMethod: fullName}, req)
	if err!=nil{
		return ctx,err
	}
	md, _ := metadata.FromIncomingContext(ctx)
	if identity := md.Get(gosdk.HeaderXIdentity);len(identity)>0{
		headers.Set(strings.ToLower(gosdk.HeaderXIdentity), identity[0])
	}
	return ctx,nil
	
}
func (a *AccessKeyAuthMiddleware) StreamRequestBefore(ctx context.Context, ss grpc.ServerStream, info *grpc.StreamServerInfo, req interface{}) (grpc.ServerStream, error) {
	grpcStream := NewGrpcStream(ss, info.FullMethod, ss.Context(), a)
	// defer grpcStream.Release()
	return grpcStream, nil

}
func (a *AccessKeyAuthMiddleware) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if !IfNeedValidate(ctx, info.FullMethod) {
		return handler(ctx, req)
	}
	ctx, err = a.RequestBefore(ctx, info, req)
	if err != nil {
		return nil, err

	}
	defer func() {
		_ = a.ResponseAfter(ctx, info, req, resp)
	}()
	resp, err = handler(ctx, req)

	return resp, err
}
func (a *AccessKeyAuthMiddleware) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	grpcStream, err := a.StreamRequestBefore(ss.Context(), ss, info, srv)
	if err != nil {
		return err
	}
	defer func() {
		err := a.StreamResponseAfter(ss.Context(), ss, info)
		if err != nil {
			a.log.Errorf(ss.Context(), "StreamResponseAfter error,%s", err.Error())
		}
	}()
	err = handler(srv, grpcStream)

	return err

}
func (a *AccessKeyAuthMiddleware) ResponseAfter(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}, resp interface{}) error {
	return nil
}
func (a *AccessKeyAuthMiddleware) StreamResponseAfter(ctx context.Context, ss grpc.ServerStream, info *grpc.StreamServerInfo) error {
	if grpcStream, ok := ss.(*grpcServerStream); ok {
		grpcStream.Release()
	}
	return nil
}

func (a *AccessKeyAuthMiddleware) SetPriority(priority int) {
	a.priority = priority
}
func (a *AccessKeyAuthMiddleware) Priority() int {
	return a.priority
}
func (a *AccessKeyAuthMiddleware) Name() string {
	return a.name
}
