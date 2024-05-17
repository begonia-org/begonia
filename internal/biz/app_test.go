package biz_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/utils"
	api "github.com/begonia-org/go-sdk/api/app/v1"

	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var appid = ""
var accessKey = ""
var secret = ""
var appName = ""
var appName2 = ""

func newAppBiz() *biz.AppUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewAppRepo(config, gateway.Log)
	cnf := cfg.NewConfig(config)
	return biz.NewAppUsecase(repo, cnf)
}

func testPutApp(t *testing.T) {
	appBiz := newAppBiz()
	var app *api.Apps
	var err error
	snk, _ := tiga.NewSnowflake(1)

	c.Convey("test app put success", t, func() {
		appName = fmt.Sprintf("app-%s", time.Now().Format("20060102150405"))
		app, err = appBiz.CreateApp(context.TODO(), &api.AppsRequest{
			Name:        appName,
			Description: "test",
			Tags:        []string{"test-app"},
		}, "396870469984194560")
		c.So(err, c.ShouldBeNil)
		accessKey = app.AccessKey
		secret = app.Secret
		appid = app.Appid

		layered := data.NewLayered(config.ReadConfig("dev"), gateway.Log)
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		prefix := cnf.GetAPPAccessKeyPrefix()
		key := fmt.Sprintf("%s:%s", prefix, accessKey)
		sec, err := layered.Get(context.Background(), key)
		c.So(err, c.ShouldBeNil)
		c.So(string(sec), c.ShouldEqual, secret)

	})
	c.Convey("test app put failed", t, func() {
		app.Appid = snk.GenerateIDString()
		app.Id = 0

		err = appBiz.Put(context.TODO(), app, "396870469984194560")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")
		newAccessKey, _ := utils.GenerateRandomString(32)
		newSecret, _ := utils.GenerateRandomString(64)
		newApp := &api.Apps{
			Appid:       snk.GenerateIDString(),
			AccessKey:   newAccessKey + newSecret,
			Secret:      newSecret,
			Description: "test",
		}

		err = appBiz.Put(context.TODO(), newApp, "396870469984194560")
		c.So(err, c.ShouldNotBeNil)

		c.So(err.Error(), c.ShouldContainSubstring, "too long")
		appName2 = fmt.Sprintf("app-biz-2-%s", time.Now().Format("20060102150405"))
		access2, _ := utils.GenerateRandomString(32)
		secret2, _ := utils.GenerateRandomString(64)
		app2 := &api.Apps{
			Appid:       snk.GenerateIDString(),
			AccessKey:   access2,
			Secret:      secret2,
			Status:      api.APPStatus_APP_ENABLED,
			IsDeleted:   false,
			Name:        appName2,
			Description: "test",
			CreatedAt:   timestamppb.New(time.Now()),
			UpdatedAt:   timestamppb.New(time.Now()),
		}
		err = appBiz.Put(context.TODO(), app2, "396870469984194560")
		c.So(err, c.ShouldBeNil)
		err = appBiz.Put(context.TODO(), app2, "396870469984194560")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")
	})
}

func testGetApp(t *testing.T) {
	appBiz := newAppBiz()

	c.Convey("test app get success", t, func() {
		app, err := appBiz.Get(context.TODO(), appid)
		c.So(err, c.ShouldBeNil)
		c.So(app, c.ShouldNotBeNil)
		c.So(app.Appid, c.ShouldEqual, appid)
		c.So(app.Name, c.ShouldEqual, appName)
		c.So(app.AccessKey, c.ShouldEqual, accessKey)
		c.So(app.Secret, c.ShouldEqual, secret)

	})
	c.Convey("test app get failed", t, func() {
		_, err := appBiz.Get(context.TODO(), "123456")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}

func testPatchApp(t *testing.T) {
	appBiz := newAppBiz()

	c.Convey("test app patch success", t, func() {
		app, err := appBiz.Get(context.TODO(), appid)
		c.So(err, c.ShouldBeNil)
		c.So(app, c.ShouldNotBeNil)
		c.So(app.Appid, c.ShouldEqual, appid)
		c.So(app.Name, c.ShouldEqual, appName)
		c.So(app.AccessKey, c.ShouldEqual, accessKey)
		c.So(app.Secret, c.ShouldEqual, secret)
		app.Name = "patch"
		updated, err := appBiz.Patch(context.TODO(), &api.AppsRequest{
			Appid: appid,
			Name:  "patch-app",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
				"name",
			}},
		}, "")
		c.So(err, c.ShouldBeNil)
		app, err = appBiz.Get(context.TODO(), appid)
		c.So(err, c.ShouldBeNil)
		c.So(updated, c.ShouldNotBeNil)
		c.So(updated.Name, c.ShouldEqual, "patch-app")
		c.So(updated.AccessKey, c.ShouldEqual, app.AccessKey)
		c.So(updated.Secret, c.ShouldEqual, app.Secret)
		c.So(updated.Name, c.ShouldEqual, app.Name)
	})

	c.Convey("test app patch failed", t, func() {
		_, err := appBiz.Patch(context.TODO(), &api.AppsRequest{
			Appid:      "123456",
			Name:       "patch-app-2",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		}, "")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

		_, err = appBiz.Patch(context.TODO(), &api.AppsRequest{
			Appid:      appid,
			Name:       appName2,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		}, "")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")

	})
}

func testListApp(t *testing.T) {
	appBiz := newAppBiz()

	c.Convey("test app list success", t, func() {
		apps, err := appBiz.List(context.TODO(), &api.AppsListRequest{
			PageSize: 10,
			Page:     1,
		})
		c.So(err, c.ShouldBeNil)
		c.So(apps, c.ShouldNotBeNil)
		c.So(len(apps), c.ShouldBeGreaterThan, 0)

	})
	c.Convey("test app list failed", t, func() {
		apps, err := appBiz.List(context.TODO(), &api.AppsListRequest{
			PageSize: 10,
			Page:     1,
			Tags:     []string{"not-exist"},
			Status:   []api.APPStatus{api.APPStatus_APP_DISABLED},
		})
		c.So(len(apps), c.ShouldEqual, 0)
		c.So(err, c.ShouldBeNil)
	})
}

func testDelApp(t *testing.T) {
	c.Convey("test app del success", t, func() {
		appBiz := newAppBiz()
		err := appBiz.Del(context.TODO(), appid)
		c.So(err, c.ShouldBeNil)
		_, err = appBiz.Get(context.TODO(), appid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}

func TestAppBiz(t *testing.T) {
	t.Run("testPutApp", testPutApp)
	t.Run("testGetApp", testGetApp)
	t.Run("testPatchApp", testPatchApp)
	t.Run("testListApp", testListApp)
	t.Run("testDelApp", testDelApp)
}
