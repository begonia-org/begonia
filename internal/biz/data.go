package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/begonia-org/begonia/internal/biz/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	u "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/bsm/redislock"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/grpc/status"
)

type DataLock interface {
	UnLock(ctx context.Context) error
	Lock(ctx context.Context) error
}

type DataOperatorRepo interface {
	GetAllAppsFromDB(ctx context.Context) ([]*api.Apps, error)
	FlashAppsCache(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) error
	FlashUsersCache(ctx context.Context, prefix string, models []*u.Users, exp time.Duration) error
	// LoadAppsLocalCache(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) error
	GetAllForbiddenUsersFromDB(ctx context.Context) ([]*u.Users, error)
	// LoadUsersLocalCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error
	Lock(ctx context.Context, key string, exp time.Duration) (DataLock, error)
	LastUpdated(ctx context.Context, key string) (time.Time, error)
	Watcher(ctx context.Context, prefix string, handle func(ctx context.Context, op mvccpb.Event_EventType, key, value string) error) error
}
type operationAction func(ctx context.Context) error
type DataOperatorUsecase struct {
	repo            DataOperatorRepo
	endpoint        gateway.EndpointRepo
	config          *config.Config
	log             logger.Logger
	endpointWatcher *gateway.GatewayWatcher
}

func NewDataOperatorUsecase(repo DataOperatorRepo, config *config.Config, log logger.Logger, endpointWatch *gateway.GatewayWatcher, endpoint gateway.EndpointRepo) *DataOperatorUsecase {
	log.WithField("module", "data")
	log.SetReportCaller(true)
	return &DataOperatorUsecase{repo: repo, config: config, log: log, endpointWatcher: endpointWatch, endpoint: endpoint}
}

func (d *DataOperatorUsecase) Do(ctx context.Context) {
	// d.LoadCache(context.Background())
	err := d.OnStart(ctx)
	if err != nil {
		d.log.Error(err)
	}
	log.Println("start watch")
	d.handle(ctx)

}

func (d *DataOperatorUsecase) handle(ctx context.Context) {
	errChan := make(chan error, 3)
	wg := &sync.WaitGroup{}
	actions := []operationAction{
		d.loadUsersBlacklist,
		d.loadApps,
		d.watch,
		// d.loadLocalBloom,
	}
	for _, action := range actions {
		wg.Add(1)
		a := action
		go func(action operationAction) {
			defer wg.Done()
			errChan <- action(ctx)
		}(a)

	}
	go func() {
		for err := range errChan {
			if err != nil {
				if st, ok := status.FromError(err); ok {
					st.Details()
				}
				d.log.Error(err)
			}

		}
	}()
	wg.Wait()
	close(errChan)
}

// loadUsersBlacklist 加载用户黑名单
func (d *DataOperatorUsecase) loadUsersBlacklist(ctx context.Context) error {
	exp := d.config.GetUserBlackListExpiration() - 1
	if exp <= 0 {
		return fmt.Errorf("expiration time is too short")
	}
	lockKey := d.config.GetUserBlackListLockKey()
	lock, err := d.repo.Lock(ctx, lockKey, time.Second*time.Duration(exp))
	if err != nil {
		// d.log.Error("get lock error", err)
		return fmt.Errorf("get lock error: %w", err)

	}

	if err = lock.Lock(ctx); err != nil && err != redislock.ErrNotObtained {
		// d.log.Error("lock error:", err)
		return fmt.Errorf("lock error: %w", err)
	}
	defer func() {

		err = lock.UnLock(ctx)
		if err != nil {
			// d.log.Error("unlock error", err)
			d.log.Error(fmt.Errorf("unlock error: %w", err))

		}
	}()
	prefix := d.config.GetUserBlackListPrefix()
	lastUpdate, err := d.repo.LastUpdated(ctx, prefix)
	// 如果缓存时间小于3秒，说明刚刚更新过，不需要再次更新
	// 直接加载远程缓存到本地
	// lastUpdate ttl<exp,避免更新不到缓存的情况
	if lastUpdate.IsZero() || time.Since(lastUpdate) < 3*time.Second {
		users, err := d.repo.GetAllForbiddenUsersFromDB(ctx)
		if err != nil {
			return err
		}
		exp = d.config.GetUserBlackListExpiration()
		err = d.repo.FlashUsersCache(ctx, prefix, users, time.Duration(exp)*time.Second)
		if err != nil {
			return err
		}
	}

	// d.repo.LoadUsersLocalCache()
	return err
}

// loadApps 加载可用的app信息
func (d *DataOperatorUsecase) loadApps(ctx context.Context) error {
	apps, err := d.repo.GetAllAppsFromDB(ctx)
	if err != nil {
		return err
	}
	prefix := d.config.GetAPPAccessKeyPrefix()
	exp := d.config.GetAPPAccessKeyExpiration()
	return d.repo.FlashAppsCache(ctx, prefix, apps, time.Duration(exp)*time.Second)
}

func (d *DataOperatorUsecase) Refresh(duration time.Duration) {
	ticker := time.NewTicker(duration)
	for range ticker.C {
		d.handle(context.Background())
	}
}

func PutConfig(ctx context.Context, key string, value string) error {
	return nil
}
func (d *DataOperatorUsecase) OnStart(ctx context.Context) error {
	endpoints, err := d.endpoint.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("list endpoints error,%s", err.Error())
	}
	for _, in := range endpoints {
		bData, _ := json.Marshal(in)

		err := d.endpointWatcher.Update(ctx, in.UniqueKey, string(bData))
		if err != nil {
			d.log.Errorf("init endpoints error,%s", err.Error())
			continue
		}
	}
	return nil
}
func (d *DataOperatorUsecase) watch(ctx context.Context) error {
	prefix := d.config.GetServicePrefix()

	return d.repo.Watcher(context.Background(), prefix, d.endpointWatcher.Handle)
}
