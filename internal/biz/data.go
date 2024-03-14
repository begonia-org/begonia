package biz

import (
	"context"
	"fmt"
	"sync"
	"time"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/bsm/redislock"
	"github.com/sirupsen/logrus"
)

type DataLock interface {
	UnLock(ctx context.Context) error
	Lock(ctx context.Context) error
}

type DataOperatorRepo interface {
	GetAllAppsFromDB(ctx context.Context) ([]*api.Apps, error)
	FlashAppsCache(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) error
	FlashUsersCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error
	// LoadAppsLocalCache(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) error
	GetAllForbiddenUsersFromDB(ctx context.Context) ([]*api.Users, error)
	// LoadUsersLocalCache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration) error
	Lock(ctx context.Context, key string, exp time.Duration) (DataLock, error)
	LastUpdated(ctx context.Context, key string) (time.Time, error)
}
type operationAction func(ctx context.Context) error
type DataOperatorUsecase struct {
	repo   DataOperatorRepo
	config *config.Config
	log    *logrus.Logger
}

func NewDataOperatorUsecase(repo DataOperatorRepo, config *config.Config, log *logrus.Logger) *DataOperatorUsecase {
	return &DataOperatorUsecase{repo: repo, config: config, log: log}
}

func (d *DataOperatorUsecase) Do() {
	d.LoadCache(context.Background())

}

func (d *DataOperatorUsecase) LoadCache(ctx context.Context) {
	errChan := make(chan error, 3)
	wg := &sync.WaitGroup{}
	actions := []operationAction{
		d.loadUsersBlacklist,
		d.loadApps,
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

	lockKey := d.config.GetUserBlackListLockKey()
	lock, err := d.repo.Lock(ctx, lockKey, time.Second*time.Duration(exp))
	if err != nil {
		return fmt.Errorf("lock error: %w", err)

	}

	if err = lock.Lock(ctx); err != nil && err != redislock.ErrNotObtained {
		return fmt.Errorf("lock error: %w", err)
	}
	defer lock.UnLock(ctx)
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

func (d *DataOperatorUsecase) Refresh(duratoin time.Duration) {
	ticker := time.NewTicker(duratoin)
	for range ticker.C {
		d.LoadCache(context.Background())
	}
}
