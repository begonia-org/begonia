package service

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
)

type UserService struct {
	api.UnimplementedUserServiceServer
	biz    *biz.UserUsecase
	log    logger.Logger
	config *config.Config
}

func NewUserService(biz *biz.UserUsecase, log logger.Logger, config *config.Config) *UserService {
	return &UserService{biz: biz, log: log, config: config}
}

func (u *UserService) Register(ctx context.Context, in *api.PostUserRequest) (*api.Users, error) {
	owner := GetIdentity(ctx)
	if in.Owner != "" {
		owner = in.Owner
	}
	user := &api.Users{
		Name:     in.Name,
		Password: in.Password,
		Email:    in.Email,
		Phone:    in.Phone,
		Role:     in.Role,
		Status:   in.Status,
		Dept:     in.Dept,
		Owner:    owner,
		Avatar:   in.Avatar,
	}
	err := u.biz.Add(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) Update(ctx context.Context, in *api.PostUserRequest) (*api.Users, error) {
	owner := GetIdentity(ctx)
	if in.Owner != "" {
		owner = in.Owner
	
	}
	user := &api.Users{
		Name:       in.Name,
		Password:   in.Password,
		Email:      in.Email,
		Phone:      in.Phone,
		Role:       in.Role,
		Status:     in.Status,
		Dept:       in.Dept,
		Owner:      owner,
		Avatar:     in.Avatar,
		UpdateMask: in.UpdateMask,
	}
	err := u.biz.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) Get(ctx context.Context, in *api.GetUserRequest) (*api.Users, error) {
	user, err := u.biz.Get(ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) Delete(ctx context.Context, in *api.DeleteUserRequest) (*api.DeleteUserResponse, error) {
	err := u.biz.Delete(ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	return &api.DeleteUserResponse{}, nil
}

func (app *UserService) Desc() *grpc.ServiceDesc {
	return &api.UserService_ServiceDesc
}
