package biz

import (
	"context"
	"fmt"
	"strings"

	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
)

type BusinessRepo interface {
	Add(ctx context.Context, business *api.Business) error
	Get(ctx context.Context, key string) (*api.Business, error)
	Del(ctx context.Context, key string) error
	List(ctx context.Context, tags []string, page, pageSize int32) ([]*api.Business, error)
	Patch(ctx context.Context, model *api.Business) error
}

type BusinessUsecase struct {
	repo      BusinessRepo
	snowflake *tiga.Snowflake
}

func NewBusinessUsecase(repo BusinessRepo) *BusinessUsecase {
	snk, _ := tiga.NewSnowflake(1)
	return &BusinessUsecase{repo: repo, snowflake: snk}
}

func (u *BusinessUsecase) Add(ctx context.Context, in *api.PostBusinessRequest, createdBy string) (business *api.Business, err error) {
	defer func() {
		if err != nil {
			// log.Println(err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = gosdk.NewError(err, int32(common.Code_CONFLICT), codes.AlreadyExists, "commit_app")
			} else {
				err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "add_business")

			}
		}
	}()
	// business.BusinessId = u.snowflake.GenerateIDString()
	business = &api.Business{
		BusinessId:   u.snowflake.GenerateIDString(),
		BusinessName: in.BusinessName,
		Description:  in.Description,
		Tags:         in.Tags,
		CreatedBy:    createdBy,
	}

	err = u.repo.Add(ctx, business)
	return
}
func (u *BusinessUsecase) Get(ctx context.Context, key string) (*api.Business, error) {
	business, err := u.repo.Get(ctx, key)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("get business fail:%w", err), int32(common.Code_INTERNAL_ERROR), codes.NotFound, "get_business")
	}
	return business, nil
}
func (u *BusinessUsecase) Del(ctx context.Context, key string) error {
	err := u.repo.Del(ctx, key)
	if err != nil {
		return gosdk.NewError(fmt.Errorf("del business fail:%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "del_business")
	}
	return nil
}
func (u *BusinessUsecase) List(ctx context.Context, tags []string, page, pageSize int32) ([]*api.Business, error) {
	list, err := u.repo.List(ctx, tags, page, pageSize)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("list business fail:%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "list_business")

	}
	return list, nil
}
func (u *BusinessUsecase) Patch(ctx context.Context, model *api.Business) error {
	err := u.repo.Patch(ctx, model)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return gosdk.NewError(err, int32(common.Code_CONFLICT), codes.AlreadyExists, "patch_business")
		}
		return gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "patch_business")
	}
	return nil
}
