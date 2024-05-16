package service

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
)

type AppService struct {
	api.UnimplementedAppsServiceServer
	biz    *biz.AppUsecase
	log    logger.Logger
	config *config.Config
}

func (app *AppService) Put(ctx context.Context, in *api.AppsRequest) (*api.AddAppResponse, error) {
	owner := GetIdentity(ctx)

	appInstance, err := app.biz.CreateApp(ctx, in, owner)
	if err != nil {
		// app.log.Errorf("CreateApp failed: %v", err)
		return nil, err
	}
	return &api.AddAppResponse{Appid: appInstance.Appid, AccessKey: appInstance.AccessKey, Secret: appInstance.Secret}, nil
}
func (app *AppService) Get(ctx context.Context, in *api.GetAPPRequest) (*api.Apps, error) {
	apps, err := app.biz.Get(ctx, in.Appid)
	if err != nil {
		app.log.Errorf(ctx,"GetApps failed: %v", err)
		return nil, err
	}
	return apps, nil
}

func (app *AppService) Desc() *grpc.ServiceDesc {
	return &api.AppsService_ServiceDesc
}

func NewAppService(biz *biz.AppUsecase, log logger.Logger, config *config.Config) *AppService {
	return &AppService{biz: biz, log: log, config: config}
}
func (app *AppService) Patch(ctx context.Context, in *api.AppsRequest) (*api.Apps, error) {
	owner := GetIdentity(ctx)
	appInstance, err := app.biz.Patch(ctx, in, owner)
	if err != nil {
		// app.log.Errorf("CreateApp failed: %v", err)
		return nil, err
	}
	return appInstance, nil
}
func (app *AppService) Delete(ctx context.Context, in *api.DeleteAppRequest) (*api.DeleteAppResponse, error) {
	err := app.biz.Del(ctx, in.Appid)
	if err != nil {
		return nil, err
	}
	return &api.DeleteAppResponse{}, nil
}
