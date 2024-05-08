package data

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/cockroachdb/errors"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var appid = ""
var accessKey = ""
var secret = ""
var appName = ""

func generateRandomString(n int) (string, error) {
	const lettersAndDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("Failed to generate random string: %w", err)
	}

	for i := 0; i < n; i++ {
		// 将随机字节转换为lettersAndDigits中的一个有效字符
		b[i] = lettersAndDigits[b[i]%byte(len(lettersAndDigits))]
	}

	return string(b), nil
}
func addTest(t *testing.T) {
	c.Convey("test app add success", t, func() {
		t.Log("add test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAppRepo(cfg.ReadConfig(env), logger.Log)
		snk, _ := tiga.NewSnowflake(1)
		access, _ := generateRandomString(32)
		accessKey = access
		secret, _ = generateRandomString(62)
		appid = snk.GenerateIDString()
		appName = fmt.Sprintf("app-%s", time.Now().Format("20060102150405"))
		err := repo.Add(context.TODO(), &api.Apps{
			Appid:       appid,
			AccessKey:   access,
			Secret:      secret,
			Status:      api.APPStatus_APP_ENABLED,
			IsDeleted:   false,
			Name:        appName,
			Description: "test",
			CreatedAt:   timestamppb.New(time.Now()),
			UpdatedAt:   timestamppb.New(time.Now()),
		})

		c.So(err, c.ShouldBeNil)
		cfg := config.NewConfig(cfg.ReadConfig(env))
		cacheKey := cfg.GetAPPAccessKey(access)

		value, err := layered.Get(context.Background(), cacheKey)
		c.So(err, c.ShouldBeNil)
		c.So(string(value), c.ShouldEqual, secret)

		patch := gomonkey.ApplyFuncReturn((*LayeredCache).Get, nil, errors.New("error"))

		defer patch.Reset()
		val, err := repo.GetSecret(context.Background(), access)
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldEqual, secret)

	})
}
func getTest(t *testing.T) {
	c.Convey("test app get success", t, func() {
		t.Log("get test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAppRepo(cfg.ReadConfig(env), logger.Log)
		app, err := repo.Get(context.TODO(), appid)
		c.So(err, c.ShouldBeNil)
		c.So(app.Appid, c.ShouldEqual, appid)
	    app, err = repo.Get(context.TODO(), accessKey)
		c.So(err, c.ShouldBeNil)
		c.So(app.AccessKey, c.ShouldEqual, accessKey)
		_, err = repo.Get(context.TODO(), "123")
		c.So(err, c.ShouldNotBeNil)

	})
}

func duplicateNameTest(t *testing.T) {
	c.Convey("test app add duplicate name", t, func() {
		t.Log("add test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAppRepo(cfg.ReadConfig(env), logger.Log)
		// snk, _ := tiga.NewSnowflake(1)
		access, _ := generateRandomString(32)
		accessKey = access
		secret, _ = generateRandomString(62)
		// appid = snk.GenerateIDString()
		err := repo.Add(context.TODO(), &api.Apps{
			Appid:       appid,
			AccessKey:   access,
			Secret:      secret,
			Status:      api.APPStatus_APP_ENABLED,
			IsDeleted:   false,
			Name:        fmt.Sprintf("app-%s", time.Now().Format("20060102150405")),
			Description: "test",
			CreatedAt:   timestamppb.New(time.Now()),
			UpdatedAt:   timestamppb.New(time.Now()),
		})

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")
	})
}
func patchTest(t *testing.T) {
	c.Convey("test app patch success", t, func() {
		t.Log("patch test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAppRepo(cfg.ReadConfig(env), logger.Log)
		err := repo.Patch(context.TODO(), &api.Apps{
			Appid:       appid,
			AccessKey:   accessKey,
			Secret:      secret,
			Status:      api.APPStatus_APP_DISABLED,
			Name:        fmt.Sprintf("app-%s", time.Now().Format("20060102150405")),
			Description: "test UPDATE",
			CreatedAt:   timestamppb.New(time.Now()),
			UpdatedAt:   timestamppb.New(time.Now()),
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"status", "description"},
			},
		})

		c.So(err, c.ShouldBeNil)
		t.Log("get test", appid)
		updated,err:=repo.Get(context.Background(),appid)
		c.So(err,c.ShouldBeNil)
		c.So(updated.Status,c.ShouldEqual,api.APPStatus_APP_DISABLED)
		c.So(updated.Description,c.ShouldEqual,"test UPDATE")
		c.So(updated.Name,c.ShouldEqual,appName)
	})

}
func delTest(t *testing.T) {
	c.Convey("test app delete success", t, func() {
		t.Log("delete test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAppRepo(cfg.ReadConfig(env), logger.Log)
		err := repo.Del(context.TODO(), appid)
		c.So(err, c.ShouldBeNil)
		_, err = repo.Get(context.TODO(), appid)
		c.So(err, c.ShouldNotBeNil)
	})
}
func TestApp(t *testing.T) {
	t.Run("add app", addTest)
	t.Run("get app", getTest)
	t.Run("add duplicate name", duplicateNameTest)
	t.Run("patch app", patchTest)
	t.Run("del app", delTest)
}
