package biz_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/begonia-org/begonia/internal/pkg/utils"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var akskAccess = ""
var akskSecret = ""
var akskAppid = ""

func newGatewayRequest() (*gosdk.GatewayRequest, error) {
	signer := gosdk.NewAppAuthSigner(akskAccess, akskSecret)

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:9527/api/v1/helloworld", strings.NewReader(`{"msg":"hello"}`))
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")

	gw, err := gosdk.NewGatewayRequestFromHttp(req)
	if err != nil {
		return nil, err
	}
	err = signer.SignRequest(gw)
	if err != nil {
		return nil, err
	}
	return gw, nil

}
func newAKSK() *biz.AccessKeyAuth {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewAppRepo(config, logger.Log)
	cnf := cfg.NewConfig(config)
	return biz.NewAccessKeyAuth(repo, cnf, logger.Log)
}

func testGetSecret(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewAppRepo(config, logger.Log)
	snk, _ := tiga.NewSnowflake(1)
	access, _ := utils.GenerateRandomString(32)
	akskAccess = access
	akskSecret, _ = utils.GenerateRandomString(62)
	akskAppid = snk.GenerateIDString()
	appName := fmt.Sprintf("app-AKSK-%s", time.Now().Format("20060102150405"))
	err := repo.Add(context.TODO(), &api.Apps{
		Appid:       akskAppid,
		AccessKey:   akskAccess,
		Secret:      akskSecret,
		Status:      api.APPStatus_APP_ENABLED,
		IsDeleted:   false,
		Name:        appName,
		Description: "test",
		CreatedAt:   timestamppb.New(time.Now()),
		UpdatedAt:   timestamppb.New(time.Now()),
	})

	if err != nil {
		t.Errorf("add app error: %v", err)
	}
	aksk := newAKSK()
	c.Convey("get secret", t, func() {
		sec, err := aksk.GetSecret(context.TODO(), akskAccess)
		c.So(err, c.ShouldBeNil)
		c.So(sec, c.ShouldEqual, akskSecret)
	})
	c.Convey("get secret fail", t, func() {
		_, err := aksk.GetSecret(context.TODO(), "ddddeeeeeedede")
		c.So(err, c.ShouldNotBeNil)
	})
}

func testGetAPPID(t *testing.T) {
	aksk := newAKSK()
	c.Convey("get appid", t, func() {
		appid, err := aksk.GetAppid(context.TODO(), akskAccess)
		c.So(err, c.ShouldBeNil)
		c.So(appid, c.ShouldEqual, akskAppid)
	})

	c.Convey("get appid fail", t, func() {
		appid, err := aksk.GetAppid(context.TODO(), "dddddeeedwecccwcqdq")
		c.So(err, c.ShouldNotBeNil)
		c.So(appid, c.ShouldBeEmpty)
	})
}

func testValidator(t *testing.T) {
	signer := gosdk.NewAppAuthSigner(akskAccess, akskSecret)
	c.Convey("test validator success", t, func() {
		req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:9527/api/v1/helloworld", strings.NewReader(`{"msg":"hello"}`))
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("content-type", "application/json")

		gw, err := gosdk.NewGatewayRequestFromHttp(req)
		c.So(err, c.ShouldBeNil)
		err = signer.SignRequest(gw)
		c.So(err, c.ShouldBeNil)

		aksk := newAKSK()
		ak, err := aksk.AppValidator(context.TODO(), gw)
		c.So(err, c.ShouldBeNil)
		c.So(ak, c.ShouldEqual, akskAccess)
	})
	c.Convey("test validator fail with missing key params", t, func() {
		gw, err := newGatewayRequest()
		c.So(err, c.ShouldBeNil)
		err = signer.SignRequest(gw)
		c.So(err, c.ShouldBeNil)
		xdate := gw.Headers.Get(gosdk.HeaderXDateTime)
		gw.Headers.Del(gosdk.HeaderXDateTime)
		aksk := newAKSK()
		_, err = aksk.AppValidator(context.TODO(), gw)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrAppXDateMissing.Error())

		gw.Headers.Set(gosdk.HeaderXDateTime, xdate)

		authz := gw.Headers.Get(gosdk.HeaderXAuthorization)
		gw.Headers.Del(gosdk.HeaderXAuthorization)
		_, err = aksk.AppValidator(context.TODO(), gw)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrAppSignatureMissing.Error())
		gw.Headers.Set(gosdk.HeaderXAuthorization, authz)

		ak := gw.Headers.Get(gosdk.HeaderXAccessKey)
		gw.Headers.Del(gosdk.HeaderXAccessKey)

		_, err = aksk.AppValidator(context.TODO(), gw)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrAppAccessKeyMissing.Error())

		gw.Headers.Set(gosdk.HeaderXAccessKey, ak)

		newXDate := time.Now().Add(-time.Minute * 120).Format(gosdk.DateFormat)
        gw.Headers.Set(gosdk.HeaderXDateTime, newXDate)

        _, err = aksk.AppValidator(context.TODO(), gw)
        c.So(err.Error(),c.ShouldContainSubstring,errors.ErrRequestExpired.Error())

        newXDate = time.Now().Format("2006-01-02 15:04:05")
        gw.Headers.Set(gosdk.HeaderXDateTime, newXDate)

        _, err = aksk.AppValidator(context.TODO(), gw)
        c.So(err.Error(),c.ShouldContainSubstring,"cannot parse")

        gw.Headers.Set(gosdk.HeaderXDateTime,xdate)
        patch:=gomonkey.ApplyFuncReturn((*gosdk.AppAuthSignerImpl).Sign,"dhewivbdcvnwvwrfvwecfddcddc",nil)
        defer patch.Reset()
        _,err=aksk.AppValidator(context.TODO(),gw)
        c.So(err,c.ShouldNotBeNil)
        c.So(err.Error(),c.ShouldContainSubstring,errors.ErrAppSignatureInvalid.Error())
	})
}

func TestAKSK(t *testing.T){
    t.Run("get secret",testGetSecret)
    t.Run("get appid",testGetAPPID)
    t.Run("validator",testValidator)
}