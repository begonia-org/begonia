package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/v1"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/internal/pkg/web"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
	srvErr "github.com/spark-lence/tiga/errors"
)

type APIValidator struct {
	biz        *biz.UsersUsecase
	app        biz.AppRepo
	config     *config.Config
	rdb        *tiga.RedisDao
	localCache *data.LayeredCache
	log        *logrus.Logger
}

var onceValidate sync.Once
var validator *APIValidator

func NewAPIValidator(rdb *tiga.RedisDao, log *logrus.Logger, user *biz.UsersUsecase, config *config.Config, mysql *tiga.MySQLDao, local *data.LayeredCache) *APIValidator {
	onceValidate.Do(func() {
		validator = &APIValidator{
			rdb:        rdb,
			config:     config,
			log:        log,
			biz:        user,
			localCache: local,
		}
	})
	return validator

}

func (a *APIValidator) writeRsp(w http.ResponseWriter, requestId string, err error) {
	w.Header().Set("x-request-id", requestId)
	rsp, err := web.MakeResponse(nil, err)
	data := []byte("")
	if err != nil {
		a.log.Errorf("序列化响应失败,%s", err.Error())
	} else {
		data, _ = json.Marshal(rsp)

	}
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(data)

}
func (a *APIValidator) JWTLock(uid string) (*redislock.Lock, error) {
	key := a.config.GetJWTLockKey(uid)
	return a.rdb.Lock(context.Background(), key, time.Second*10)
}

func (a *APIValidator) checkJWTItem(ctx context.Context, payload *api.BasicAuth, token string) (bool, error) {
	if payload.Expiration < time.Now().Unix() {
		return false, srvErr.New(errors.ErrTokenExpired, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_EXPIRE_ERR), "请重新登陆"))
	}
	if payload.NotBefore > time.Now().Unix() {
		remain := payload.NotBefore - time.Now().Unix()
		msg := fmt.Sprintf("请%d秒后重试", remain)
		return false, srvErr.New(errors.ErrTokenNotActive, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_NOT_ACTIVTE_ERR), msg))
	}
	if payload.Issuer != "gateway" {
		return false, srvErr.New(errors.ErrTokenIssuer, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "请重新登陆"))
	}
	if ok, err := a.biz.CheckInBlackList(ctx, tiga.GetMd5(token)); ok {
		if err != nil {
			return false, err
		}
		return false, srvErr.New(errors.ErrTokenBlackList, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "非法的token"))
	}
	return true, nil
}

func (a *APIValidator) jwt2BasicAuth(authorization string) (*api.BasicAuth, error) {
	// Typically JWT is in a header in the format "Bearer {token}"
	strArr := strings.Split(authorization, " ")
	token := ""
	if len(strArr) == 2 {
		token = strArr[1]
	}
	if token == "" {
		return nil, srvErr.New(errors.ErrHeaderTokenFormat, "Token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "缺少token"))
	}
	jwtInfo := strings.Split(token, ".")
	if len(jwtInfo) != 3 {
		return nil, srvErr.New(errors.ErrHeaderTokenFormat, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "非法的token"))
	}
	// 生成signature
	sig := fmt.Sprintf("%s.%s", jwtInfo[0], jwtInfo[1])
	secret := a.config.GetJWTSecret()

	sig = tiga.ComputeHmacSha256(sig, secret)
	if sig != jwtInfo[2] {
		return nil, srvErr.New(errors.ErrTokenInvalid, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "非法的token"))
	}
	payload := &api.BasicAuth{}
	payloadBytes, err := tiga.Base64URL2Bytes(jwtInfo[1])
	if err != nil {
		return nil, srvErr.New(errors.ErrAuthDecrypt, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "非法的token"))
	}
	err = json.Unmarshal(payloadBytes, payload)
	if err != nil {
		return nil, srvErr.New(errors.ErrDecode, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "非法的token"))
	}
	return payload, nil
}

// checkJWT 校验jwt
// 提前刷新token
func (a *APIValidator) checkJWT(ctx context.Context, authorization string, rspHeader Header, reqHeader Header) (ok bool, err error) {
	payload, err := a.jwt2BasicAuth(authorization)
	if err != nil {
		return false, err
	}
	strArr := strings.Split(authorization, " ")
	token := strArr[1]
	ok, err = a.checkJWTItem(ctx, payload, token)
	if err != nil || !ok {
		return false, err
	}

	left := payload.Expiration - time.Now().Unix()
	expiration := a.config.GetJWTExpiration()
	// 10%的时间刷新token
	if left <= int64(float64(expiration)*0.1) {
		// 锁住之后再刷新
		lock, err := a.JWTLock(payload.Uid)
		defer func() {
			if p := recover(); p != nil {
				ok = false
				err = fmt.Errorf("刷新token失败,%v", p)
				a.log.Errorf("刷新token失败,%s", p)
			}
			if lock != nil {
				err := lock.Release(ctx)
				if err != nil {
					a.log.Errorf("释放锁失败,%s", err.Error())
				}
			}
		}()
		// 正在刷新token
		if err == redislock.ErrNotObtained {
			return true, nil
		}

		exp := time.Hour * 2
		if payload.IsKeepLogin {
			exp = time.Hour * 24 * 3
		}
		payload.Expiration = time.Now().Add(exp).Unix()
		payload.NotBefore = time.Now().Unix()
		payload.IssuedAt = time.Now().Unix()
		secret := a.config.GetJWTSecret()
		newToken, err := tiga.GenerateJWT(payload, secret)
		if err != nil {
			return false, srvErr.New(err, "刷新token")
		}
		// 旧token加入黑名单
		go a.biz.PutBlackList(ctx, a.config.GetUserBlackListKey(tiga.GetMd5(token)))
		rspHeader.SendHeader("Authorization", fmt.Sprintf("Bearer %s", newToken))
		token = newToken

	}
	// 设置uid
	reqHeader.Set("x-token", token)
	reqHeader.Set("x-uid", payload.Uid)
	return true, nil

}
func (a *APIValidator) getAuthorizationFromMetadata(md metadata.MD) string {

	for k, v := range md {
		if strings.EqualFold(k, "authorization") {
			return v[0]
		}
	}
	return ""
}
// func (a *APIValidator) ValidateStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
// 	if !a.IfNeedValidate(ss.Context(), info.FullMethod) {
// 		return handler(srv, ss)
// 	}
// 	grpcStream := NewGrpcStream(ss, info.FullMethod, ss.Context(),a)
// 	defer grpcStream.Release()
// 	return handler(srv, grpcStream)
// }
func (a *APIValidator) ValidateGrpcRequest(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	gwRequest, err := gosdk.NewGatewayRequestFromGrpc(ctx, req, fullName)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse request error,%v", err)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")
	}
	authorization := a.getAuthorizationFromMetadata(md)

	if authorization == "" {
		return nil, errors.New(errors.ErrTokenMissing, int32(api.UserSvrCode_USER_AUTH_MISSING_ERR), codes.Unauthenticated, "authorization_check")
	}
	// JWT
	if strings.Contains(authorization, "Bearer") {
		ctx, err := a.GrpcJTWValidator(ctx, req, fullName, headers)
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}
	// AK/SK
	err = a.AppValidator(ctx, gwRequest)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}
func (a *APIValidator) IfNeedValidate(ctx context.Context, fullMethod string) bool {
	routersList := routers.Get()
	router := routersList.GetRouteByGrpcMethod(strings.ToUpper(fullMethod))
	if router == nil {
		return false
	}
	return router.AuthRequired

}
func (a *APIValidator) ValidateUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !a.IfNeedValidate(ctx, info.FullMethod) {
		return handler(ctx, req)

	}
	in, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")

	}
	out, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		out = metadata.MD{}
	}
	header := NewGrpcHeader(in, ctx, out)
	defer header.Release()
	_, err := a.ValidateGrpcRequest(ctx, req, info.FullMethod, header)
	if err != nil {
		return nil, err
	}
	return handler(header.ctx, req)
}

func (a *APIValidator) GrpcJTWValidator(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	// 获取请求的方法名
	fullMethodName := fullName
	// 获取路由
	routersList := routers.Get()
	router := routersList.GetRouteByGrpcMethod(fullMethodName)
	if router == nil {
		return nil, status.Errorf(codes.Unimplemented, "method not exists in context")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")

	}

	token := a.getAuthorizationFromMetadata(md)
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token not exists in context")
	}
	// serverStream.SetHeader(metadata.Pairs("x-request-id", reqId[0]))
	// header := &GrpcHeader{md}
	ok, err := a.checkJWT(ctx, token, headers, headers)
	newCtx := metadata.NewIncomingContext(ctx, md)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "check token error,%v", err)
	}
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "token check failed")
	}
	return newCtx, nil
	// return handler(newCtx, req)
}

func (a *APIValidator) HttpHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		routersList := routers.Get()
		router := routersList.GetRoute(uri)
		reqId := w.Header().Get("x-request-id")
		if reqId == "" {
			reqId = uuid.New().String()
		}

		if router != nil {

			if router.AuthRequired {
				logger := a.log.WithFields(logrus.Fields{
					"x-request-id": reqId,
					"uri":          uri,
					"method":       r.Method,
					"remote_addr":  r.RemoteAddr,
					"status":       http.StatusUnauthorized,
				})
				if token, ok := r.Header[http.CanonicalHeaderKey("Authorization")]; ok {
					if !strings.Contains(token[0], "Bearer") {
						gwReq, _ := gosdk.NewGatewayRequestFromHttp(r)
						err := a.AppValidator(r.Context(), gwReq)
						if err != nil {
							logger.Warn("app校验失败")
							a.writeRsp(w, reqId, err)
							return
						}
					}
					// 校验token
					header := &httpHeader{w, r}
					ok, err := a.checkJWT(r.Context(), token[0], header, header)
					if err != nil {
						a.writeRsp(w, reqId, err)
						return
					}
					if !ok {
						logger.Warn("token校验失败")
						a.writeRsp(w, reqId, err)
						return
					}
				} else {
					logger.Warn("请求头缺失Authorization")
					a.writeRsp(w, reqId, srvErr.New(errors.ErrTokenMissing, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), "缺少token")))
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}
func (a *APIValidator) getSecret(ctx context.Context, accessKey string) (string, error) {
	cacheKey := a.config.GetAPPAccessKey(accessKey)
	secretBytes, err := a.localCache.Get(ctx, cacheKey)
	secret := string(secretBytes)
	if err != nil {
		apps, err := a.app.GetApps(ctx, []string{accessKey})
		if err != nil {
			return "", err
		}
		if len(apps) > 0 {
			secret = apps[0].Secret
		}
		// _ = a.rdb.Set(ctx, cacheKey, secret, time.Hour*24*3)
		_ = a.localCache.Set(ctx, cacheKey, []byte(secret), time.Hour*24*3)
	}
	return secret, nil
}
func (a *APIValidator) getSignature(auth string) string {
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
func (a *APIValidator) AppValidator(ctx context.Context, req *gosdk.GatewayRequest) error {
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
		return errors.New(errors.ErrAppXDateMissing, int32(api.APPSvrCode_APP_XDATE_MISSING_ERR), codes.Unauthenticated, "app_timestamp")
	}
	if auth == "" {
		return errors.New(errors.ErrAppSignatureMissing, int32(api.APPSvrCode_APP_AUTH_MISSING_ERR), codes.Unauthenticated, "app_signature")
	}
	if accessKey == "" {
		return errors.New(errors.ErrAppAccessKeyMissing, int32(api.APPSvrCode_APP_ACCESS_KEY_MISSING_ERR), codes.Unauthenticated, "app_access_key")
	}
	t, err := time.Parse(gosdk.DateFormat, xDate)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "parse %s error,%v", gosdk.HeaderXDateTime, err)
	}
	// check timestamp
	if time.Since(t) > time.Minute*1 {
		return errors.New(errors.ErrRequestExpired, int32(api.APPSvrCode_APP_REQUEST_EXPIRED_ERR), codes.DeadlineExceeded, "app_timestamp")
	}
	secret, err := a.getSecret(ctx, accessKey)
	// a.log.Info("secret:", secret)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "get secret error,%v", err)
	}
	signer := gosdk.NewAppAuthSigner(accessKey, secret)
	// a.log.Infof("req:%v,%v,%v,%v,%v", req.Headers, req.Host, req.Method, req.Host, req.URL.String())
	sign, err := signer.Sign(req)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "sign error,%v", err)
	}
	if sign != a.getSignature(auth) {
		return errors.New(errors.ErrAppSignatureInvalid, int32(api.APPSvrCode_APP_SIGNATURE_ERR), codes.Unauthenticated, "app签名校验")
	}
	return nil
}
