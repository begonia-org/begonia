package data

import (
	"context"
	"fmt"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	api "github.com/begonia-org/go-sdk/api/v1"
	"github.com/redis/go-redis/v9"
)

type userRepo struct {
	data  *Data
	log   logger.Logger
	local *LayeredCache
}

func NewUserRepo(data *Data, log logger.Logger, local *LayeredCache) biz.UsersRepo {
	return &userRepo{data: data, log: log, local: local}
}

func (r *userRepo) ListUsers(ctx context.Context, conds ...interface{}) ([]*api.Users, error) {
	users := make([]*api.Users, 0)
	if err := r.data.List(&api.Users{}, &users, conds...); err != nil {
		return nil, err
	}
	return users, nil

}
func (r *userRepo) GetUser(ctx context.Context, conds ...interface{}) (*api.Users, error) {
	user := &api.Users{}
	err := r.data.Get(user, user, conds...)
	if err != nil {
		return nil, err
	}
	return user, nil
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
	return t.local.Set(ctx, key, []byte(token), exp)
}
func (t *userRepo) GetToken(ctx context.Context, key string) string {
	token, _ := t.local.Get(ctx, key)
	if token != nil {
		return string(token)
	}

	return ""
}
func (t *userRepo) DelToken(ctx context.Context, key string) error {
	err := t.local.Del(ctx, key)
	// err := t.data.DelCache(ctx, key)
	return err
}
func (t *userRepo) CheckInBlackList(ctx context.Context, token string) (bool, error) {
	key := t.local.config.GetUserTokenBlackListBloom()
	return t.local.CheckInFilter(ctx, key, []byte(token))
}

func (t *userRepo) PutBlackList(ctx context.Context, token string) error {
	key := t.local.config.GetUserTokenBlackListBloom()
	return t.local.AddToFilter(ctx, key, []byte(token))

}

func (u *userRepo) CacheUsers(ctx context.Context, prefix string, models []*api.Users, exp time.Duration, getValue func(user *api.Users) ([]byte, interface{})) redis.Pipeliner {
	kv := make([]interface{}, 0)
	for _, model := range models {
		// status,_:=tiga.IntToBytes(int(model.Status))
		valByte, val := getValue(model)
		key := fmt.Sprintf("%s:%s", prefix, model.Uid)
		if err := u.cacheUsers(ctx, prefix, model.Uid, valByte, exp); err != nil {
			return nil
		}
		kv = append(kv, key, val)
	}
	return u.data.BatchCacheByTx(ctx, exp, kv...)
}
func (u *userRepo) cacheUsers(ctx context.Context, prefix string, uid string, value []byte, exp time.Duration) error {
	key := fmt.Sprintf("%s:%s", prefix, uid)
	err := u.local.Set(ctx, key, value, exp)
	if err != nil {
		return err
	}
	return nil
}
