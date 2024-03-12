package data

import (
	"context"
	"fmt"
	"time"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type appRepoImpl struct {
	data  *Data
	local *LocalCache
}

func NewAppRepoImpl(data *Data, local *LocalCache) biz.AppRepo {
	return &appRepoImpl{data: data, local: local}
}

func (r *appRepoImpl) AddApps(ctx context.Context, apps []*api.Apps) (*gorm.DB, error) {
	sources := NewSourceTypeArray(apps)
	return r.data.CreateInBatchesByTx(sources)
}
func (r *appRepoImpl) GetApps(ctx context.Context, keys []string) ([]*api.Apps, error) {
	apps := make([]*api.Apps, 0)
	err := r.data.List(&api.Apps{}, &apps, "appid in (?) or access_key in (?)", keys)
	return apps, err
}

func (r *appRepoImpl) CacheApps(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) redis.Pipeliner {
	kv := make([]interface{}, 0)
	for _, model := range models {
		key:=fmt.Sprintf("%s:%s", prefix, model.AccessKey)
		err := r.local.cache.Set(key, []byte(model.Secret))
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
		err := r.local.cache.Delete(fmt.Sprintf("%s:%s", prefix, model.AccessKey))
		if err != nil {
			return err
		}
		keys = append(keys, fmt.Sprintf("%s:%s", prefix, model.AccessKey))
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
