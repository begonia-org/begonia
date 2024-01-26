package middlerware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
	srvErr "github.com/spark-lence/tiga/errors"
	api "github.com/wetrycode/begonia/api/v1"
	"github.com/wetrycode/begonia/internal/biz"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/errors"
	"github.com/wetrycode/begonia/internal/pkg/routers"
	"github.com/wetrycode/begonia/internal/pkg/web"
)

type APIVildator struct {
	biz     biz.UsersRepo
	routers *routers.HttpURIRouteToSrvMethod
	config  *config.Config
	rdb     *tiga.RedisDao
	log     *logrus.Logger
}

func NewAPIVildator(rdb *tiga.RedisDao, log *logrus.Logger, repo biz.UsersRepo, config *config.Config) *APIVildator {
	routers := routers.NewHttpURIRouteToSrvMethod()
	routers.LoadAllRouters()
	return &APIVildator{
		rdb:     rdb,
		routers: routers,
		config:  config,
		log:     log,
		biz:     repo,
	}
}

func (a *APIVildator) writeRsp(w http.ResponseWriter, requestId string, err error) {
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
func (a *APIVildator) JWTLock(uid string) (*redislock.Lock, error) {
	key := a.config.GetJWTLockKey(uid)
	return a.rdb.Lock(context.Background(), key, time.Second*10)
}

func (a *APIVildator) checkJWTItem(ctx context.Context, payload *api.BasicAuth, token string) (bool, error) {
	if payload.Expiration < time.Now().Unix() {
		return false, srvErr.New(errors.ErrTokenExpired, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_EXPIRE_ERR), "请重新登陆"))
	}
	if payload.NotBefore > time.Now().Unix() {
		remain := payload.NotBefore - time.Now().Unix()
		msg := fmt.Sprintf("请%d秒后重试", remain)
		return false, srvErr.New(errors.ErrTokenNotActive, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_NOT_ACTIVTE_ERR), msg))
	}
	if payload.Issuer != "gateway" {
		return false, srvErr.New(errors.ErrTokenIssuer, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "请重新登陆"))
	}
	if ok, err := a.biz.CheckInBlackList(ctx, token); ok || err != nil {
		return false, srvErr.New(errors.ErrTokenBlackList, "登陆状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "非法的token"))
	}
	return true, nil
}

func (a *APIVildator) jwt2BasicAuth(authorization string) (*api.BasicAuth, error) {
	// Typically JWT is in a header in the format "Bearer {token}"
	strArr := strings.Split(authorization, " ")
	token := ""
	if len(strArr) == 2 {
		token = strArr[1]
	}
	if token == "" {
		return nil, srvErr.New(errors.ErrHeaderTokenFormat, "Token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "缺少token"))
	}
	jwtInfo := strings.Split(token, ".")
	if len(jwtInfo) != 3 {
		return nil, srvErr.New(errors.ErrHeaderTokenFormat, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "非法的token"))
	}
	// 生成signature
	sig := fmt.Sprintf("%s.%s", jwtInfo[0], jwtInfo[1])
	secret := a.config.GetJWTSecret()

	sig = tiga.ComputeHmacSha256(sig, secret)
	if sig != jwtInfo[2] {
		return nil, srvErr.New(errors.ErrTokenInvalid, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "非法的token"))
	}
	payload := &api.BasicAuth{}
	payloadBytes, err := tiga.Base64URL2Bytes(jwtInfo[1])
	if err != nil {
		return nil, srvErr.New(errors.ErrAuthDecrypt, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "非法的token"))
	}
	err = json.Unmarshal(payloadBytes, payload)
	if err != nil {
		return nil, srvErr.New(errors.ErrDecode, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "非法的token"))
	}
	return payload, nil
}

// checkJWT 校验jwt
// 提前刷新token
func (a *APIVildator) checkJWT(authorization string, w http.ResponseWriter, r *http.Request) (bool, error) {
	payload, err := a.jwt2BasicAuth(authorization)
	if err != nil {
		return false, err
	}
	strArr := strings.Split(authorization, " ")
	token := strArr[1]
	ok, err := a.checkJWTItem(context.Background(), payload, token)
	if err != nil || !ok {
		return false, err
	}
	// a.data.ch
	// 15min快过期了，刷新token
	left := payload.Expiration - time.Now().Unix()
	if left <= 1*60*15 {
		lock, err := a.JWTLock(payload.Uid)
		defer func() {
			if p := recover(); p != nil {
				a.log.Errorf("刷新token失败,%s", p)
			}
			if lock != nil {
				err := lock.Release(context.Background())
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
		_ = a.biz.CacheToken(context.Background(), a.config.GetUserBlackListKey(tiga.GetMd5(token)), "1", time.Hour*24*3)
		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", newToken))
		token = newToken

	}
	// 设置uid
	r.Header.Set("x-token", token)
	r.Header.Set("x-uid", payload.Uid)
	return true, nil

}
func (a *APIVildator) Vildator(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		router := a.routers.GetRoute(uri)
		reqId := w.Header().Get("x-request-id")
		if reqId == "" {
			reqId = uuid.New().String()
		}

		defer func() {
			a.log.WithFields(logrus.Fields{
				"x-request-id": reqId,
				"uri":          uri,
				"method":       r.Method,
				"remote_addr":  r.RemoteAddr,
			}).Info("请求开始")
		}()
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
					// 校验token
					ok, err := a.checkJWT(token[0], w, r)
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
					a.writeRsp(w, reqId, srvErr.New(errors.ErrTokenMissing, "token状态校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_INVILDAT_ERR), "缺少token")))
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}
