package data

import (
	"context"
	"fmt"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	api "github.com/begonia-org/go-sdk/api/v1"
	"github.com/spark-lence/tiga"
)

type dataOperatorRepo struct {
	data  *Data
	app   biz.AppRepo
	user  biz.UsersRepo
	local *LayeredCache
}

func NewDataOperatorRepo(data *Data, app biz.AppRepo, user biz.UsersRepo, local *LayeredCache, lock biz.DataLock) biz.DataOperatorRepo {
	return &dataOperatorRepo{data: data, app: app, user: user, local: local}
}

// DistributedLock 分布式锁,拿到锁对象后需要调用DistributedUnlock释放锁
func (r *dataOperatorRepo) Lock(ctx context.Context, key string, exp time.Duration) (biz.DataLock, error) {
	return NewDataLock(r.data.rdb.GetClient(), key, exp, 3), nil
}

// DistributedUnlock 释放锁
// func (r *dataOperatorRepo) DistributedUnlock(ctx context.Context, lock *redislock.Lock) error {
// 	return lock.Release(ctx)
// }

// GetAllForbiddenUsers 获取所有被禁用的用户
func (r *dataOperatorRepo) GetAllForbiddenUsersFromDB(ctx context.Context) ([]*api.Users, error) {

	// return r.data.Get(ctx, key)
	users, err := r.user.ListUsers(ctx, "status in (?,?)", api.USER_STATUS_LOCKED, api.USER_STATUS_DELETED)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *dataOperatorRepo) GetAllAppsFromDB(ctx context.Context) ([]*api.Apps, error) {
	apps, err := r.app.ListApps(ctx)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (d *dataOperatorRepo) FlashAppsCache(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) error {

	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Key)
		kv = append(kv, key, model.Secret)
	}
	pipe := d.data.BatchCacheByTx(ctx, exp, kv...)
	_, err := pipe.Exec(ctx)
	return err
}
func (d *dataOperatorRepo) FlashUsersCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error {

	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Uid)
		kv = append(kv, key, model)
	}
	pipe := d.data.BatchCacheByTx(ctx, exp, kv...)
	// 记录最后更新时间
	pipe.Set(ctx, fmt.Sprintf("%s:last_updated", prefix), time.Now().UnixMilli(), exp)
	_, err := pipe.Exec(ctx)
	return err
}
func (d *dataOperatorRepo) LoadAppsLayeredCache(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) error {
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Key)
		if err := d.local.Set(ctx, key, []byte(model.Secret), exp); err != nil {
			return err
		}
	}
	return nil
}
func (d *dataOperatorRepo) LoadUsersLayeredCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error {
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Uid)
		val, _ := tiga.IntToBytes(int(model.Status))
		if err := d.local.Set(ctx, key, val, exp); err != nil {
			return err
		}

	}
	return nil
}

// func (r *dataOperatorRepo) ScanUsersFromCache(ctx context.Context, prefix string, exp time.Duration) ([]*api.Users, error) {
// 	key := fmt.Sprintf("%s:*", prefix)
// 	var cursor uint64
// 	var n int
// for {

// }
// }
func (r *dataOperatorRepo) CacheUsers(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error {

	pipe := r.user.CacheUsers(ctx, prefix, models, exp, func(user *api.Users) ([]byte, interface{}) {
		val, _ := tiga.IntToBytes(int(user.Status))
		return val, int(user.Status)
	})
	_, err := pipe.Exec(ctx)
	return err
}
func (r *dataOperatorRepo) PullBloom(ctx context.Context, key string) []byte {
	return r.data.rdb.GetBytes(ctx, key)
}

// func (d *dataOperatorRepo) LoadLocalBloom(ctx context.Context, keys []*golayeredbloom.BloomConfig) error {
// 	return d.local.filters.LoadFrom(ctx, keys)
// }

func (d *dataOperatorRepo) LastUpdated(ctx context.Context, key string) (time.Time, error) {
	// return d.data.rdb.SetBytes(ctx, keys, exp)
	key = fmt.Sprintf("%s:last_updated", key)
	timestamp, err := d.data.rdb.GetClient().Get(ctx, key).Int64()
	if err != nil {
		return time.Time{}, err
	}
	// 将毫秒转换为秒和纳秒
	sec := timestamp / 1000          // 秒
	nsec := (timestamp % 1000) * 1e6 // 剩余的毫秒转换为纳秒

	// 使用秒和纳秒创建一个time.Time对象
	t := time.Unix(sec, nsec)
	return t, nil
}
