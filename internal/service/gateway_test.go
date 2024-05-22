package service_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/service"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func postEndpoint(t *testing.T) {
	apiClient := client.NewEndpointAPI(apiAddr, accessKey, secret)

	c.Convey("test create endpoint api", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		endpoint := &api.EndpointSrvConfig{
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
			Tags: []string{"test"},
		}
		resp, err := apiClient.PostEndpointConfig(context.Background(), endpoint)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(resp.Id, c.ShouldNotBeEmpty)
		shareEndpoint = resp.Id
		time.Sleep(5 * time.Second)

		req, err := http.NewRequest("GET", "http://127.0.0.1:12140/api/v1/example/helloworld", nil)
		c.So(err, c.ShouldBeNil)
		req.Header.Set("accept", "application/json")
		helloRsp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(helloRsp.StatusCode, c.ShouldEqual, http.StatusOK)

	})
}

func patchEndpoint(t *testing.T) {
	c.Convey("test patch endpoint api", t, func() {
		apiClient := client.NewEndpointAPI(apiAddr, accessKey, secret)

		patch := &api.EndpointSrvUpdateRequest{
			UniqueKey:   shareEndpoint,
			Description: "test patch",
			Endpoints: []*api.EndpointMeta{
				{
					Addr:   "127.0.0.1:21213",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21214",
					Weight: 0,
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description", "endpoints"}},
		}

		resp, err := apiClient.PatchEndpointConfig(context.Background(), patch)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, common.Code_OK)
		time.Sleep(3 * time.Second)
	})
}

func getEndpoint(t *testing.T) {
	apiClient := client.NewEndpointAPI(apiAddr, accessKey, secret)

	c.Convey("test get endpoint api", t, func() {
		rsp, err := apiClient.GetEndpointDetails(context.Background(), shareEndpoint)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Endpoints.Description, c.ShouldEqual, "test patch")

		// c.So(rsp.Details.Endpoints.Name, c.ShouldEqual, "test")
	})
}
func listEndpoint(t *testing.T) {
	apiClient := client.NewEndpointAPI(apiAddr, accessKey, secret)
	c.Convey("test list endpoint api", t, func() {
		rsp, err := apiClient.List(context.Background(), []string{"test", "test2"}, nil)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(len(rsp.Endpoints), c.ShouldBeGreaterThan, 0)
	})
}
func delEndpoint(t *testing.T) {

	apiClient := client.NewEndpointAPI(apiAddr, accessKey, secret)
	c.Convey("test delete endpoint api", t, func() {
		rsp, err := apiClient.DeleteEndpointConfig(context.Background(), shareEndpoint)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		time.Sleep(5 * time.Second)
		req, err := http.NewRequest("GET", "http://127.0.0.1:12140/api/v1/example/helloworld", nil)
		c.So(err, c.ShouldBeNil)
		req.Header.Set("accept", "application/json")
		helloRsp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(helloRsp.StatusCode, c.ShouldEqual, http.StatusNotFound)
	})

}
func testEndpointSvrErr(t *testing.T) {
	c.Convey("test endpoint server error", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewEndpointSvrForTest(cnf, gateway.Log)
		// _,err:=srv.PostEndpointConfig(context.Background(),nil)
		patch := gomonkey.ApplyFuncReturn((*endpoint.EndpointUsecase).AddConfig, nil, fmt.Errorf("test add endpoint error"))
		defer patch.Reset()
		_, err := srv.Post(context.Background(), nil)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldEqual, "test add endpoint error")
		patch.Reset()

		_, err = srv.Update(context.Background(), &api.EndpointSrvUpdateRequest{})
		c.So(err, c.ShouldNotBeNil)

		_, err = srv.Get(context.Background(), &api.DetailsEndpointRequest{})
		c.So(err, c.ShouldNotBeNil)
		patch = patch.ApplyFuncReturn((*endpoint.EndpointUsecase).Delete, fmt.Errorf("test DEL endpoint error"))
		_, err = srv.Delete(context.Background(), &api.DeleteEndpointRequest{})
		c.So(err, c.ShouldNotBeNil)
		patch.Reset()
		patch = patch.ApplyFuncReturn((*endpoint.EndpointUsecase).List, nil, fmt.Errorf("test list endpoint error"))
		_, err = srv.List(context.Background(), &api.ListEndpointRequest{})
		c.So(err, c.ShouldNotBeNil)
		patch.Reset()
	})
}

func TestEndpoint(t *testing.T) {
	t.Run("post", postEndpoint)
	t.Run("patch", patchEndpoint)
	t.Run("get", getEndpoint)
	t.Run("list", listEndpoint)
	t.Run("testErr", testEndpointSvrErr)
	t.Run("del", delEndpoint)
}
