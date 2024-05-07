package auth

import (
	"context"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AccessKeyAuth struct {
	app    biz.AppRepo
	config *config.Config
	// rdb        *tiga.RedisDao
	localCache *data.LayeredCache
	log        logger.Logger
	priority   int
	name       string
}

func NewAccessKeyAuth(app biz.AppRepo, config *config.Config, local *data.LayeredCache, log logger.Logger) *AccessKeyAuth {
	return &AccessKeyAuth{
		app:        app,
		config:     config,
		localCache: local,
		log:        log,
		name:       "ak_auth",
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
func (a *AccessKeyAuth) getUid(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	uid, ok := md["x-uid"]
	if !ok {
		return ""
	}
	return uid[0]

}
func (a *AccessKeyAuth) RequestBefore(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) (context.Context, error) {
	gwRequest, err := gosdk.NewGatewayRequestFromGrpc(ctx, req, info.FullMethod)
	if err != nil {
		return ctx, status.Errorf(codes.InvalidArgument, "parse request error,%v", err)
	}
	accessKey, err := a.appValidator(ctx, gwRequest)
	if err != nil {
		return ctx, err

	}
	if a.getUid(ctx) != "" {
		return ctx, nil
	}
	appid, err := a.getAppid(ctx, accessKey)
	if err != nil {
		return ctx, err

	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md =metadata.MD{}
	}
	md.Set("x-identity", appid)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx, nil

}
func (a *AccessKeyAuth) appValidator(ctx context.Context, req *gosdk.GatewayRequest) (string, error) {
	xDate := ""
	auth := ""
	accessKey := ""
	for _, k := range req.Headers.Keys() {
		v := req.Headers.Get(k)
		if strings.EqualFold(k, gosdk.HeaderXDateTime) {
			xDate = v
		}
		if strings.EqualFold(k, gosdk.HeaderXAuthorization) {
			auth = v
		}
		if strings.EqualFold(k, gosdk.HeaderXAccessKey) {
			accessKey = v

		}
	}
	if xDate == "" {
		return "", errors.New(errors.ErrAppXDateMissing, int32(api.APPSvrCode_APP_XDATE_MISSING_ERR), codes.Unauthenticated, "app_timestamp")
	}
	if auth == "" {
		return "", errors.New(errors.ErrAppSignatureMissing, int32(api.APPSvrCode_APP_AUTH_MISSING_ERR), codes.Unauthenticated, "app_signature")
	}
	if accessKey == "" {
		return "", errors.New(errors.ErrAppAccessKeyMissing, int32(api.APPSvrCode_APP_ACCESS_KEY_MISSING_ERR), codes.Unauthenticated, "app_access_key")
	}
	t, err := time.Parse(gosdk.DateFormat, xDate)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "parse %s error,%v", gosdk.HeaderXDateTime, err)
	}
	// check timestamp
	if time.Since(t) > time.Minute*1 {
		return "", errors.New(errors.ErrRequestExpired, int32(api.APPSvrCode_APP_REQUEST_EXPIRED_ERR), codes.DeadlineExceeded, "app_timestamp")
	}
	secret, err := a.getSecret(ctx, accessKey)
	// a.log.Info("secret:", secret)
	if err != nil {
		return "", errors.New(err, int32(api.APPSvrCode_APP_UNKONW), codes.Unauthenticated, "app_secret")
	}
	signer := gosdk.NewAppAuthSigner(accessKey, secret)
	// a.log.Infof("req:%v,%v,%v,%v,%v", req.Headers, req.Host, req.Method, req.Host, req.URL.String())
	sign, err := signer.Sign(req)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "sign error,%v", err)
	}
	if sign != a.getSignature(auth) {
		return "", errors.New(errors.ErrAppSignatureInvalid, int32(api.APPSvrCode_APP_SIGNATURE_ERR), codes.Unauthenticated, "app签名校验")
	}
	return accessKey, nil
}
func (a *AccessKeyAuth) getSecret(ctx context.Context, accessKey string) (string, error) {
	cacheKey := a.config.GetAPPAccessKey(accessKey)
	secretBytes, err := a.localCache.Get(ctx, cacheKey)
	secret := string(secretBytes)
	if err != nil {
		apps, err := a.app.Get(ctx, accessKey)
		if err != nil {
			return "", err
		}
		secret = apps.Secret

		// _ = a.rdb.Set(ctx, cacheKey, secret, time.Hour*24*3)
		_ = a.localCache.Set(ctx, cacheKey, []byte(secret), time.Hour*24*3)
	}
	return secret, nil
}
func (a *AccessKeyAuth) getAppid(ctx context.Context, accessKey string) (string, error) {
	cacheKey := a.config.GetAppidKey(accessKey)
	secretBytes, err := a.localCache.Get(ctx, cacheKey)
	appid := string(secretBytes)
	if err != nil {
		apps, err := a.app.Get(ctx, accessKey)
		if err != nil {
			return "", err
		}
		appid = apps.Appid

		// _ = a.rdb.Set(ctx, cacheKey, secret, time.Hour*24*3)
		_ = a.localCache.Set(ctx, cacheKey, []byte(appid), time.Hour*24*3)
	}
	return appid, nil
}
func (a *AccessKeyAuth) getSignature(auth string) string {
	strArr := strings.Split(auth, ",")
	for _, v := range strArr {
		if strings.Contains(strings.ToLower(v), "signature") {
			signature := strings.Split(v, "=")
			if len(signature) == 2 {
				return signature[1]
			}
		}
	}
	return ""
}
func (a *AccessKeyAuth) ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	gwRequest, err := gosdk.NewGatewayRequestFromGrpc(ctx, req, fullName)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse request error,%v", err)
	}
	accessKey, err := a.appValidator(ctx, gwRequest)
	if err != nil {
		return ctx, err

	}
	if a.getUid(ctx) != "" {
		return ctx, nil
	}
	appid, err := a.getAppid(ctx, accessKey)
	if err != nil {
		return ctx, err

	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md =metadata.MD{}
	}
	md.Set("x-identity", appid)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx, nil
}
func (a *AccessKeyAuth) StreamRequestBefore(ctx context.Context, ss grpc.ServerStream, info *grpc.StreamServerInfo, req interface{}) (grpc.ServerStream, error) {
	grpcStream := NewGrpcStream(ss, info.FullMethod, ss.Context(), a)
	// defer grpcStream.Release()
	return grpcStream, nil

}
func (a *AccessKeyAuth) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if !IfNeedValidate(ctx, info.FullMethod) {
		return handler(ctx, req)
	}
	ctx, err = a.RequestBefore(ctx, info, req)
	if err != nil {
		return nil, err

	}
	resp, err = handler(ctx, req)
	if err != nil {
		err = a.ResponseAfter(ctx, info, req, resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, err
}
func (a *AccessKeyAuth) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !IfNeedValidate(ss.Context(), info.FullMethod) {
		return handler(srv, ss)
	}
	grpcStream, err := a.StreamRequestBefore(ss.Context(), ss, info, srv)
	if err != nil {
		return err
	}
	err = handler(srv, grpcStream)
	if err != nil {
		err = a.StreamResponseAfter(ss.Context(), ss, info)
		if err != nil {
			return err
		}
	}
	return err

}
func (a *AccessKeyAuth) ResponseAfter(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}, resp interface{}) error {
	return nil
}
func (a *AccessKeyAuth) StreamResponseAfter(ctx context.Context, ss grpc.ServerStream, info *grpc.StreamServerInfo) error {
	if grpcStream, ok := ss.(*grpcServerStream); ok {
		grpcStream.Release()
	}
	return nil
}

func (a *AccessKeyAuth) SetPriority(priority int) {
	a.priority = priority
}
func (a *AccessKeyAuth) Priority() int {
	return a.priority
}
func (a *AccessKeyAuth) Name() string {
	return a.name
}
