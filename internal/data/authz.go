package data

import (
	"context"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/go-sdk/logger"
)

type authzRepo struct {
	data  *Data
	log   logger.Logger
	local *LayeredCache
}

func NewAuthzRepoImpl(data *Data, log logger.Logger, local *LayeredCache) biz.AuthzRepo {
	return &authzRepo{data: data, log: log, local: local}
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
