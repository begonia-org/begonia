package biz

import (
	"context"
	"fmt"
	"time"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AppRepo interface {
	AddApps(ctx context.Context, apps []*api.Apps) (*gorm.DB, error)
	GetApps(ctx context.Context, keys []string) ([]*api.Apps, error)
	CacheApps(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) redis.Pipeliner
	DelAppsCache(ctx context.Context, prefix string, models []*api.Apps) error
	ListApps(ctx context.Context, conds ...interface{}) ([]*api.Apps, error) 
}

type AppUsecase struct {
	repo   AppRepo
	config *config.Config
}

func NewAppUsecase(repo AppRepo, config *config.Config) *AppUsecase {
	return &AppUsecase{repo: repo, config: config}
}

// AddApps 新增并缓存app
func (a *AppUsecase) AddApps(ctx context.Context, apps []*api.Apps) error {
	db, err := a.repo.AddApps(ctx, apps)
	if err != nil {
		return err

	}
	prefix := a.config.GetAPPAccessKeyPrefix()
	pipe := a.repo.CacheApps(ctx, prefix, apps, 0)
	if _, err := pipe.Exec(ctx); err != nil {
		db.Rollback()
		return fmt.Errorf("cache apps failed: %w", err)
	}
	err = db.Commit().Error
	if err != nil {
		_=a.repo.DelAppsCache(ctx, prefix, apps)
		return fmt.Errorf("commit db failed: %w", err)

	}
	return nil
	// return a.repo.AddApps(ctx, apps)
}
func (a *AppUsecase) GetApps(ctx context.Context, keys []string) ([]*api.Apps, error) {
	return a.repo.GetApps(ctx, keys)
}

func (a *AppUsecase) CacheApps(ctx context.Context, prefix string, models []*api.Apps, exp time.Duration) redis.Pipeliner {
	return a.repo.CacheApps(ctx, prefix, models, exp)

}
func (a *AppUsecase) DelAppsCache(ctx context.Context, prefix string, models []*api.Apps) error {
	return a.repo.DelAppsCache(ctx, prefix, models)
}
