package endpoint_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/transport"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var epId = ""

func newEndpointBiz() *endpoint.EndpointUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := config.ReadConfig(env)
	cnf := cfg.NewConfig(conf)
	repo := data.NewEndpointRepo(conf, logger.Log)
	return endpoint.NewEndpointUsecase(repo, nil, cnf)
}

func testAddEndpoint(t *testing.T) {
	endpointBiz := newEndpointBiz()
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "integration", "testdata", "helloworld.pb")
	pb, err := os.ReadFile(pbFile)
	c.Convey("Test Add Endpoint", t, func() {

		c.So(err, c.ShouldBeNil)
		endpointSvr := &api.EndpointSrvConfig{
			DescriptorSet: pb,
			Name:          "test",
			ServiceName:   "test",
			Description:   "test",
			Balance:       string(goloadbalancer.RRBalanceType),
			Endpoints: []*api.EndpointMeta{
				{
					Addr:   "127.0.0.1:21213",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21214",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21215",
					Weight: 0,
				},
			},
			Tags: []string{"test-biz=0"},
		}
		endpointId, err := endpointBiz.AddConfig(context.TODO(), endpointSvr)
		c.So(err, c.ShouldBeNil)
		c.So(endpointId, c.ShouldNotBeEmpty)
		epId = endpointId
	})
	c.Convey("Test Add Endpoint Fail", t, func() {
		endpointSvr := &api.EndpointSrvConfig{
			DescriptorSet: pb,
			Name:          "test",
			ServiceName:   "test",
			Description:   "test",
			Balance:       "test",
			Endpoints: []*api.EndpointMeta{
				{
					Addr:   "127.0.0.1:21213",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21214",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21215",
					Weight: 0,
				},
			},
			Tags: []string{"test-biz=0"},
		}
		_, err := endpointBiz.AddConfig(context.TODO(), endpointSvr)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrUnknownLoadBalancer.Error())

		patch := gomonkey.ApplyFuncReturn((*data.Data).PutEtcdWithTxn, false, fmt.Errorf("test error"))
		defer patch.Reset()
		endpointSvr.Balance = string(goloadbalancer.RRBalanceType)
		_, err = endpointBiz.AddConfig(context.TODO(), endpointSvr)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test error")
		patch.Reset()
	})

}

func testGetEndpoint(t *testing.T) {
	endpointBiz := newEndpointBiz()

	c.Convey("Test Get Endpoint Fail", t, func() {

		patch := gomonkey.ApplyFuncReturn((tiga.EtcdDao).GetString, "{false", nil)
		defer patch.Reset()
		val, err := endpointBiz.Get(context.TODO(), epId)
		c.So(val, c.ShouldBeNil)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "invalid character")
		patch.Reset()

		_, err = endpointBiz.Get(context.TODO(), "test")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrEndpointNotExists.Error())
	})
	c.Convey("Test Get Endpoint", t, func() {
		data, err := endpointBiz.Get(context.TODO(), epId)
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldNotBeEmpty)
	})
}

func testPatchEndpoint(t *testing.T) {
	endpointBiz := newEndpointBiz()
	c.Convey("Test Patch Endpoint", t, func() {
		updated_at, err := endpointBiz.Patch(context.TODO(), &api.EndpointSrvUpdateRequest{
			UniqueKey:   epId,
			Name:        "test_patch",
			Description: "test patch",
			Tags:        []string{"test-biz-1"},
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{"description", "tags"}},
		})
		c.So(err, c.ShouldBeNil)
		c.So(updated_at, c.ShouldNotBeEmpty)

		data, err := endpointBiz.Get(context.TODO(), epId)
		c.So(err, c.ShouldBeNil)
		c.So(data.Tags, c.ShouldContain, "test-biz-1")
		c.So(data.Description, c.ShouldEqual, "test patch")
		c.So(data.UpdatedAt, c.ShouldEqual, updated_at)
		c.So(data.Name, c.ShouldEqual, "test")

	})
	c.Convey("Test Patch Endpoint Fail", t, func() {
		_, err := endpointBiz.Patch(context.TODO(), &api.EndpointSrvUpdateRequest{
			UniqueKey:   "test",
			Name:        "test_patch",
			Description: "test patch",
			Tags:        []string{"test-biz-1"},
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{"description", "tags"}},
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrEndpointNotExists.Error())

		patch := gomonkey.ApplyFuncReturn((tiga.EtcdDao).GetString, "{false", nil)
		defer patch.Reset()
		_, err = endpointBiz.Patch(context.TODO(), &api.EndpointSrvUpdateRequest{
			UniqueKey:   epId,
			Name:        "test_patch",
			Description: "test patch",
			Tags:        []string{"test-biz-1"},
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{"description", "tags"}},
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "invalid character")
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn(json.Marshal, nil, fmt.Errorf("test marshal error"))
		defer patch2.Reset()
		_, err = endpointBiz.Patch(context.TODO(), &api.EndpointSrvUpdateRequest{
			UniqueKey:  epId,
			Name:       "test_patch2",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test marshal error")
		patch2.Reset()

		patch3 := gomonkey.ApplyFuncReturn(json.Unmarshal, fmt.Errorf("test Unmarshal error"))

		defer patch3.Reset()
		_, err = endpointBiz.Patch(context.TODO(), &api.EndpointSrvUpdateRequest{
			UniqueKey:  epId,
			Name:       "test_patch3",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test Unmarshal error")
		patch3.Reset()
	})
}

func testListEndpoints(t *testing.T) {
	endpointBiz := newEndpointBiz()
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "integration", "testdata", "helloworld.pb")
	pb, err := os.ReadFile(pbFile)
	if err != nil {
		t.Error(err)
	}
	c.Convey("Test List Endpoint", t, func() {

		rander := rand.New(rand.NewSource(time.Now().UnixNano()))

		tags := []string{
			fmt.Sprintf("test-biz-%d-%s", 0, time.Now().Format("20060102150405")),
			fmt.Sprintf("test-biz-%d-%s", 1, time.Now().Format("20060102150405")),
			fmt.Sprintf("test-biz-%d-%s", 2, time.Now().Format("20060102150405")),
			fmt.Sprintf("test-biz-%d-%s", 3, time.Now().Format("20060102150405")),
			fmt.Sprintf("test-biz-%d-%s", 4, time.Now().Format("20060102150405")),
		}
		eps := make([]string, 0)
		for i := 0; i < 10; i++ {

			id, err := endpointBiz.AddConfig(context.TODO(), &api.EndpointSrvConfig{
				DescriptorSet: pb,
				Name:          fmt.Sprintf("test-%d", i),
				ServiceName:   fmt.Sprintf("test-%d", i),
				Description:   fmt.Sprintf("test-%d", i),
				Balance:       string(goloadbalancer.RRBalanceType),
				Endpoints: []*api.EndpointMeta{
					{
						Addr: ""}},
				Tags: []string{
					tags[0:2][rander.Intn(2)],
					tags[2:4][rander.Intn(2)],
				}})
			c.So(err, c.ShouldBeNil)
			c.So(id, c.ShouldNotBeEmpty)
			eps = append(eps, id)
		}
		data, err := endpointBiz.List(context.TODO(), &api.ListEndpointRequest{
			Tags:       []string{tags[0], tags[1], tags[2]},
			UniqueKeys: []string{eps[0], eps[1], eps[2]},
		})
		c.So(err, c.ShouldBeNil)
		c.So(len(data), c.ShouldBeGreaterThan, 0)

		data, err = endpointBiz.List(context.TODO(), &api.ListEndpointRequest{
			Tags: []string{tags[0], tags[2]},
		})
		c.So(err, c.ShouldBeNil)
		c.So(len(data), c.ShouldBeGreaterThan, 0)

		data, err = endpointBiz.List(context.TODO(), &api.ListEndpointRequest{
			UniqueKeys: []string{eps[0], eps[2]},
		})
		c.So(err, c.ShouldBeNil)
		c.So(len(data), c.ShouldEqual, 2)

	})
}

func testDelEndpoint(t *testing.T) {
	endpointBiz := newEndpointBiz()
	c.Convey("Test Del Endpoint", t, func() {
		err := endpointBiz.Delete(context.TODO(), epId)
		c.So(err, c.ShouldBeNil)
	})
}

func testWatcherUpdate(t *testing.T) {
	watcher := newWatcher()
	endpointBiz := newEndpointBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := config.ReadConfig(env)
	cnf := cfg.NewConfig(conf)
	value, err := endpointBiz.Get(context.TODO(), epId)
	if err != nil {
		t.Error(err)
		return
	}
	val, _ := json.Marshal(value)
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
	routers.NewHttpURIRouteToSrvMethod()
	c.Convey("Test Watcher Update", t, func() {

		err = watcher.Handle(context.TODO(), mvccpb.PUT, cnf.GetServiceKey(epId), string(val))
		c.So(err, c.ShouldBeNil)
		r := routers.Get()
		detail := r.GetRoute("/api/v1/example/helloworld")
		c.So(detail, c.ShouldNotBeNil)

	})
	c.Convey("test watcher update fail", t, func() {
		err := watcher.Handle(context.TODO(), mvccpb.PUT, cnf.GetServiceKey("test"), "{")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "unexpected end of JSON inpu")

		err = watcher.Handle(context.TODO(), mvccpb.PUT, cnf.GetServiceKey(epId), "{}")
		c.So(err, c.ShouldNotBeNil)

		patch := gomonkey.ApplyFuncReturn(loadbalance.New, nil, fmt.Errorf("Unknown load balance type"))
		defer patch.Reset()
		err = watcher.Handle(context.TODO(), mvccpb.PUT, cnf.GetServiceKey(epId), string(val))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Unknown load balance type")
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn((*transport.GatewayServer).RegisterService, fmt.Errorf("register error"))
		defer patch2.Reset()

		err = watcher.Handle(context.TODO(), mvccpb.PUT, cnf.GetServiceKey(epId), string(val))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "register error")

	})
}
func testWatcherDel(t *testing.T) {
	watcher := newWatcher()
	endpointBiz := newEndpointBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := config.ReadConfig(env)
	cnf := cfg.NewConfig(conf)
	value, err := endpointBiz.Get(context.TODO(), epId)
	if err != nil {
		t.Error(err)
		return
	}
	val, _ := json.Marshal(value)
	c.Convey("Test Watcher Del", t, func() {
		err := watcher.Handle(context.TODO(), mvccpb.DELETE, cnf.GetServiceKey(epId), string(val))
		c.So(err, c.ShouldBeNil)
		r := routers.Get()
		detail := r.GetRoute("/api/v1/example/helloworld")
		c.So(detail, c.ShouldBeNil)
	})
	c.Convey("test watcher del fail", t, func() {
		err := watcher.Handle(context.TODO(), mvccpb.DELETE, cnf.GetServiceKey("test"), "{")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "end of JSON input")

		err = watcher.Handle(context.TODO(), mvccpb.DELETE, cnf.GetServiceKey(epId), "{}")
		c.So(err, c.ShouldNotBeNil)

		patch := gomonkey.ApplyFuncReturn((*transport.HttpEndpointImpl).DeleteEndpoint, fmt.Errorf("unregister error"))
		defer patch.Reset()
		err = watcher.Handle(context.TODO(), mvccpb.DELETE, cnf.GetServiceKey(epId), string(val))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "unregister error")

	})
}
func TestEndpoint(t *testing.T) {
	t.Run("Test Add Endpoint", testAddEndpoint)
	t.Run("Test Watcher Update", testWatcherUpdate)

	t.Run("Test Get Endpoint", testGetEndpoint)
	t.Run("Test Patch Endpoint", testPatchEndpoint)
	t.Run("Test List Endpoint", testListEndpoints)
	t.Run("Test Watcher Update", testWatcherUpdate)
	t.Run("Test Watcher Del", testWatcherDel)
	t.Run("Test Del Endpoint", testDelEndpoint)
}
