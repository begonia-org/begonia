package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type AuthzRepo interface {
	CacheToken(ctx context.Context, key, token string, exp time.Duration) error
	GetToken(ctx context.Context, key string) string
	DelToken(ctx context.Context, key string) error
	CheckInBlackList(ctx context.Context, key string) (bool, error)
	PutBlackList(ctx context.Context, token string) error
}

type AuthzUsecase struct {
	repo       AuthzRepo
	log        logger.Logger
	authCrypto *crypto.UsersAuth
	config     *config.Config
	user       UserRepo
}

func NewAuthzUsecase(repo AuthzRepo, user UserRepo, log logger.Logger, crypto *crypto.UsersAuth, config *config.Config) *AuthzUsecase {
	return &AuthzUsecase{repo: repo, log: log, authCrypto: crypto, config: config, user: user}
}

func (u *AuthzUsecase) DelToken(ctx context.Context, key string) error {
	return u.repo.DelToken(ctx, key)
}

func (u *AuthzUsecase) AuthSeed(ctx context.Context, in *api.AuthLogAPIRequest) (string, error) {
	token, err := u.authCrypto.GenerateAuthSeed(in.Token)
	if err != nil {
		return "", gosdk.NewError(fmt.Errorf("auth seed generate %w", err), int32(api.UserSvrCode_USER_LOGIN_ERR), codes.InvalidArgument, "generate_seed")
	}
	return token, nil

}
func (u *AuthzUsecase) PutBlackList(ctx context.Context, token string) error {
	return u.repo.PutBlackList(ctx, token)
}
func (u *AuthzUsecase) CheckInBlackList(ctx context.Context, token string) (bool, error) {
	return u.repo.CheckInBlackList(ctx, token)
}
func (u *AuthzUsecase) getUserAuth(_ context.Context, in *api.LoginAPIRequest) (*api.UserAuth, error) {

	timestamp := in.Seed / 10000
	now := time.Now().Unix()
	if now-timestamp > 60 {

		return nil, gosdk.NewError(pkg.ErrTokenExpired, int32(api.UserSvrCode_USER_TOKEN_EXPIRE_ERR.Number()), codes.InvalidArgument, "种子有效期校验")
	}
	auth := in.Auth
	authBytes, err := u.authCrypto.RSADecrypt(auth)

	if err != nil {
		return nil, gosdk.NewError(pkg.ErrAuthDecrypt, int32(api.UserSvrCode_USER_AUTH_DECRYPT_ERR.Number()), codes.InvalidArgument, "login_info_rsa")
	}
	userAuth := &api.UserAuth{}
	err = json.Unmarshal([]byte(authBytes), userAuth)
	if err != nil {
		return nil, gosdk.NewError(pkg.ErrDecode, int32(common.Code_AUTH_ERROR), codes.InvalidArgument, "login_info_decode")
	}
	return userAuth, nil
}
func (u *AuthzUsecase) GenerateJWT(ctx context.Context, user *api.Users, isKeepLogin bool) (string, error) {
	exp := time.Duration(u.config.GetJWTExpiration()) * time.Second
	if isKeepLogin {
		exp = time.Hour * 24 * 3
	}
	secret := u.config.GetString("auth.jwt_secret")
	validateToken := tiga.ComputeHmacSha256(fmt.Sprintf("%s:%d", user.Uid, time.Now().Unix()), secret)
	payload := &api.BasicAuth{
		Uid:         user.Uid,
		Name:        user.Name,
		Role:        user.Role,
		Audience:    user.Name,
		Expiration:  time.Now().Add(exp).Unix(),
		Issuer:      "gateway",
		NotBefore:   time.Now().Unix(),
		IssuedAt:    time.Now().Unix(),
		IsKeepLogin: isKeepLogin,
		Token:       validateToken,
	}
	// err := u.repo.DelToken(ctx, u.config.GetUserBlackListKey(user.Uid))

	token, err := tiga.GenerateJWT(payload, secret)
	if err != nil {
		return "", gosdk.NewError(err, int32(api.UserSvrCode_USER_UNKNOWN), codes.Internal, "jwt_generate")

	}
	return token, nil
}

func (u *AuthzUsecase) Login(ctx context.Context, in *api.LoginAPIRequest) (*api.LoginAPIResponse, error) {
	// 解密账号密码
	userAuth, err := u.getUserAuth(ctx, in)
	if err != nil {
		return nil, err
	}
	// 登陆验证
	key, iv := u.config.GetAesConfig()
	account, err := tiga.EncryptAES([]byte(key), userAuth.Account, iv)
	if err != nil {
		err := gosdk.NewError(pkg.ErrEncrypt, int32(api.UserSvrCode_USER_ACCOUNT_ERR), codes.InvalidArgument, "accout_encrypt")
		return nil, err
	}

	user, err := u.user.Get(ctx, account)
	if err != nil || user == nil {
		if err == nil || strings.Contains(err.Error(), "not found") {
			err = pkg.ErrUserNotFound
		}
		err := gosdk.NewError(err, int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "user_query")
		return nil, err
	}
	if user.Password != userAuth.Password {
		err := gosdk.NewError(pkg.ErrUserPasswordInvalid, int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "password_match")
		return nil, err
	}
	if user.Status != api.USER_STATUS_ACTIVE {
		err := gosdk.NewError(pkg.ErrUserDisabled, int32(api.UserSvrCode_USER_DISABLED_ERR), codes.Unauthenticated, "user_query")
		return nil, err

	}
	user.Password = ""
	// 生成jwt
	token, err := u.GenerateJWT(ctx, user, in.IsKeepLogin)
	if err != nil {
		return nil, err
	}
	return &api.LoginAPIResponse{
		User:  user,
		Token: token,
	}, nil
}

func (u *AuthzUsecase) Logout(ctx context.Context, req *api.LogoutAPIRequest) error {

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return gosdk.NewError(pkg.ErrNoMetadata, int32(common.Code_METADATA_MISSING), codes.InvalidArgument, "metadata_missing")
	}
	token := md.Get("x-token")
	if len(token) == 0 {
		return gosdk.NewError(pkg.ErrTokenMissing, int32(common.Code_TOKEN_NOT_FOUND), codes.InvalidArgument, "token_missing")
	}
	err := u.PutBlackList(ctx, tiga.GetMd5(token[0]))
	if err != nil {
		return gosdk.NewError(err, int32(common.Code_AUTH_ERROR), codes.Internal, "add_black_list")
	}
	return nil

}
