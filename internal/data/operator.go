package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	app "github.com/begonia-org/go-sdk/api/app/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spark-lence/tiga"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
)

type dataOperatorRepo struct {
	data  *Data
	app   biz.AppRepo
	authz biz.AuthzRepo
	user  biz.UserRepo
	local *LayeredCache
	log   logger.Logger
}

func NewDataOperatorRepo(data *Data, app biz.AppRepo,user biz.UserRepo, authz biz.AuthzRepo, local *LayeredCache, log logger.Logger) biz.DataOperatorRepo {
	log.WithField("module", "dataOperatorRepo")
	log.SetReportCaller(true)
	return &dataOperatorRepo{data: data, app: app, authz: authz,user: user, local: local, log: log}
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
func (r *dataOperatorRepo) GetAllForbiddenUsers(ctx context.Context) ([]*api.Users, error) {
	users := make([]*api.Users, 0)
	// return r.data.Get(ctx, key)
	page := int32(1)
	for {
		user, err := r.user.List(ctx, nil, []api.USER_STATUS{api.USER_STATUS_LOCKED, api.USER_STATUS_DELETED}, page, 100)
		// users, err := r.user.ListUsers(ctx, "status in (?,?)", api.USER_STATUS_LOCKED, api.USER_STATUS_DELETED)
		if err != nil {
			if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
				break
			}
			return users, err
		}
		if len(user) == 0 {
			break
		}
		page++
		users = append(users, user...)
	}

	return users, nil
}

func (r *dataOperatorRepo) GetAllApps(ctx context.Context) ([]*app.Apps, error) {
	apps := make([]*app.Apps, 0)
	page := int32(1)
	for {
		app, err := r.app.List(ctx, nil, nil, page, 100)
		if err != nil {
			if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
				break
			}
			return apps, err
		}
		if len(app) == 0 {
			break

		}
		apps = append(apps, app...)
		page++
	}

	return apps, nil
}

func (d *dataOperatorRepo) FlashAppsCache(ctx context.Context, prefix string, models []*app.Apps, exp time.Duration) error {

	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:access_key:%s", prefix, model.AccessKey)
		kv = append(kv, key, model.Secret)
		kv = append(kv, fmt.Sprintf("%s:appid:%s", prefix, model.AccessKey), model.Appid)
	}
	pipe := d.data.BatchCacheByTx(ctx, exp, kv...)
	_, err := pipe.Exec(ctx)
	return err
}
func (d *dataOperatorRepo) FlashAppidCache(ctx context.Context, prefix string, models []*app.Apps, exp time.Duration) error {

	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.AccessKey)
		kv = append(kv, key, model.Appid)
	}
	pipe := d.data.BatchCacheByTx(ctx, exp, kv...)
	_, err := pipe.Exec(ctx)
	return err
}
func (d *dataOperatorRepo) FlashUsersCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error {

	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Uid)
		val, _ := protojson.Marshal(model)
		kv = append(kv, key, string(val))
	}
	pipe := d.data.BatchCacheByTx(ctx, exp, kv...)
	// 记录最后更新时间
	pipe.Set(ctx, fmt.Sprintf("%s:last_updated", prefix), time.Now().UnixMilli(), exp)
	_, err := pipe.Exec(ctx)
	return err
}
func (d *dataOperatorRepo) LoadAppsLayeredCache(ctx context.Context, prefix string, models []*app.Apps, exp time.Duration) error {
	for _, model := range models {
		key := fmt.Sprintf("%s:access_key:%s", prefix, model.AccessKey)
		if err := d.local.Set(ctx, key, []byte(model.Secret), exp); err != nil {
			return err
		}
		key = fmt.Sprintf("%s:appid:%s", prefix, model.AccessKey)
		if err := d.local.Set(ctx, key, []byte(model.Appid), exp); err != nil {
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

func (r *dataOperatorRepo) CacheUsers(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error {

	pipe := r.user.Cache(ctx, prefix, models, exp, func(user *api.Users) ([]byte, interface{}) {
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
		if err == redis.Nil {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	// 将毫秒转换为秒和纳秒
	sec := timestamp / 1000          // 秒
	nsec := (timestamp % 1000) * 1e6 // 剩余的毫秒转换为纳秒

	// 使用秒和纳秒创建一个time.Time对象
	t := time.Unix(sec, nsec)
	return t, nil
}

func (d *dataOperatorRepo) Watcher(ctx context.Context, prefix string, handle func(ctx context.Context, op mvccpb.Event_EventType, key, value string) error) error {
	// prefix := d.local.config.GetEndpointsPrefix()
	// prefix = filepath.Join(prefix, "details")
	watcher := d.data.etcd.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for wresp := range watcher {
		for _, ev := range wresp.Events {
			val := ev.Kv.Value
			if ev.Type == mvccpb.DELETE {
				val = ev.PrevKv.Value

			}
			err := handle(ctx, ev.Type, string(ev.Kv.Key), string(val))
			if err != nil {
				d.log.Error(err)
			}

		}
	}
	return nil
}
