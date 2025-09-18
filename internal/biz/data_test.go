package biz_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	appApi "github.com/begonia-org/go-sdk/api/app/v1"
	ep "github.com/begonia-org/go-sdk/api/endpoint/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	c "github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/spark-lence/tiga"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// func TestCreateInBatches(t *testing.T) {
func TestMain(m *testing.M) {
	log.Printf("Start testing")
	// setup()
	code := m.Run()
	log.Printf("All tests passed with code %d", code)
	// setup()
}

func setup() {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := config.ReadConfig(env)

	// cnf:=config.NewConfig(conf)
	rdb := tiga.NewRedisDao(conf)
	luaScript := `
		local prefix = KEYS[1]
		local cursor = "0"
		local count = 100
		repeat
			local result = redis.call("SCAN", cursor, "MATCH", prefix, "COUNT", count)
			cursor = result[1]
			local keys = result[2]
			if #keys > 0 then
				redis.call("DEL", unpack(keys))
			end
		until cursor == "0"
		return "OK"
		`

	_, err := rdb.GetClient().Eval(context.Background(), luaScript, []string{"test:*"}).Result()
	if err != nil {
		log.Fatalf("Could not execute Lua script: %v", err)
	}
	etcd := tiga.NewEtcdDao(conf)
	// 设置前缀
	prefix := "/test"

	// 使用前缀删除所有键
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = etcd.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("Failed to delete keys with prefix %s: %v", prefix, err)
	}
	mysql := tiga.NewMySQLDao(conf)
	mysql.RegisterTimeSerializer()
	err = mysql.GetModel(&user.Users{}).Where("`group` = ?", "test-user-01").Delete(&user.Users{}).Error
	if err != nil {
		log.Fatalf("Failed to delete keys with prefix %s: %v", prefix, err)
	}
	log.Printf("Cleaned up test data")
}

func newDataOperatorUsecase() *biz.DataOperatorUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewEndpointRepo(config, gateway.Log)
	repoData := data.NewOperator(config, gateway.Log)
	cnf := cfg.NewConfig(config)
	watcher := endpoint.NewWatcher(cnf, repo)
	return biz.NewDataOperatorUsecase(repoData, cnf, gateway.Log, watcher, repo)
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
	cache := data.NewLayered(config, gateway.Log)
	_ = cache.Del(context.Background(), "begonia:user:black:lock")
	_ = cache.Del(context.Background(), "begonia:user:black:last_updated")
	opts := &gateway.GrpcServerOptions{
		Middlewares:     make([]gateway.GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]loadbalance.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]gwRuntime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	gwCnf := &gateway.GatewayConfig{
		GatewayAddr:   "127.0.0.1:1949",
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
		patch := gomonkey.ApplyFuncReturn((*cfg.Config).GetUserBlackListExpiration, 3)
		defer patch.Reset()
		go dataOperator.Do(context.Background())
		go dataOperator.Do(context.Background())

		time.Sleep(5 * time.Second)
		prefix := cnf.GetUserBlackListPrefix()
		val, err := cache.GetFromLocal(context.TODO(), fmt.Sprintf("%s:%s", prefix, u1.Uid))
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldNotBeEmpty)
		appPrefix := cnf.GetAppPrefix()
		val, err = cache.GetFromLocal(context.TODO(), fmt.Sprintf("%s:access_key:%s", appPrefix, app.AccessKey))
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldNotBeEmpty)

		patch1 := gomonkey.ApplyFuncReturn((*biz.DataOperatorUsecase).OnStart, fmt.Errorf("test data operator on start error"))
		defer patch1.Reset()
		repo := data.NewOperator(config, gateway.Log)
		patch2 := gomonkey.ApplyMethodReturn(repo, "Watcher", fmt.Errorf("test data watcher list error"))
		defer patch2.Reset()
		ctx, cancel := context.WithCancel(context.Background())
		go dataOperator.Do(ctx)
		time.Sleep(1 * time.Second)
		patch1.Reset()
		patch2.Reset()
		cancel()
		ctx, cancel = context.WithCancel(context.Background())
		patch3 := gomonkey.ApplyFuncReturn((*cfg.Config).GetUserBlackListExpiration, 0)
		defer patch3.Reset()
		go dataOperator.Handle(ctx)
		time.Sleep(1 * time.Second)
		patch3.Reset()
		cancel()
		// patch2 := gomonkey.ApplyFuncReturn((*biz.DataOperatorUsecase).Refresh, fmt.Errorf("test data operator refresh error"))
	})

}

func TestHandleError(t *testing.T) {
	dataOperator := newDataOperatorUsecase()
	c.Convey("test data operator handle error", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		repo := data.NewOperator(config, gateway.Log)
		patch := gomonkey.ApplyMethodReturn(repo, "Locker", nil, fmt.Errorf("test new lock error"))
		defer patch.Reset()
		ctx, cancel := context.WithCancel(context.Background())
		go dataOperator.Handle(ctx)
		time.Sleep(1 * time.Second)
		patch.Reset()
		cancel()
		locker := data.NewLocker(config, gateway.Log, "test-test-test", 3*time.Second, 0)
		patch2 := gomonkey.ApplyMethodReturn(locker, "Lock", fmt.Errorf("test lock error"))
		defer patch2.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		go dataOperator.Handle(ctx)
		time.Sleep(1 * time.Second)
		patch2.Reset()
		cancel()

		patch3 := gomonkey.ApplyMethodReturn(locker, "UnLock", fmt.Errorf("test unlock error"))
		defer patch3.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		go dataOperator.Handle(ctx)
		time.Sleep(1 * time.Second)
		patch3.Reset()
		cancel()

		patch4 := gomonkey.ApplyMethodReturn(repo, "GetAllForbiddenUsers", nil, fmt.Errorf("test latest GetAllForbiddenUsers error"))
		patch4 = patch4.ApplyMethodReturn(repo, "LastUpdated", time.Time{}, nil)
		defer patch4.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		go dataOperator.Handle(ctx)
		time.Sleep(1 * time.Second)
		patch4.Reset()
		cancel()

		patch5 := gomonkey.ApplyMethodReturn(repo, "FlashUsersCache", fmt.Errorf("test latest FlashUsersCache error"))
		patch5 = patch5.ApplyMethodReturn(repo, "LastUpdated", time.Time{}, nil)
		defer patch5.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		go dataOperator.Handle(ctx)
		time.Sleep(1 * time.Second)
		patch5.Reset()
		cancel()

	})
}

func TestOnStart(t *testing.T) {
	dataOperator := newDataOperatorUsecase()
	c.Convey("test data operator on start", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		repo := data.NewEndpointRepo(config, gateway.Log)
		patch := gomonkey.ApplyMethodReturn(repo, "List", nil, fmt.Errorf("test endpoint list error"))
		defer patch.Reset()
		ctx, cancel := context.WithCancel(context.Background())
		err := dataOperator.OnStart(ctx)
		time.Sleep(4 * time.Second)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test endpoint list error")
		patch.Reset()
		cancel()
		// []*api.Endpoints, error
		patch2 := gomonkey.ApplyMethodReturn(repo, "List", []*ep.Endpoints{{
			Name:        "test",
			Owner:       "test",
			Description: "test",
		}}, nil)
		patch2 = patch2.ApplyFuncReturn((*endpoint.EndpointWatcher).Update, fmt.Errorf("test endpoint watcher update error"))

		defer patch2.Reset()
		ctx, cancel = context.WithCancel(context.Background())
		err = dataOperator.OnStart(ctx)

		time.Sleep(2 * time.Second)

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test endpoint watcher update error")
		patch2.Reset()
		cancel()

	})
}

func TestRefresh(t *testing.T) {
	c.Convey("test data operator refresh", t, func() {
		dataOperator := newDataOperatorUsecase()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go dataOperator.Refresh(ctx, 1*time.Second)
		time.Sleep(3 * time.Second)
	})
}
