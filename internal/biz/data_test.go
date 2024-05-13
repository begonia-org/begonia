package biz_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/begonia-org/begonia/transport"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	appApi "github.com/begonia-org/go-sdk/api/app/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	c "github.com/smartystreets/goconvey/convey"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newDataOperatorUsecase() *biz.DataOperatorUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewEndpointRepo(config, logger.Log)
	repoData := data.NewOperator(config, logger.Log)
	cnf := cfg.NewConfig(config)
	watcher := endpoint.NewWatcher(cnf, repo)
	return biz.NewDataOperatorUsecase(repoData, cnf, logger.Log, watcher, repo)
}

func TestDo(t *testing.T) {
	dataOperator := newDataOperatorUsecase()
	userBiz := newUserBiz()
	appBiz := newAppBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	cache := data.NewLayered(config, logger.Log)
	_ = cache.Del(context.Background(), "begonia:user:black:lock")
	_ = cache.Del(context.Background(), "begonia:user:black:last_updated")
	opts := &transport.GrpcServerOptions{
		Middlewares:     make([]transport.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]loadbalance.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]gwRuntime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	gwCnf := &transport.GatewayConfig{
		GatewayAddr:   "127.0.0.1:9527",
		GrpcProxyAddr: "127.0.0.1:12148",
	}
	gateway.New(gwCnf, opts)
	c.Convey("test data operator do success", t, func() {
		u1 := &api.Users{
			Name:      fmt.Sprintf("user-data-operator-%s", time.Now().Format("20060102150405")),
			Dept:      "dev",
			Email:     fmt.Sprintf("user-op-1-biz%s@example.com", time.Now().Format("20060102150405")),
			Phone:     time.Now().Format("20060102150405"),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_LOCKED,
		}
		u2 := &api.Users{
			Name:      fmt.Sprintf("user-data-operator-2-%s", time.Now().Format("20060102150405")),
			Dept:      "dev",
			Email:     fmt.Sprintf("user-op-2-biz%s@example.com", time.Now().Format("20060102150405")),
			Phone:     fmt.Sprintf("2%s", time.Now().Format("20060102150405")),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		}
		err := userBiz.Add(context.Background(), u1)
		c.So(err, c.ShouldBeNil)
		err = userBiz.Add(context.Background(), u2)
		c.So(err, c.ShouldBeNil)
		app := &appApi.Apps{
			Name:        fmt.Sprintf("app-data-operator-%s", time.Now().Format("20060102150405")),
			Owner:       "test-user-01",
			Description: "dev",
			Tags:        []string{"dev"},
			Status:      appApi.APPStatus_APP_ENABLED,
		}
		err = appBiz.Put(context.TODO(), app, u2.Uid)
		c.So(err, c.ShouldBeNil)
		go dataOperator.Do(context.Background())
		go dataOperator.Do(context.Background())

		time.Sleep(3 * time.Second)
		prefix := cnf.GetUserBlackListPrefix()
		val, err := cache.GetFromLocal(context.TODO(), fmt.Sprintf("%s:%s", prefix, u1.Uid))
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldNotBeEmpty)
		appPrefix := cnf.GetAppPrefix()
		val, err = cache.GetFromLocal(context.TODO(), fmt.Sprintf("%s:access_key:%s", appPrefix, app.AccessKey))
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldNotBeEmpty)
	})

}
