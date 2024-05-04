// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package internal

import (
	"context"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/biz/gateway"
	"github.com/begonia-org/begonia/internal/daemon"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/begonia-org/begonia/internal/pkg/migrate"
	"github.com/begonia-org/begonia/internal/server"
	"github.com/begonia-org/begonia/internal/service"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/spark-lence/tiga"
)

// Injectors from wire.go:

func InitOperatorApp(config2 *tiga.Configuration) *migrate.InitOperator {
	mySQLDao := data.NewMySQL(config2)
	v := migrate.NewTableModels()
	mySQLMigrate := migrate.NewMySQLMigrate(mySQLDao, v...)
	usersOperator := migrate.NewUsersOperator(mySQLDao)
	configConfig := config.NewConfig(config2)
	initOperator := migrate.NewInitOperator(mySQLMigrate, usersOperator, configConfig)
	return initOperator
}

func New(config2 *tiga.Configuration, log logger.Logger, endpoint string) GatewayWorker {
	configConfig := config.NewConfig(config2)
	mySQLDao := data.NewMySQL(config2)
	redisDao := data.NewRDB(config2)
	etcdDao := data.NewEtcd(config2)
	dataData := data.NewData(mySQLDao, redisDao, etcdDao)
	contextContext := context.Background()
	layeredCache := data.NewLayeredCache(contextContext, dataData, configConfig, log)
	appRepo := data.NewAppRepoImpl(dataData, layeredCache, configConfig)
	authzRepo := data.NewAuthzRepo(dataData, log, layeredCache)
	dataOperatorRepo := data.NewDataOperatorRepo(dataData, appRepo, authzRepo, layeredCache, log)
	endpointRepo := data.NewEndpointRepoImpl(dataData, configConfig)
	gatewayWatcher := gateway.NewWatcher(configConfig, endpointRepo)
	dataOperatorUsecase := biz.NewDataOperatorUsecase(dataOperatorRepo, configConfig, log, gatewayWatcher, endpointRepo)
	daemonDaemon := daemon.NewDaemonImpl(configConfig, dataOperatorUsecase)
	gatewayConfig := server.NewGatewayConfig(endpoint)
	fileRepo := data.NewFileRepoImpl(dataData)
	fileUsecase := file.NewFileUsecase(fileRepo, configConfig)
	fileService := service.NewFileService(fileUsecase, configConfig)
	usersAuth := crypto.NewUsersAuth()
	authzUsecase := biz.NewAuthzUsecase(authzRepo, log, usersAuth, configConfig)
	authzService := service.NewAuthzService(authzUsecase, log, usersAuth, configConfig)
	endpointUsecase := gateway.NewEndpointUsecase(endpointRepo, fileUsecase, configConfig)
	endpointsService := service.NewEndpointsService(endpointUsecase, log, configConfig)
	appUsecase := biz.NewAppUsecase(appRepo, configConfig)
	appService := service.NewAppService(appUsecase, log, configConfig)
	sysService := service.NewSysService()
	userRepo := data.NewUserRepoImpl(dataData, layeredCache, configConfig)
	userUsecase := biz.NewUserUsecase(userRepo, configConfig)
	userService := service.NewUserService(userUsecase, log, configConfig)
	v := service.NewServices(fileService, authzService, endpointsService, appService, sysService, userService)
	pluginsApply := middleware.New(configConfig, redisDao, authzUsecase, log, appRepo, layeredCache)
	gatewayServer := server.NewGateway(gatewayConfig, configConfig, v, pluginsApply)
	gatewayWorker := NewGatewayWorkerImpl(daemonDaemon, gatewayServer)
	return gatewayWorker
}
