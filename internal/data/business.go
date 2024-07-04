package data

import (
	"context"
	"fmt"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/spark-lence/tiga"
)

type businessRepoImpl struct {
	data *Data
	cfg  *config.Config
	curd biz.CURD
}

func NewBusinessRepoImpl(data *Data, curd biz.CURD, cfg *config.Config) biz.BusinessRepo {
	return &businessRepoImpl{data: data, cfg: cfg, curd: curd}
}

func (b *businessRepoImpl) Add(ctx context.Context, business *api.Business) error {
	return b.curd.Add(ctx, business, false, nil)
}
func (b *businessRepoImpl) Get(ctx context.Context, key string) (*api.Business, error) {
	business := &api.Business{}
	err := b.curd.Get(ctx, business, false, "business_id=? or business_name=?", key, key)
	if err != nil || business.BusinessId == "" {
		return nil, fmt.Errorf("get business failed: %w or not found business", err)
	}
	return business, nil
}
func (b *businessRepoImpl) Del(ctx context.Context, key string) error {
	business, err := b.Get(ctx, key)
	if err != nil {
		return err
	}
	return b.curd.Del(ctx, business, false, nil)
}
func (b *businessRepoImpl) List(ctx context.Context, tags []string, page, pageSize int32) ([]*api.Business, error) {
	businesses := make([]*api.Business, 0)
	query := ""
	conds := make([]interface{}, 0)
	if len(tags) > 0 {
		query = "json_contains(json_array(?),tags)"
		conds = append(conds, tags)
	}
	pagination := &tiga.Pagination{
		Page:     page,
		PageSize: pageSize,
		Query:    query,
		Args:     conds,
	}
	err := b.curd.List(ctx, &businesses, pagination)
	if err != nil {
		return nil, err
	}
	return businesses, nil
}
func (b *businessRepoImpl) Patch(ctx context.Context, model *api.Business) error {
	return b.curd.Update(ctx, model, false, nil)
}
