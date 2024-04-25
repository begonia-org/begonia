package data

import (
	"context"
	"fmt"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	api "github.com/begonia-org/go-sdk/api/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type appRepoImpl struct {
	data  *Data
	local *LayeredCache
}

func NewAppRepoImpl(data *Data, local *LayeredCache) biz.AppRepo {
	return &appRepoImpl{data: data, local: local}
}

func (r *appRepoImpl) AddApps(ctx context.Context, apps []*api.Apps) (*gorm.DB, error) {

	for _, app := range apps {
		app.CreatedAt = timestamppb.New(time.Now())
		app.UpdatedAt = timestamppb.New(time.Now())
	}
	// sources := NewSourceTypeArray(apps)
	return r.data.CreateInBatchesByTx(apps)
}
func (r *appRepoImpl) GetApps(ctx context.Context, keys []string) ([]*api.Apps, error) {
	apps := make([]*api.Apps, 0)

	sql := "appid in (?) or access_key in (?)"
	if len(keys) == 0 {
		return apps, nil
	}
	if len(keys) == 1 {
		sql = "appid = ? or access_key = ?"
	}
	err := r.data.List(&api.Apps{}, &apps, sql, keys)
	return apps, err
}

func (r *appRepoImpl) CacheApps(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) redis.Pipeliner {
	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Key)
		err := r.local.Set(ctx, key, []byte(model.Secret), 0)
		if err != nil {
			return nil
		}
		kv = append(kv, key, model.Secret)
	}
	return r.data.BatchCacheByTx(ctx, exp, kv...)
}
func (r *appRepoImpl) DelAppsCache(ctx context.Context, prefix string, models []*api.Apps) error {
	keys := make([]string, 0)
	for _, model := range models {
		err := r.local.Del(ctx, fmt.Sprintf("%s:%s", prefix, model.Key))
		if err != nil {
			return err
		}
		keys = append(keys, fmt.Sprintf("%s:%s", prefix, model.Key))
	}
	pipe := r.data.DelCacheByTx(ctx, keys...)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (r *appRepoImpl) ListApps(ctx context.Context, conds ...interface{}) ([]*api.Apps, error) {
	apps := make([]*api.Apps, 0)
	if err := r.data.List(&api.Apps{}, &apps, conds...); err != nil {
		return nil, err
	}
	return apps, nil

}
