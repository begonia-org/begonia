package data

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	api "github.com/wetrycode/begonia/api/v1"
	"github.com/wetrycode/begonia/internal/biz"
)

type userRepo struct {
	data  *Data
	log   *logrus.Logger
	local *LocalCache
}

func NewUserRepo(data *Data, log *logrus.Logger, local *LocalCache) biz.UsersRepo {
	return &userRepo{data: data, log: log, local: local}
}

func (r *userRepo) ListUsers(ctx context.Context, conds ...interface{}) ([]*api.Users, error) {
	users := make([]*api.Users, 0)
	if err := r.data.List(&api.Users{}, &users, conds...); err != nil {
		return nil, err
	}
	return users, nil

}
func (r *userRepo) CreateUsers(ctx context.Context, users []*api.Users) error {
	sources := NewSourceTypeArray(users)
	return r.data.CreateInBatches(sources)
}
func (t *userRepo) UpdateUsers(ctx context.Context, models []*api.Users) error {
	sources := NewSourceTypeArray(models)

	return t.data.BatchUpdates(sources, &api.Users{})
}

func (t *userRepo) DeleteUsers(ctx context.Context, models []*api.Users) error {
	sources := NewSourceTypeArray(models)

	return t.data.BatchDelete(sources, &api.Users{})
}
func (t *userRepo) CacheToken(ctx context.Context, key, token string, exp time.Duration) error {
	_ = t.local.Set(ctx, key, []byte(token))
	return t.data.Cache(ctx, key, token, exp)
}
func (t *userRepo) GetToken(ctx context.Context, key string) string {
	token, _ := t.local.Get(ctx, key)
	if token != nil {
		return string(token)
	}

	val := t.data.GetCache(ctx, key)
	if val != "" {
		_ = t.local.Set(ctx, key, []byte(val))
	}
	return val
}
func (t *userRepo) DelToken(ctx context.Context, key string) error {
	_ = t.local.Del(ctx, key)
	err := t.data.DelCache(ctx, key)
	return err
}
func (t *userRepo) CheckInBlackList(ctx context.Context, key string) (bool, error) {
	val, err := t.local.Get(ctx, key)
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, errors.Wrap(err, "获取本地缓存失败")
	}
	return string(val) != "", nil
}
