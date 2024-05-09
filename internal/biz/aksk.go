package biz

import (
	"context"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccessKeyAuth struct {
	app    AppRepo
	config *config.Config
	log    logger.Logger
}

func NewAccessKeyAuth(app AppRepo, config *config.Config, log logger.Logger) *AccessKeyAuth {
	return &AccessKeyAuth{
		app:    app,
		config: config,
		log:    log,
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


func (a *AccessKeyAuth) AppValidator(ctx context.Context, req *gosdk.GatewayRequest) (string, error) {
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
	secret, err := a.app.GetSecret(ctx, accessKey)
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
func (a *AccessKeyAuth) GetSecret(ctx context.Context, accessKey string) (string, error) {
	secret, err := a.app.GetSecret(ctx, accessKey)
	if err != nil {
		return "", errors.New(err, int32(api.APPSvrCode_APP_UNKONW), codes.Unauthenticated, "app_secret")
	}
	return secret, nil
}

func (a *AccessKeyAuth)GetAppid(ctx context.Context, accessKey string) (string, error) {
	appid, err := a.app.GetAppid(ctx, accessKey)
	if err != nil {
		return "", errors.New(err, int32(api.APPSvrCode_APP_UNKONW), codes.Unauthenticated, "app_secret")
	}
	return appid, nil
}