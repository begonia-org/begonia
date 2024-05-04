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
	key := r.cfg.GetAPPAccessKey(apps.AccessKey)
	exp := r.cfg.GetAPPAccessKeyExpiration()
	err = r.local.Set(ctx, key, []byte(apps.Secret), time.Duration(exp)*time.Second)
	return err
}
func (r *appRepoImpl) Get(ctx context.Context, key string) (*api.Apps, error) {

	app := &api.Apps{}
	err := r.data.Get(app, app, "(access_key = ? or appid=?) and is_deleted=0", key, key)
	return app, err
}

func (r *appRepoImpl) Cache(ctx context.Context, prefix string, models *api.Apps, exp time.Duration) error {

	key := fmt.Sprintf("%s:%s", prefix, models.AccessKey)
	return r.local.Set(ctx, key, []byte(models.Secret), exp)
}
func (r *appRepoImpl) Del(ctx context.Context, key string) error {
	app, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	_ = r.local.Del(ctx, r.cfg.GetAPPAccessKey(app.AccessKey))
	name := fmt.Sprintf("%s_%s", app.Name, time.Now().Format("20060102150405"))
	app.IsDeleted = true
	app.UpdatedAt = timestamppb.Now()
	app.Name = name
	app.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"is_deleted", "name"}}
	err = r.data.Update(ctx, app)
	return err
}
func (r *appRepoImpl) Patch(ctx context.Context, model *api.Apps) error {

	return r.data.Update(ctx, model)
}
func (r *appRepoImpl) List(ctx context.Context, conds ...interface{}) ([]*api.Apps, error) {
	apps := make([]*api.Apps, 0)
	if err := r.data.List(&api.Apps{}, &apps, conds...); err != nil {
		return nil, err
	}
	return apps, nil

}
