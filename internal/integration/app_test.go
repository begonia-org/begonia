package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
)

var appid = ""

func addApp(t *testing.T) {
	c.Convey(
		"test add app",
		t,
		func() {
			apiClient := client.NewAppAPI(apiAddr, accessKey, secret)
			rsp, err := apiClient.PostAppConfig(context.Background(), &api.AppsRequest{Name: fmt.Sprintf("app-%s",time.Now().Format("20060102150405")), Description: "test"})
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp.Appid, c.ShouldNotBeEmpty)
			appid = rsp.Appid

		},
	)
}

func getApp(t *testing.T) {
	c.Convey(
		"test get app",
		t,
		func() {
			apiClient := client.NewAppAPI(apiAddr, accessKey, secret)
			rsp2, err := apiClient.GetAPP(context.Background(), appid)
			c.So(err, c.ShouldBeNil)
			c.So(rsp2.StatusCode, c.ShouldEqual, common.Code_OK)

		},
	)
}
func delApp(t *testing.T) {
	c.Convey(
		"test del app",
		t,
		func() {
			apiClient := client.NewAppAPI(apiAddr, accessKey, secret)

			rsp3, err := apiClient.DeleteAPP(context.TODO(), appid)
			c.So(err, c.ShouldBeNil)
			c.So(rsp3.StatusCode, c.ShouldEqual, common.Code_OK)

			_, err = apiClient.GetAPP(context.Background(), appid)
			c.So(err, c.ShouldNotBeNil)
		})
}
func TestApp(t *testing.T) {
	t.Run("add app", addApp)
	t.Run("get app", getApp)
	// appid = "442568851213783040"
	t.Run("del app", delApp)
}
