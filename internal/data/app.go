package data

import (
	"context"
	"fmt"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type appRepoImpl struct {
	data  *Data
	local *LayeredCache
	cfg   *config.Config
}

func NewAppRepoImpl(data *Data, local *LayeredCache, cfg *config.Config) biz.AppRepo {
	return &appRepoImpl{data: data, local: local, cfg: cfg}
}

func (r *appRepoImpl) Add(ctx context.Context, apps *api.Apps) error {
	apps.CreatedAt = timestamppb.Now()
	apps.UpdatedAt = timestamppb.Now()
	// sources := NewSourceTypeArray(apps)
	err := r.data.Create(apps)
	if err != nil {
		return err
	}
	key := r.cfg.GetAPPAccessKey(apps.Key)
	exp := r.cfg.GetAPPAccessKeyExpiration()
	err = r.local.Set(ctx, key, []byte(apps.Secret), time.Duration(exp)*time.Second)
	return err
}
func (r *appRepoImpl) Get(ctx context.Context, key string) (*api.Apps, error) {

	app := &api.Apps{}
	err := r.data.Get(app, app, "key = ?", key)
	return app, err
}

func (r *appRepoImpl) Cache(ctx context.Context, prefix string, models *api.Apps, exp time.Duration) error {
	// kv := make([]interface{}, 0)
	// for _, model := range models {
	// 	key := fmt.Sprintf("%s:%s", prefix, model.Key)
	// 	err := r.local.Set(ctx, key, []byte(model.Secret), 0)
	// 	if err != nil {
	// 		return nil
	// 	}
	// 	kv = append(kv, key, model.Secret)
	// }
	// return r.data.BatchCacheByTx(ctx, exp, kv...)
	key := fmt.Sprintf("%s:%s", prefix, models.Key)
	return r.local.Set(ctx, key, []byte(models.Secret), exp)
}
func (r *appRepoImpl) Del(ctx context.Context, key string) error {
	r.local.Del(ctx, r.cfg.GetAPPAccessKey(key))
	err := r.data.Update(ctx, &api.Apps{Key: key, IsDeleted: true, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"is_deleted"}}}, &api.Apps{})
	return err
}
func (r *appRepoImpl) Patch(ctx context.Context, model *api.Apps) error {

	return r.data.Update(ctx, model, model)
}
func (r *appRepoImpl) List(ctx context.Context, conds ...interface{}) ([]*api.Apps, error) {
	apps := make([]*api.Apps, 0)
	if err := r.data.List(&api.Apps{}, &apps, conds...); err != nil {
		return nil, err
	}
	return apps, nil

}
