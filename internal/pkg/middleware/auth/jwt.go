package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/bsm/redislock"
	"github.com/spark-lence/tiga"
	srvErr "github.com/spark-lence/tiga/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTAuth struct {
	config   *config.Config
	rdb      *tiga.RedisDao
	biz      *biz.AuthzUsecase
	log      logger.Logger
	priority int
	name     string
}

func NewJWTAuth(config *config.Config, rdb *tiga.RedisDao, biz *biz.AuthzUsecase, log logger.Logger) *JWTAuth {
	return &JWTAuth{
		config: config,
		rdb:    rdb,
		biz:    biz,
		log:    log,
		name:   "jwt_auth",
	}
}
func (a *JWTAuth) GetAuthorizationFromMetadata(md metadata.MD) string {

	for k, v := range md {
		if strings.EqualFold(k, "authorization") {
			return v[0]
		}
	}
	return ""
}
func (a *JWTAuth) jwt2BasicAuth(authorization string) (*api.BasicAuth, error) {
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
func (a *JWTAuth) JWTLock(uid string) (*redislock.Lock, error) {
	key := a.config.GetJWTLockKey(uid)
	return a.rdb.Lock(context.Background(), key, time.Second*10)
}

func (a *JWTAuth) checkJWTItem(ctx context.Context, payload *api.BasicAuth, token string) (bool, error) {
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
func (a *JWTAuth) checkJWT(ctx context.Context, authorization string, rspHeader Header, reqHeader Header) (ok bool, err error) {
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
func (a *JWTAuth) jwtValidator(ctx context.Context, fullName string, headers Header) (context.Context, error) {
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

	token := a.GetAuthorizationFromMetadata(md)
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token not exists in context")
	}

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

func (a *JWTAuth) RequestBefore(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) (context.Context, error) {
	in, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata not exists in context")

	}
	out, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		out = metadata.MD{}
	}
	headers := NewGrpcHeader(in, ctx, out)
	defer headers.Release()
	newCtx, err := a.jwtValidator(ctx, info.FullMethod, headers)
	if err != nil {
		return nil, err
	}
	return newCtx, nil
}

func (a *JWTAuth) ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	// headers := NewGrpcStreamHeader(in, ctx, out,ss)
	ctx, err := a.jwtValidator(ctx, fullName, headers)
	return ctx, err

}
func (a *JWTAuth) StreamRequestBefore(ctx context.Context, ss grpc.ServerStream, info *grpc.StreamServerInfo, req interface{}) (grpc.ServerStream, error) {
	grpcStream := NewGrpcStream(ss, info.FullMethod, ss.Context(), a)
	return grpcStream, nil
}

func (a *JWTAuth) ResponseAfter(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}, resp interface{}) error {
	return nil
}
func (a *JWTAuth) StreamResponseAfter(ctx context.Context, ss grpc.ServerStream, info *grpc.StreamServerInfo) error {
	if grpcStream, ok := ss.(*grpcServerStream); ok {
		grpcStream.Release()
	}
	return nil
}

func (a *JWTAuth) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
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

func (a *JWTAuth) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
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

func (jwt *JWTAuth) SetPriority(priority int) {
	jwt.priority = priority
}

func (jwt *JWTAuth) Priority() int {
	return jwt.priority
}

func (jwt *JWTAuth) Name() string {
	return jwt.name
}
