// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package service

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	v1_2 "github.com/begonia-org/go-sdk/api/app/v1"
	v1_3 "github.com/begonia-org/go-sdk/api/endpoint/v1"
	v1_4 "github.com/begonia-org/go-sdk/api/file/v1"
	v1_5 "github.com/begonia-org/go-sdk/api/sys/v1"
	"github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/spark-lence/tiga"
)

// Injectors from wire.go:

func NewAuthzSvrForTest(config2 *tiga.Configuration, log logger.Logger) v1.AuthServiceServer {
	redisDao := data.NewRDB(config2)
	configConfig := config.NewConfig(config2)
	layeredCache := data.NewLayeredCache(redisDao, configConfig, log)
	authzRepo := data.NewAuthzRepoImpl(log, layeredCache)
	mySQLDao := data.NewMySQL(config2)
	etcdDao := data.NewEtcd(config2)
	dataData := data.NewData(mySQLDao, redisDao, etcdDao)
	curd := data.NewCurdImpl(mySQLDao, configConfig)
	userRepo := data.NewUserRepoImpl(dataData, layeredCache, curd, configConfig)
	usersAuth := crypto.NewUsersAuth()
	authzUsecase := biz.NewAuthzUsecase(authzRepo, userRepo, log, usersAuth, configConfig)
	authServiceServer := NewAuthzService(authzUsecase, log, usersAuth, configConfig)
	return authServiceServer
}

func NewAPPSvrForTest(config2 *tiga.Configuration, log logger.Logger) v1_2.AppsServiceServer {
	mySQLDao := data.NewMySQL(config2)
	configConfig := config.NewConfig(config2)
	curd := data.NewCurdImpl(mySQLDao, configConfig)
	redisDao := data.NewRDB(config2)
	layeredCache := data.NewLayeredCache(redisDao, configConfig, log)
	appRepo := data.NewAppRepoImpl(curd, layeredCache, configConfig)
	appUsecase := biz.NewAppUsecase(appRepo, configConfig)
	appsServiceServer := NewAppService(appUsecase, log, configConfig)
	return appsServiceServer
}

func NewEndpointSvrForTest(config2 *tiga.Configuration, log logger.Logger) v1_3.EndpointServiceServer {
	mySQLDao := data.NewMySQL(config2)
	redisDao := data.NewRDB(config2)
	etcdDao := data.NewEtcd(config2)
	dataData := data.NewData(mySQLDao, redisDao, etcdDao)
	configConfig := config.NewConfig(config2)
	endpointRepo := data.NewEndpointRepoImpl(dataData, configConfig)
	fileUsecase := file.NewFileUsecase(configConfig)
	endpointUsecase := endpoint.NewEndpointUsecase(endpointRepo, fileUsecase, configConfig)
	endpointServiceServer := NewEndpointsService(endpointUsecase, log, configConfig)
	return endpointServiceServer
}

func NewFileSvrForTest(config2 *tiga.Configuration, log logger.Logger) v1_4.FileServiceServer {
	configConfig := config.NewConfig(config2)
	fileUsecase := file.NewFileUsecase(configConfig)
	fileServiceServer := NewFileService(fileUsecase, configConfig)
	return fileServiceServer
}

func NewSysSvrForTest(config2 *tiga.Configuration, log logger.Logger) v1_5.SystemServiceServer {
	systemServiceServer := NewSysService()
	return systemServiceServer
}

func NewUserSvrForTest(config2 *tiga.Configuration, log logger.Logger) v1.UserServiceServer {
	mySQLDao := data.NewMySQL(config2)
	redisDao := data.NewRDB(config2)
	etcdDao := data.NewEtcd(config2)
	dataData := data.NewData(mySQLDao, redisDao, etcdDao)
	configConfig := config.NewConfig(config2)
	layeredCache := data.NewLayeredCache(redisDao, configConfig, log)
	curd := data.NewCurdImpl(mySQLDao, configConfig)
	userRepo := data.NewUserRepoImpl(dataData, layeredCache, curd, configConfig)
	userUsecase := biz.NewUserUsecase(userRepo, configConfig)
	userServiceServer := NewUserService(userUsecase, log, configConfig)
	return userServiceServer
}
