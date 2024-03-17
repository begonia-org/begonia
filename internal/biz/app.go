package biz

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
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
	repo      AppRepo
	config    *config.Config
	snowflake *tiga.Snowflake
}

func NewAppUsecase(repo AppRepo, config *config.Config) *AppUsecase {
	sn, _ := tiga.NewSnowflake(1)
	return &AppUsecase{repo: repo, config: config, snowflake: sn}
}
func (a *AppUsecase) generateRandomString(n int) (string, error) {
	const lettersAndDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("Failed to generate random string: %w", err)
	}

	for i := 0; i < n; i++ {
		// 将随机字节转换为lettersAndDigits中的一个有效字符
		b[i] = lettersAndDigits[b[i]%byte(len(lettersAndDigits))]
	}

	return string(b), nil
}
func (a *AppUsecase) newApp() *api.Apps {
	return &api.Apps{
		Status:    api.APPStatus_APP_ENABLED,
		IsDeleted: false,
	}

}
func (a *AppUsecase) CreateApp(ctx context.Context, in *api.CreateAppRequest) (*api.Apps, error) {
	// return a.repo.ListApps(ctx, conds...)
	appid := a.generateAppid(ctx)
	accessKey, err := a.generateAppAccessKey(ctx)
	if err != nil {
		return nil, errors.New(err, int32(api.APPSvrCode_APP_CREATE_ERR), codes.Internal, "generate_app_access_key")

	}
	secret, err := a.generateAppSecret(ctx)
	if err != nil {
		return nil, errors.New(err, int32(api.APPSvrCode_APP_CREATE_ERR), codes.Internal, "generate_app_secret_key")
	}
	app := a.newApp()
	app.Key = accessKey
	app.Secret = secret
	app.Appid = appid
	app.Name = in.Name
	app.Description = in.Description
	app.Tags = in.Tags
	app.Owner = in.Owner
	err = a.AddApps(ctx, []*api.Apps{app})
	if err != nil {
		return nil, err

	}
	return app, nil
}
func (a *AppUsecase) generateAppid(_ context.Context) string {
	appid := a.snowflake.GenerateIDString()
	return appid
}
func (a *AppUsecase) generateAppAccessKey(_ context.Context) (string, error) {
	return a.generateRandomString(32)
}
func (a *AppUsecase) generateAppSecret(_ context.Context) (string, error) {
	return a.generateRandomString(64)
}

// AddApps 新增并缓存app
func (a *AppUsecase) AddApps(ctx context.Context, apps []*api.Apps) (err error) {
	defer func() {
		if err != nil {
			// log.Println(err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = errors.New(err, int32(api.APPSvrCode_APP_DUPLICATE_ERR), codes.AlreadyExists, "commit_app")
			} else {
				err = errors.New(err, int32(api.APPSvrCode_APP_CREATE_ERR), codes.Internal, "cache_apps")

			}
		}
	}()
	db, err := a.repo.AddApps(ctx, apps)
	if err != nil {

		return err
	}
	prefix := a.config.GetAPPAccessKeyPrefix()
	pipe := a.repo.CacheApps(ctx, prefix, apps, time.Duration(0)*time.Second)
	if _, err = pipe.Exec(ctx); err != nil {
		db.Rollback()
		return err
	}
	err = db.Commit().Error
	if err != nil {
		_ = a.repo.DelAppsCache(ctx, prefix, apps)

		return err

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
