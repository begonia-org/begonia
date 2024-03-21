package service

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

func (u *UsersService) AuthSeed(ctx context.Context, in *api.AuthLogAPIRequest) (*api.AuthLogAPIResponse, error) {
	token, err := u.biz.AuthSeed(ctx, in)
	if err != nil {
		return nil, err
	}
	rsp:= &api.AuthLogAPIResponse{
		Msg:       token,
		Timestamp: in.Timestamp,
	}
	return rsp, nil

}

func (u *UsersService) Login(ctx context.Context, in *api.LoginAPIRequest) (*api.LoginAPIResponse, error) {
	rsp, err := u.biz.Login(ctx, in)

	return rsp, err
}

func (u *UsersService) Logout(ctx context.Context, req *api.LogoutAPIRequest) (*api.LogoutAPIResponse, error) {
	err := u.biz.Logout(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.LogoutAPIResponse{}, nil

}

func (u *UsersService) Account(ctx context.Context, req *api.AccountAPIRequest) (*api.AccountAPIResponse, error) {
	rsp, err := u.biz.Account(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.AccountAPIResponse{
		Users: rsp,
	}, nil
}
func (u *UsersService) Register(context.Context, *api.RegsiterAPIRequest) (*api.RegsiterAPIResponse, error) {
	return nil, nil
}

func (u *UsersService) Desc() *grpc.ServiceDesc {
	return &api.AuthService_ServiceDesc
}
