// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware/validator"
	"github.com/begonia-org/begonia/internal/service"
	"github.com/begonia-org/dynamic-proto"
	"github.com/begonia-org/go-layered-bloom"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
)

// Injectors from wire.go:

func New(config2 *tiga.Configuration, log *logrus.Logger, endpoint string, name golayeredbloom.ConsumerName) *dynamicproto.GatewayServer {
	gatewayConfig := NewGatewayConfig(endpoint)
	configConfig := config.NewConfig(config2)
	mySQLDao := data.NewMySQL(config2)
	redisDao := data.NewRDB(config2)
	dataData := data.NewData(mySQLDao, redisDao)
	fileRepo := data.NewFileRepoImpl(dataData)
	fileUsecase := biz.NewFileUsecase(fileRepo, configConfig)
	fileService := service.NewFileService(fileUsecase, configConfig)
	client := data.GetRDBClient(redisDao)
	groupName := config.GetBlacklistPubSubGroup(configConfig)
	bloomPubSub := golayeredbloom.NewBloomPubSub(client, groupName, name, log)
	channelName := config.GetBlacklistPubSubChannel(configConfig)
	layeredBloomFilter := golayeredbloom.NewLayeredBloomFilter(bloomPubSub, channelName, name)
	localCache := data.NewLocalCache(dataData, configConfig, log, layeredBloomFilter)
	usersRepo := data.NewUserRepo(dataData, log, localCache)
	usersAuth := crypto.NewUsersAuth()
	usersUsecase := biz.NewUsersUsecase(usersRepo, log, usersAuth, configConfig)
	usersService := service.NewUserService(usersUsecase, log, usersAuth, configConfig)
	endpointRepo := data.NewEndpointRepoImpl(dataData)
	endpointUsecase := biz.NewEndpointUsecase(endpointRepo)
	endpointsService := service.NewEndpointsService(endpointUsecase, log, configConfig)
	v := service.NewServices(fileService, usersService, endpointsService)
	apiValidator := validator.NewAPIValidator(redisDao, log, usersUsecase, configConfig, mySQLDao, localCache)
	gatewayServer := NewGateway(gatewayConfig, configConfig, v, apiValidator)
	return gatewayServer
}
