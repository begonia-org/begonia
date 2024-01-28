package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
	srvErr "github.com/spark-lence/tiga/errors"
	api "github.com/wetrycode/begonia/api/v1"
	common "github.com/wetrycode/begonia/common/api/v1"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/crypto"
	"github.com/wetrycode/begonia/internal/pkg/errors"
	"google.golang.org/grpc/metadata"
)

type UsersRepo interface {
	// mysql
	ListUsers(ctx context.Context, conds ...interface{}) ([]*api.Users, error)
	CreateUsers(ctx context.Context, Users []*api.Users) error
	UpdateUsers(ctx context.Context, models []*api.Users) error
	DeleteUsers(ctx context.Context, models []*api.Users) error

	// redis

	CacheToken(ctx context.Context, key, token string, exp time.Duration) error
	DelToken(ctx context.Context, key string) error
	CheckInBlackList(ctx context.Context, key string) (bool, error)
}

type UsersUsecase struct {
	repo       UsersRepo
	log        *logrus.Logger
	authCrypto *crypto.UsersAuth
	config     *config.Config
}

func NewUsersUsecase(repo UsersRepo, log *logrus.Logger, crypto *crypto.UsersAuth, config *config.Config) *UsersUsecase {
	return &UsersUsecase{repo: repo, log: log, authCrypto: crypto, config: config}
}
func (u *UsersUsecase) ListUsers(ctx context.Context, conds ...interface{}) ([]*api.Users, error) {
	return u.repo.ListUsers(ctx, conds...)
}

func (u *UsersUsecase) CreateUsers(ctx context.Context, Users []*api.Users) error {
	return u.repo.CreateUsers(ctx, Users)
}
func (u *UsersUsecase) UpdateUsers(ctx context.Context, conditions interface{}, model []*api.Users) error {
	return u.repo.UpdateUsers(ctx, model)
}

func (u *UsersUsecase) DeleteUsers(ctx context.Context, model []*api.Users) error {
	return u.repo.DeleteUsers(ctx, model)
}
func (u *UsersUsecase) CacheToken(ctx context.Context, key, token string, exp time.Duration) error {
	return u.repo.CacheToken(ctx, key, token, exp)
}
func (u *UsersUsecase) DelToken(ctx context.Context, key string) error {
	return u.repo.DelToken(ctx, key)
}

func (u *UsersUsecase) AuthSeed(ctx context.Context, in *api.AuthLogAPIRequest) (string, error) {
	token, err := u.authCrypto.GenerateAuthSeed(in.Timestamp)
	if err != nil {
		return "", srvErr.New(err, "auth seed生成错误", srvErr.WithMsgAndCode(int32(common.Code_INTERNAL_ERROR), "登陆失败"))
	}
	return token, nil

}
func (u *UsersUsecase) getUserAuth(ctx context.Context, in *api.LoginAPIRequest) (*api.UserAuth, error) {

	timestamp := in.Seed / 10000
	now := time.Now().Unix()
	if now-timestamp > 60 {
		err := srvErr.New(errors.ErrTokenExpired, "token校验", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_TOKEN_EXPIRE_ERR.Number()), "token过期"))
		return nil, err
	}
	auth := in.Auth
	authBytes, err := u.authCrypto.RSADecrypt(auth)

	if err != nil {
		err := srvErr.New(errors.ErrAuthDecrypt, "登陆", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_AUTH_DECRYPT_ERR.Number()), "登录失败"))
		return nil, err
	}
	userAuth := &api.UserAuth{}
	err = json.Unmarshal([]byte(authBytes), userAuth)
	if err != nil {
		err := srvErr.New(errors.ErrDecode, "登陆信息序列化", srvErr.WithMsgAndCode(int32(common.Code_AUTH_ERROR), "登录失败"))
		return nil, err
	}
	return userAuth, nil
}
func (u *UsersUsecase) generateJWT(ctx context.Context, user *api.Users, isKeepLogin bool) (string, error) {
	exp := time.Hour * 2
	if isKeepLogin {
		exp = time.Hour * 24 * 3
	}
	secret := u.config.GetString("auth.jwt_secret")
	vildateToken := tiga.ComputeHmacSha256(fmt.Sprintf("%s:%d", user.Uid, time.Now().Unix()), secret)
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
		Token:       vildateToken,
	}
	err := u.repo.DelToken(ctx, u.config.GetUserBlackListKey(user.Uid))
	if err != nil {
		return "", srvErr.New(errors.ErrRemoveBlackList, "生成新的JWT TOKEN", srvErr.WithMsgAndCode(int32(common.Code_AUTH_ERROR), "登录失败"))
	}

	return tiga.GenerateJWT(payload, secret)
}

func (u *UsersUsecase) Login(ctx context.Context, in *api.LoginAPIRequest) (*api.LoginAPIResponse, error) {
	// 解密账号密码
	userAuth, err := u.getUserAuth(ctx, in)
	if err != nil {
		return nil, err
	}
	// 登陆验证
	key, iv := u.config.GetAesConfig()
	account, err := tiga.EncryptAES([]byte(key), userAuth.Account, iv)
	if err != nil {
		err := srvErr.New(errors.ErrEncrypt, "账号加密", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_ACCOUNT_ERR), "账号或密码错误"))
		return nil, err
	}
	passwd, err := tiga.EncryptAES([]byte(key), userAuth.Password, iv)
	if err != nil {
		err := srvErr.New(errors.ErrEncrypt, "密码加密", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_PASSWORD_ERR), "账号或密码错误"))
		return nil, err
	}
	user := &api.Users{}
	users, err := u.repo.ListUsers(ctx, "(name = ? OR email = ? OR phone = ?) and password=(?)", account, account, account, passwd)
	if err != nil {
		err := srvErr.New(err, "查询用户", srvErr.WithMessage("账号或密码错误"))
		return nil, err
	}
	if len(users) == 0 {
		err := srvErr.New(errors.ErrUserNotFound, "查询用户", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_NOT_FOUND_ERR), "账号或密码错误"))
		return nil, err
	}
	user = users[0]
	user.Password = ""
	err = tiga.DecryptStructAES([]byte(key), user, iv)
	if err != nil {
		err := srvErr.New(err, "用户信息解密")
		return nil, err
	}
	// 生成jwt
	token, err := u.generateJWT(ctx, user, in.IsKeepLogin)
	if err != nil {
		return nil, err
	}
	return &api.LoginAPIResponse{
		User:  user,
		Token: token,
	}, nil
}

func (u *UsersUsecase) Logout(ctx context.Context, req *api.LogoutAPIRequest) error {

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		err := srvErr.New(errors.ErrNoMetadata, "登出账号", srvErr.WithMsgAndCode(int32(common.Code_METADATA_MISSING), "登出失败"))
		return err
	}
	token := md.Get("x-token")
	if len(token) == 0 {
		err := srvErr.New(errors.ErrTokenMissing, "登出账号", srvErr.WithMsgAndCode(int32(common.Code_TOKEN_NOT_FOUND), "登出失败"))
		return err
	}
	key := u.config.GetUserBlackListKey(tiga.GetMd5(token[0]))
	err := u.repo.CacheToken(ctx, key, "1", time.Hour*24*3)
	if err != nil {
		err = srvErr.New(err, "添加黑名单", srvErr.WithMessage("登出失败"))
		return err
	}
	return nil

}

func (u *UsersUsecase) Account(ctx context.Context, req *api.AccountAPIRequest) ([]*api.Users, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		err := srvErr.New(errors.ErrNoMetadata, "账号信息", srvErr.WithMsgAndCode(int32(common.Code_METADATA_MISSING), "请重新登陆"))
		return nil, err
	}
	token := md.Get("x-token")
	if len(token) == 0 {
		err := srvErr.New(errors.ErrTokenMissing, "账号信息", srvErr.WithMsgAndCode(int32(common.Code_TOKEN_NOT_FOUND), "请重新登陆"))
		return nil, err
	}
	clientUids := md.Get("x-uid")
	uid := ""

	if len(clientUids) > 0 {
		uid = clientUids[0]
	}
	if uid == "" {
		err := srvErr.New(errors.ErrUidMissing, "账号信息", srvErr.WithMsgAndCode(int32(common.Code_TOKEN_NOT_FOUND), "请重新登陆"))
		return nil, err
	}
	uids := req.Uids
	users, err := u.repo.ListUsers(ctx, "uid in (?)", uids)
	if err != nil {
		return nil, srvErr.New(err, "获取用户信息", srvErr.WithMessage("没有找到用户信息"))
	}
	if len(users) == 0 {
		return nil, srvErr.New(errors.ErrUserNotFound, "获取用户信息", srvErr.WithMsgAndCode(int32(api.UserSvrCode_USER_NOT_FOUND_ERR.Number()), "没有找到用户信息"))
	}
	return users, nil
}
