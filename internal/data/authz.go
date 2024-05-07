package data

import (
	"context"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
)

type authzRepo struct {
	data  *Data
	log   logger.Logger
	local *LayeredCache
}

func NewAuthzRepo(data *Data, log logger.Logger, local *LayeredCache) biz.AuthzRepo {
	return &authzRepo{data: data, log: log, local: local}
}

func (r *authzRepo) ListUsers(ctx context.Context, page, pageSize int32, conds ...interface{}) ([]*api.Users, error) {
	users := make([]*api.Users, 0)
	if err := r.data.List(&api.Users{}, &users, page, pageSize, conds...); err != nil {
		return nil, err
	}
	return users, nil

}
func (r *authzRepo) GetUser(ctx context.Context, conds ...interface{}) (*api.Users, error) {
	user := &api.Users{}
	err := r.data.Get(user, user, conds...)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (r *authzRepo) CreateUsers(ctx context.Context, users []*api.Users) error {
	sources := NewSourceTypeArray(users)
	return r.data.CreateInBatches(sources)
}
func (t *authzRepo) UpdateUsers(ctx context.Context, models []*api.Users) error {
	sources := NewSourceTypeArray(models)

	return t.data.BatchUpdates(ctx,sources)
}

func (t *authzRepo) DeleteUsers(ctx context.Context, models []*api.Users) error {
	sources := NewSourceTypeArray(models)

	return t.data.BatchDelete(sources)
}
func (t *authzRepo) CacheToken(ctx context.Context, key, token string, exp time.Duration) error {
	return t.local.Set(ctx, key, []byte(token), exp)
}
func (t *authzRepo) GetToken(ctx context.Context, key string) string {
	token, _ := t.local.Get(ctx, key)
	if token != nil {
		return string(token)
	}

	return ""
}
func (t *authzRepo) DelToken(ctx context.Context, key string) error {
	err := t.local.Del(ctx, key)
	// err := t.data.DelCache(ctx, key)
	return err
}
func (t *authzRepo) CheckInBlackList(ctx context.Context, token string) (bool, error) {
	key := t.local.config.GetUserTokenBlackListBloom()
	return t.local.CheckInFilter(ctx, key, []byte(token))
}

func (t *authzRepo) PutBlackList(ctx context.Context, token string) error {
	key := t.local.config.GetUserTokenBlackListBloom()
	return t.local.AddToFilter(ctx, key, []byte(token))

}
