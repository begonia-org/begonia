package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/service"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
)

var appid = ""
var name2 = ""

func addApp(t *testing.T) {
	c.Convey(
		"test add app",
		t,
		func() {
			apiClient := client.NewAppAPI(apiAddr, accessKey, secret)
			name := fmt.Sprintf("app-%s", time.Now().Format("20060102150405"))
			rsp, err := apiClient.PostAppConfig(context.Background(), &api.AppsRequest{Name: name, Description: "test", Tags: []string{"test"}})
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp.Appid, c.ShouldNotBeEmpty)
			appid = rsp.Appid

			rsp2, err := apiClient.GetAPP(context.Background(), appid)
			c.So(err, c.ShouldBeNil)
			c.So(rsp2.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp2.Name, c.ShouldNotBeEmpty)
			_, err = apiClient.PostAppConfig(context.Background(), &api.AppsRequest{Name: name, Description: "test"})
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "duplicate app name")
			// c.So(rsp3.StatusCode, c.ShouldEqual, common.Code_ERR)

			name2 = fmt.Sprintf("app-service-2-%s", time.Now().Format("20060102150405"))
			rsp, err = apiClient.PostAppConfig(context.Background(), &api.AppsRequest{Name: name2, Description: "test"})
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)

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
func testPatchApp(t *testing.T) {
	c.Convey(
		"test patch app",
		t,
		func() {
			apiClient := client.NewAppAPI(apiAddr, accessKey, secret)
			name := fmt.Sprintf("app-%s", time.Now().Format("20060102150405"))
			rsp2, err := apiClient.UpdateAPP(context.Background(), appid, name, "test patch", nil)
			c.So(err, c.ShouldBeNil)
			c.So(rsp2.StatusCode, c.ShouldEqual, common.Code_OK)
			rsp2, err = apiClient.GetAPP(context.Background(), appid)
			c.So(err, c.ShouldBeNil)
			c.So(rsp2.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp2.Name, c.ShouldEqual, name)

			_, err = apiClient.UpdateAPP(context.Background(), appid, name2, "test patch", nil)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "duplicate app name")

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

			_, err = apiClient.DeleteAPP(context.TODO(), appid)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "app not found")
			// c.So(rsp3.StatusCode, c.ShouldEqual, common.Code_OK)
		})
}

func listAPP(t *testing.T) {
	c.Convey(
		"test list app",
		t,
		func() {
			apiClient := client.NewAppAPI(apiAddr, accessKey, secret)
			rsp, err := apiClient.ListAPP(context.Background(), []string{"test", "test2"}, []api.APPStatus{api.APPStatus_APP_ENABLED}, 1, 10)
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp.Apps, c.ShouldNotBeEmpty)
			c.So(rsp.Apps[0].Appid, c.ShouldNotBeEmpty)
		},
	)
}
func testListErr(t *testing.T) {
	c.Convey(
		"test list app",
		t,
		func() {
			env := "dev"
			if begonia.Env != "" {
				env = begonia.Env
			}
			cnf := config.ReadConfig(env)
			srv := service.NewAPPSvrForTest(cnf, gateway.Log)
			patch := gomonkey.ApplyFuncReturn((*biz.AppUsecase).List, nil, fmt.Errorf("test list app error"))
			defer patch.Reset()
			_, err := srv.List(context.Background(), nil)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldEqual, "test list app error")
			patch.Reset()
		},
	)
}
func TestApp(t *testing.T) {
	t.Run("add app", addApp)
	t.Run("get app", getApp)
	t.Run("list app", listAPP)
	t.Run("list app err", testListErr)
	t.Run("patch app", testPatchApp)
	// appid = "442568851213783040"
	t.Run("del app", delApp)

}
