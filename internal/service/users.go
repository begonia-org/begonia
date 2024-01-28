package service

import (
	"context"

	"github.com/sirupsen/logrus"
	api "github.com/wetrycode/begonia/api/v1"
	common "github.com/wetrycode/begonia/common/api/v1"
	"github.com/wetrycode/begonia/internal/biz"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/crypto"
	"github.com/wetrycode/begonia/internal/pkg/web"
)

type UsersService struct {
	biz    *biz.UsersUsecase
	log    *logrus.Logger
	config *config.Config
	api.UnimplementedAuthServiceServer
	authCrypto *crypto.UsersAuth
}

func NewUserService(biz *biz.UsersUsecase, log *logrus.Logger, auth *crypto.UsersAuth, config *config.Config) *UsersService {
	return &UsersService{biz: biz, log: log, authCrypto: auth, config: config}
}

func (u *UsersService) AuthSeed(ctx context.Context, in *api.AuthLogAPIRequest) (*common.APIResponse, error) {
	token, err := u.biz.AuthSeed(ctx, in)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(&api.AuthLogAPIResponse{
		Msg:       token,
		Timestamp: in.Timestamp,
	}, nil)

}

func (u *UsersService) Login(ctx context.Context, in *api.LoginAPIRequest) (*common.APIResponse, error) {
	rsp, err := u.biz.Login(ctx, in)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(rsp, nil)
}

func (u *UsersService) Logout(ctx context.Context, req *api.LogoutAPIRequest) (*common.APIResponse, error) {
	err := u.biz.Logout(ctx, req)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(&api.LogoutAPIResponse{}, nil)

}

func (u *UsersService) Account(ctx context.Context, req *api.AccountAPIRequest) (*common.APIResponse, error) {
	rsp, err := u.biz.Account(ctx, req)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(&api.AccountAPIResponse{
		Users: rsp,
	}, nil)
}

