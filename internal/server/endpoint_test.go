package server

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	api "github.com/begonia-org/begonia/api/v1"
	cfg "github.com/begonia-org/begonia/config"
	c "github.com/smartystreets/goconvey/convey"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/runtime"
) // 别名导入

func TestEndpointWatch(t *testing.T) {
	c.Convey("test endpoint watch", t, func() {
		conf := config.NewConfig(cfg.ReadConfig("dev"))
		NewGatewayMux(conf)
		e := NewEndpointManagerImpl(biz.NewEndpointUsecase(nil), conf)
		ctx, cancel := context.WithCancel(context.Background())
		patch := gomonkey.ApplyFunc((*biz.EndpointUsecase).GetEndpoint, func(_ *biz.EndpointUsecase, _ context.Context, _ string) (*api.Endpoints, error) {
			return &api.Endpoints{
				PluginId: "github.com.begonia-org.example",
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
		errChan := make(chan error, 10)
		go func() {
			err := e.Watch(ctx, conf.GetPluginDir(), errChan)
			c.So(err, c.ShouldBeNil)
		}()
		go func(errChan <-chan error) {
			for err := range errChan {
				t.Log(err)
				c.So(err, c.ShouldBeNil)
			}
		}(errChan)
		load := runtime.NewProtoLoaderImpl(conf)
		err := load.LoadProto("../../example/protos.zip", "github.com/begonia-org/example2", "./api/v1", "example2")
		t.Log(err)
		c.So(err, c.ShouldBeNil)
		time.Sleep(time.Second * 2)
		cancel()
		// defer os.RemoveAll(filepath.Join(conf.GetPluginDir(), "example"))
	})
}
