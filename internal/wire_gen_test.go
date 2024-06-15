package internal_test

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal"
	"github.com/begonia-org/begonia/internal/daemon"
	"github.com/begonia-org/begonia/internal/middleware"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/server"
	"github.com/begonia-org/begonia/internal/service"
	c "github.com/smartystreets/goconvey/convey"
)

func TestNewSrv(t *testing.T) {
	c.Convey("TestNewSrv", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv1 := internal.NewAPPSvr(cnf, gateway.Log)
		c.So(srv1, c.ShouldNotBeNil)
		srv2 := internal.NewAuthzSvr(cnf, gateway.Log)
		c.So(srv2, c.ShouldNotBeNil)
		srv3 := internal.NewEndpointSvr(cnf, gateway.Log)
		c.So(srv3, c.ShouldNotBeNil)
		srv4 := internal.NewFileSvr(cnf, gateway.Log)
		c.So(srv4, c.ShouldNotBeNil)
		srv5 := internal.NewSysSvr(cnf, gateway.Log)
		c.So(srv5, c.ShouldNotBeNil)
		srv6 := internal.InitOperatorApp(cnf)
		c.So(srv6, c.ShouldNotBeNil)
		srv7 := internal.NewWorker(cnf, gateway.Log, "127.0.0.1:12148")
		c.So(srv7, c.ShouldNotBeNil)
		patch := gomonkey.ApplyFunc((*daemon.DaemonImpl).Start, func(_ *daemon.DaemonImpl, _ context.Context) {})
		patch = patch.ApplyFunc((*gateway.GatewayServer).Start, func(_ *gateway.GatewayServer) {})
		patch = patch.ApplyFunc(server.NewGateway, func(_ *gateway.GatewayConfig, _ *cfg.Config, _ []service.Service, _ *middleware.PluginsApply) *gateway.GatewayServer {
			return &gateway.GatewayServer{}
		})
		defer patch.Reset()
		go srv7.Start()
		srv := internal.New(cnf, gateway.Log, "127.0.0.1:12138")
		c.So(srv, c.ShouldNotBeNil)

	})
}
