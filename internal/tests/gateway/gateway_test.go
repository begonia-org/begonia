package gateway_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"


	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func postEndpoint(t *testing.T) {
	client := gosdk.NewBegoniaClient("http://127.0.0.1:12140", "NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")

	c.Convey("test create endpoint api", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
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
		resp, err := client.PostEndpointConfig(context.Background(), endpoint)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(resp.Id, c.ShouldNotBeEmpty)
		shareEndpoint = resp.Id
		time.Sleep(3 * time.Second)

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
		client := gosdk.NewBegoniaClient("http://127.0.0.1:12140", "NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")

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
			Mask: &fieldmaskpb.FieldMask{Paths: []string{"description", "endpoints"}},
		}

		resp, err := client.PatchEndpointConfig(context.Background(), patch)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, common.Code_OK)
		time.Sleep(3 * time.Second)
	})
}

func getEndpoint(t *testing.T) {
	client := gosdk.NewBegoniaClient("http://127.0.0.1:12140", "NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")

	c.Convey("test get endpoint api", t, func() {
		rsp, err := client.GetEndpointDetails(context.Background(), shareEndpoint)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Endpoints.Description, c.ShouldEqual, "test patch")

		// c.So(rsp.Details.Endpoints.Name, c.ShouldEqual, "test")
	})
}

func delEndpoint(t *testing.T) {

	client := gosdk.NewBegoniaClient("http://127.0.0.1:12140", "NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")
	c.Convey("test delete endpoint api", t, func() {
		rsp, err := client.DeleteEndpointConfig(context.Background(), shareEndpoint)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		time.Sleep(3 * time.Second)
		req, err := http.NewRequest("GET", "http://127.0.0.1:12140/api/v1/example/helloworld", nil)
		c.So(err, c.ShouldBeNil)
		req.Header.Set("accept", "application/json")
		helloRsp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(helloRsp.StatusCode, c.ShouldEqual, http.StatusNotFound)
	})

}
func TestEndpoint(t *testing.T) {
	t.Run("post", postEndpoint)
	t.Run("patch", patchEndpoint)
	t.Run("get", getEndpoint)
	t.Run("del", delEndpoint)
}
