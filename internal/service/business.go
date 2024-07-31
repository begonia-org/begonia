package service

import (
	"context"
	"fmt"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type BusinessService struct {
	api.UnimplementedBusinessServiceServer
	business *biz.BusinessUsecase
	cfg      *config.Config
	log      logger.Logger
}

func NewBusinessService(business *biz.BusinessUsecase, log logger.Logger, cfg *config.Config) api.BusinessServiceServer {
	return &BusinessService{business: business, log: log, cfg: cfg}
}

func (b *BusinessService) Add(ctx context.Context, in *api.PostBusinessRequest) (*api.Business, error) {
	identity := GetIdentity(ctx)
	if identity == "" {
		return nil, gosdk.NewError(pkg.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")

	}
	return b.business.Add(ctx, in, identity)
}
func (b *BusinessService) Get(ctx context.Context, in *api.GetBusinessRequest) (*api.Business, error) {
	bs, err := b.business.Get(ctx, in.Business)

	return bs, err
}
func (b *BusinessService) Update(ctx context.Context, in *api.PatchBusinessRequest) (*api.Business, error) {
	bs := &api.Business{
		BusinessName: in.BusinessName,
		Description:  in.Description,
		Tags:         in.Tags,
		BusinessId:   in.BusinessId,
		UpdateMask:   in.UpdateMask,
	}
	err := b.business.Patch(ctx, bs)
	if err != nil {
		return nil, err
	}

	return b.Get(ctx, &api.GetBusinessRequest{Business: in.BusinessId})
}
func (b *BusinessService) Delete(ctx context.Context, in *api.DeleteBusinessRequest) (*api.DeleteBusinessResponse, error) {
	err := b.business.Del(ctx, in.Business)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("Delete business %s error:%w", in.Business, err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "delete business failed")
	}
	return &api.DeleteBusinessResponse{}, nil
}

func (b *BusinessService) List(ctx context.Context, in *api.ListBusinessRequest) (*api.ListBusinessResponse, error) {
	bs, err := b.business.List(ctx, in.Tags, in.Page, in.PageSize)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("List business error:%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "list business failed")
	}
	return &api.ListBusinessResponse{Business: bs}, nil
}
func (b *BusinessService) Desc() *grpc.ServiceDesc {
	return &api.BusinessService_ServiceDesc
}
