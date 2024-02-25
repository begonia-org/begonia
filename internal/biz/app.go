package biz

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
)

type AppRepo interface {
	AddApps(ctx context.Context, apps []*api.Apps) error
	GetApps(ctx context.Context, keys []string) ([]*api.Apps, error)
}

type AppUsecase struct {
	repo   AppRepo
	config *config.Config
}

func NewAppUsecase(repo AppRepo, config *config.Config) *AppUsecase {
	return &AppUsecase{repo: repo, config: config}
}
func (a *AppUsecase) AddApps(ctx context.Context, apps []*api.Apps) error {
	return a.repo.AddApps(ctx, apps)
}
func (a *AppUsecase) GetApps(ctx context.Context, keys []string) ([]*api.Apps, error) {
	return a.repo.GetApps(ctx, keys)
}