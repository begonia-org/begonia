package data

import (
	"context"
	"fmt"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/spark-lence/tiga"
)

type appRepoImpl struct {
	local *LayeredCache
	cfg   *config.Config
	curd  biz.CURD
}

func NewAppRepoImpl(curd biz.CURD, local *LayeredCache, cfg *config.Config) biz.AppRepo {
	return &appRepoImpl{curd: curd, local: local, cfg: cfg}
}

func (r *appRepoImpl) Add(ctx context.Context, apps *api.Apps) error {

	if err := r.curd.Add(ctx, apps, false); err != nil {
		return fmt.Errorf("add app failed: %w", err)
	}
	key := r.cfg.GetAPPAccessKey(apps.AccessKey)
	exp := r.cfg.GetAPPAccessKeyExpiration()
	err := r.local.Set(ctx, key, []byte(apps.Secret), time.Duration(exp)*time.Second)
	return err
}
func (r *appRepoImpl) Get(ctx context.Context, key string) (*api.Apps, error) {

	app := &api.Apps{}
	err := r.curd.Get(ctx, app, false, "access_key = ? or appid=?", key, key)
	if err != nil || app.Appid == "" {
		return nil, fmt.Errorf("get app failed: %w", err)
	}
	return app, err
}

func (r *appRepoImpl) Cache(ctx context.Context, prefix string, models *api.Apps, exp time.Duration) error {

	key := fmt.Sprintf("%s:%s", prefix, models.AccessKey)
	return r.local.Set(ctx, key, []byte(models.Secret), exp)
}
func (r *appRepoImpl) Del(ctx context.Context, key string) error {
	app, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	_ = r.local.Del(ctx, r.cfg.GetAPPAccessKey(app.AccessKey))
	return r.curd.Del(ctx, app, false)
}
func (r *appRepoImpl) Patch(ctx context.Context, model *api.Apps) error {

	return r.curd.Update(ctx, model, false)
}
func (r *appRepoImpl) List(ctx context.Context, tags []string, status []api.APPStatus, page, pageSize int32) ([]*api.Apps, error) {
	apps := make([]*api.Apps, 0)
	query := ""
	conds := make([]interface{}, 0)
	if len(tags) > 0 {
		query = "tags in (?)"
		conds = append(conds, tags)
	}
	if len(status) > 0 {
		if query != "" {
			query += " and "
		}
		query += "status in (?)"
		conds = append(conds, status)
	}
	pagination := &tiga.Pagination{
		Page:     page,
		PageSize: pageSize,
		Query:    query,
		Args:     conds,
	}
	err := r.curd.List(ctx, &apps, pagination)
	if err != nil {
		return nil, fmt.Errorf("list app failed: %w", err)

	}
	return apps, nil

}

func (a *appRepoImpl) GetSecret(ctx context.Context, accessKey string) (string, error) {
	cacheKey := a.cfg.GetAPPAccessKey(accessKey)
	secretBytes, err := a.local.Get(ctx, cacheKey)
	secret := string(secretBytes)
	if err != nil {
		apps, err := a.Get(ctx, accessKey)
		if err != nil {
			return "", err
		}
		secret = apps.Secret

		// _ = a.rdb.Set(ctx, cacheKey, secret, time.Hour*24*3)
		_ = a.local.Set(ctx, cacheKey, []byte(secret), time.Hour*24*3)
	}
	return secret, nil
}
func (a *appRepoImpl) GetAppid(ctx context.Context, accessKey string) (string, error) {
	cacheKey := a.cfg.GetAppidKey(accessKey)
	secretBytes, err := a.local.Get(ctx, cacheKey)
	appid := string(secretBytes)
	if err != nil {
		apps, err := a.Get(ctx, accessKey)
		if err != nil {
			return "", err
		}
		appid = apps.Appid

		// _ = a.rdb.Set(ctx, cacheKey, secret, time.Hour*24*3)
		_ = a.local.Set(ctx, cacheKey, []byte(appid), time.Hour*24*3)
	}
	return appid, nil
}
