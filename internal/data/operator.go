package data

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	app "github.com/begonia-org/go-sdk/api/app/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/redis/go-redis/v9"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
)

type dataOperatorRepo struct {
	data        *Data
	app         biz.AppRepo
	authz       biz.AuthzRepo
	user        biz.UserRepo
	local       *LayeredCache
	log         logger.Logger
	onceOnStart sync.Once
}

func NewDataOperatorRepo(data *Data, app biz.AppRepo, user biz.UserRepo, authz biz.AuthzRepo, local *LayeredCache, log logger.Logger) biz.DataOperatorRepo {
	log.WithField("module", "dataOperatorRepo")
	log.SetReportCaller(true)
	return &dataOperatorRepo{data: data,
		app:         app,
		authz:       authz,
		user:        user,
		local:       local,
		log:         log,
		onceOnStart: sync.Once{},
	}
}

// DistributedLock 分布式锁,拿到锁对象后需要调用DistributedUnlock释放锁
func (r *dataOperatorRepo) Locker(ctx context.Context, key string, exp time.Duration) (biz.DataLock, error) {
	return NewDataLock(r.data.rdb.GetClient(), key, exp, 3), nil
}

// GetAllForbiddenUsers 获取所有被禁用的用户
func (r *dataOperatorRepo) GetAllForbiddenUsers(ctx context.Context) ([]*api.Users, error) {
	users := make([]*api.Users, 0)
	page := int32(1)
	for {
		user, err := r.user.List(ctx, nil, []api.USER_STATUS{api.USER_STATUS_LOCKED, api.USER_STATUS_DELETED, api.USER_STATUS_INACTIVE}, page, 100)
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
	pipe.Set(ctx, fmt.Sprintf("%s:last_updated", prefix), time.Now().UnixMilli(), exp)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < len(kv)-1; i += 2 {
		key := kv[i].(string)
		val := kv[i+1].(string)
		_ = d.local.SetToLocal(ctx, key, []byte(val), exp)
	}
	return err
}

func (d *dataOperatorRepo) FlashUsersCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error {

	kv := make([]interface{}, 0)
	for _, model := range models {
		key := fmt.Sprintf("%s:%s", prefix, model.Uid)
		val, _ := protojson.Marshal(model)
		kv = append(kv, key, string(val))
	}
	// d.local.Set()
	pipe := d.data.BatchCacheByTx(ctx, exp, kv...)
	// 记录最后更新时间
	pipe.Set(ctx, fmt.Sprintf("%s:last_updated", prefix), time.Now().UnixMilli(), exp)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < len(kv)-1; i += 2 {
		key := kv[i].(string)
		val := kv[i+1].(string)
		// d.log.Infof("set to local %s %s", key, val)
		_ = d.local.SetToLocal(ctx, key, []byte(val), exp)
	}
	return err
}

func (d *dataOperatorRepo) LoadRemoteCache(ctx context.Context) {

	err := d.local.kv.LoadDump(ctx)
	if err != nil {
		d.log.Errorf("load remote cache error %v", err)
	}
	err = d.local.filters.LoadDump(ctx)
	if err != nil {
		d.log.Errorf("load remote cache error %v", err)

	}

}

func (d *dataOperatorRepo) LastUpdated(ctx context.Context, key string) (time.Time, error) {
	key = fmt.Sprintf("%s:last_updated", key)
	timestamp, err := d.data.rdb.GetClient().Get(ctx, key).Int64()
	if err != nil||timestamp==0 {
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

func (d *dataOperatorRepo) Watcher(ctx context.Context, prefix string, handle biz.EtcdWatchHandle) error {
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

func (d *dataOperatorRepo) Sync() {
	ticker := time.NewTicker(5 * time.Minute) // 5分钟同步一次
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			d.LoadRemoteCache(context.Background())
		}
	}()

}

func (d *dataOperatorRepo) OnStart() {
	d.onceOnStart.Do(func() {
		// l.Sync()
		d.log.Info("data operator repo start")
		errChan := d.local.filters.Watch(context.Background())
		go func() {
			for err := range errChan {
				d.log.Errorf("bloom filter error %v", err)
			}
		}()
		errKvChan := d.local.kv.Watch(context.Background())
		go func() {
			for err := range errKvChan {
				d.local.log.Errorf("kv error %v", err)
			}
		}()
	})
}
