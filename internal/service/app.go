package service

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/web"
	"github.com/sirupsen/logrus"
)

type AppService struct {
	api.UnimplementedAppsServiceServer
	biz    *biz.AppUsecase
	log    *logrus.Logger
	config *config.Config
}

func (app *AppService) AddApps(ctx context.Context, in *api.AddAppsRequest) (*common.APIResponse, error) {
	err := app.biz.AddApps(ctx, in.Apps)
	if err != nil {
		app.log.Errorf("AddApps failed: %v", err)
		return web.MakeResponse(nil, err)

	}
	return web.MakeResponse(nil, nil)
}
func (app *AppService) GetApps(ctx context.Context, in *api.AppsListRequest) (*common.APIResponse, error) {
	apps, err := app.biz.GetApps(ctx, in.AccessKey)
	if err != nil {
		app.log.Errorf("GetApps failed: %v", err)
		return nil, err
	}
	if apps == nil {
		return web.MakeResponse(&api.AppsListResponse{
			Apps: nil,
		}, nil)
	}
	return web.MakeResponse(&api.AppsListResponse{Apps: apps}, nil)
}
