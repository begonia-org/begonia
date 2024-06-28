package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/bsm/redislock"
	"github.com/spark-lence/tiga"
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
		return nil, gosdk.NewError(pkg.ErrHeaderTokenFormat, int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "check_token")
	}
	jwtInfo := strings.Split(token, ".")
	if len(jwtInfo) != 3 {
		return nil, gosdk.NewError(pkg.ErrHeaderTokenFormat, int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "check_token_format")
	}
	// 生成signature
	sig := fmt.Sprintf("%s.%s", jwtInfo[0], jwtInfo[1])
	secret := a.config.GetJWTSecret()

	sig = tiga.ComputeHmacSha256(sig, secret)
	if sig != jwtInfo[2] {
		return nil, gosdk.NewError(pkg.ErrTokenInvalid, int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Internal, "check_sign")
	}
	payload := &api.BasicAuth{}
	payloadBytes, err := tiga.Base64URL2Bytes(jwtInfo[1])
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("%w:%w", pkg.ErrAuthDecrypt, err), int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "check_token")
	}
	err = json.Unmarshal(payloadBytes, payload)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("%w:%w", pkg.ErrDecode, err), int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "check_token")
	}
	return payload, nil
}
func (a *JWTAuth) JWTLock(uid string) (*redislock.Lock, error) {
	key := a.config.GetJWTLockKey(uid)
	return a.rdb.Lock(context.Background(), key, time.Second*10)
}

func (a *JWTAuth) checkJWTItem(ctx context.Context, payload *api.BasicAuth, token string) (bool, error) {
	if payload.Expiration < time.Now().Unix() {
		return false, gosdk.NewError(pkg.ErrTokenExpired, int32(api.UserSvrCode_USER_TOKEN_EXPIRE_ERR), codes.Internal, "check_expired")
	}
	if payload.NotBefore > time.Now().Unix() {
		remain := payload.NotBefore - time.Now().Unix()
		msg := fmt.Sprintf("Please retry after %d seconds ", remain)
		return false, gosdk.NewError(pkg.ErrTokenNotActive, int32(api.UserSvrCode_USER_TOKEN_NOT_ACTIVTE_ERR), codes.Canceled, "check_not_active", gosdk.WithClientMessage(msg))
	}
	if payload.Issuer != "gateway" {
		return false, gosdk.NewError(pkg.ErrTokenIssuer, int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "check_issuer")
	}
	if ok, err := a.biz.CheckInBlackList(ctx, tiga.GetMd5(token)); ok {
		return false, gosdk.NewError(fmt.Errorf("%w or %w", pkg.ErrTokenBlackList, err), int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "check_blacklist")

	}
	return true, nil
}
func (a *JWTAuth) checkJWT(ctx context.Context, authorization string, rspHeader Header, reqHeader Header) (ok bool, err error) {
	payload, errAuth := a.jwt2BasicAuth(authorization)
	err = errAuth
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
	// expiration := a.config.GetJWTExpiration()
	// 10%的时间刷新token
	if left <= int64(float64(payload.Expiration)*0.1) {
		// 锁住之后再刷新
		lock, errLock := a.JWTLock(payload.Uid)
		defer func() {
			if p := recover(); p != nil {
				ok = false
				err = fmt.Errorf("refresh token fail,%v", p)
				a.log.Errorf(ctx, "refresh token fail,%v", p)
			}
			if lock != nil {
				err := lock.Release(ctx)
				if err != nil {
					a.log.Errorf(ctx, "释放锁失败,%s", err.Error())
				}
			}
		}()
		// 正在刷新token
		err = errLock
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
			return false, gosdk.NewError(fmt.Errorf("%s:%w", "generate new token error", err), int32(api.UserSvrCode_USER_TOKEN_INVALIDATE_ERR), codes.Unauthenticated, "generate_token")
		}
		// 旧token加入黑名单
		go a.biz.PutBlackList(ctx, a.config.GetUserBlackListKey(tiga.GetMd5(token)))
		rspHeader.SendHeader("Authorization", fmt.Sprintf("Bearer %s", newToken))
		token = newToken

	}
	// 设置uid
	reqHeader.Set("x-token", token)
	reqHeader.Set("x-uid", payload.Uid)
	reqHeader.Set(gosdk.HeaderXIdentity, payload.Uid)
	return true, nil

}
func (a *JWTAuth) jwtValidator(ctx context.Context, headers Header) (context.Context, error) {
	

	md, _ := metadata.FromIncomingContext(ctx)

	token := a.GetAuthorizationFromMetadata(md)
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "token not exists in context")
	}

	ok, err := a.checkJWT(ctx, token, headers, headers)
	if err != nil || !ok {
		return nil, status.Errorf(codes.Unauthenticated, "check token error,%v", err)
	}

	newCtx := metadata.NewIncomingContext(ctx, md)
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
	_, err := a.jwtValidator(ctx, headers)
	if err != nil {
		return nil, err
	}
	return headers.ctx, nil
}

func (a *JWTAuth) ValidateStream(ctx context.Context, req interface{}, fullName string, headers Header) (context.Context, error) {
	// headers := NewGrpcStreamHeader(in, ctx, out,ss)
	ctx, err := a.jwtValidator(ctx, headers)
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
	defer func() {
		_ = a.ResponseAfter(ctx, info, req, resp)
	}()
	resp, err = handler(ctx, req)

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
	defer func() {
		err := a.StreamResponseAfter(ss.Context(), ss, info)
		if err != nil {
			a.log.Errorf(ss.Context(), "StreamResponseAfter error,%s", err.Error())
		}
	}()
	err = handler(srv, grpcStream)

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
