package server

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	c "github.com/smartystreets/goconvey/convey"
	api "github.com/wetrycode/begonia/api/v1"
	cfg "github.com/wetrycode/begonia/config"

	"github.com/wetrycode/begonia/internal/biz"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/runtime"
) // 别名导入

func TestEndpointWatch(t *testing.T) {
	c.Convey("test endpoint watch", t, func() {
		conf := config.NewConfig(cfg.ReadConfig("dev"))
		NewGatewayMux(nil, conf)
		e := NewEndpointManagerImpl(biz.NewEndpointUsecase(nil), conf)
		ctx, cancel := context.WithCancel(context.Background())
		patch := gomonkey.ApplyFunc((*biz.EndpointUsecase).GetEndpoint, func(_ *biz.EndpointUsecase, _ context.Context, _ string) (*api.Endpoints, error) {
			return &api.Endpoints{
				PluginId: "github.com.wetrycode.example",
				Endpoint: "127.0.0.1:51001",
			}, nil
		})
		defer patch.Reset()

		// patch2 := gomonkey.ApplyFunc((*EndpointManagerImpl).addEndpoints, func(imp *EndpointManagerImpl,reg endpoint.EndpointRegister, ep string) (error) {
		// 	err:=imp.addEndpoints(reg,ep)
		// 	c.So(err,c.ShouldBeNil)
		// 	return err
		// })
		// defer patch2.Reset()

		// patch3 := gomonkey.ApplyFunc((*EndpointManagerImpl).createEndpointRegister, func(imp *EndpointManagerImpl,pluginPath string) (endpoint.EndpointRegister, error) {
		// 	endpointRegister,err:=imp.createEndpointRegister(pluginPath)
		// 	c.So(err,c.ShouldBeNil)

		// 	return endpointRegister,err
		// })
		// defer patch3.Reset()
		go e.Watch(ctx, conf.GetPluginDir())
		load := runtime.NewProtoLoaderImpl(conf)
		err := load.LoadProto("../../example/protos.zip", "github.com/wetrycode/example", "./api/v1", "example")
		t.Log(err)
		c.So(err, c.ShouldBeNil)
		time.Sleep(time.Second * 2)
		cancel()
		// defer os.RemoveAll(filepath.Join(conf.GetPluginDir(), "example"))
	})
}
