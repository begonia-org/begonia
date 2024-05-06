package service

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
)

type AuthzService struct {
	biz    *biz.AuthzUsecase
	log    logger.Logger
	config *config.Config
	api.UnimplementedAuthServiceServer
	authCrypto *crypto.UsersAuth
}

func NewAuthzService(biz *biz.AuthzUsecase, log logger.Logger, auth *crypto.UsersAuth, config *config.Config) *AuthzService {
	return &AuthzService{biz: biz, log: log, authCrypto: auth, config: config}
}

func (u *AuthzService) AuthSeed(ctx context.Context, in *api.AuthLogAPIRequest) (*api.AuthLogAPIResponse, error) {
	token, err := u.biz.AuthSeed(ctx, in)
	if err != nil {
		return nil, err
	}
	rsp := &api.AuthLogAPIResponse{
		Msg:       token,
		Timestamp: in.Token,
	}
	return rsp, nil

}

func (u *AuthzService) Login(ctx context.Context, in *api.LoginAPIRequest) (*api.LoginAPIResponse, error) {
	rsp, err := u.biz.Login(ctx, in)

	return rsp, err
}

func (u *AuthzService) Logout(ctx context.Context, req *api.LogoutAPIRequest) (*api.LogoutAPIResponse, error) {
	err := u.biz.Logout(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.LogoutAPIResponse{}, nil

}



func (u *AuthzService) Desc() *grpc.ServiceDesc {
	return &api.AuthService_ServiceDesc
}
