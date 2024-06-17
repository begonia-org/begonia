package middleware_test

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/middleware"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

func TestMiddlewareUnaryInterceptorChains(t *testing.T) {
	c.Convey("test middleware unary interceptor chains", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		user := data.NewUserRepo(config, gateway.Log)
		userAuth := crypto.NewUsersAuth(cnf)
		authzRepo := data.NewAuthzRepo(config, gateway.Log)
		appRepo:=data.NewAppRepo(config,gateway.Log)
		authz := biz.NewAuthzUsecase(authzRepo, user,appRepo, gateway.Log, userAuth, cnf)
		repo := data.NewAppRepo(config, gateway.Log)

		akBiz := biz.NewAccessKeyAuth(repo, cnf, gateway.Log)
		mid := middleware.New(cnf, tiga.NewRedisDao(config), authz, gateway.Log, akBiz)
		// mid.SetPriority(1)
		c.So(len(mid.StreamInterceptorChains()), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(len(mid.UnaryInterceptorChains()), c.ShouldBeGreaterThanOrEqualTo, 0)

		plugins := cnf.GetPlugins()
		plugins["test"] = 1
		patch := gomonkey.ApplyFuncReturn((*cfg.Config).GetPlugins, plugins)
		defer patch.Reset()
		f := func() {
			middleware.New(cnf, tiga.NewRedisDao(config), authz, gateway.Log, akBiz)

		}
		c.So(f, c.ShouldPanicWith, "plugin test not found")
		patch.Reset()
		delete(plugins, "test")

		patch2 := gomonkey.ApplyFuncReturn((*cfg.Config).GetRPCPlugins, nil, fmt.Errorf("get rpc plugins error"))
		defer patch2.Reset()
		c.So(f, c.ShouldPanicWith, "get rpc plugins error:get rpc plugins error")
		patch2.Reset()
	})
}
