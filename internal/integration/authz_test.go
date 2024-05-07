package integration_test

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	sys "github.com/begonia-org/go-sdk/api/sys/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/encoding/protojson"
)

var xtoken = ""

// loadPublicKey 从文件中加载 RSA 公钥
func loadPublicKeyFromFile(filePath string) (*rsa.PublicKey, error) {
	// 读取密钥文件
	keyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解码 PEM 格式的密钥
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// 解析 PKIX 格式的公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey.(*rsa.PublicKey), nil
}
func loginTest(t *testing.T) {
	c.Convey(
		"test login",
		t,
		func() {
			apiClient := client.NewAuthzAPI(apiAddr, accessKey, secret)
			_, filename, _, _ := runtime.Caller(0)

			file := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "cert", "auth_public_key.pem")

			pubkey, err := loadPublicKeyFromFile(file)
			c.So(err, c.ShouldBeNil)
			rsp, err := apiClient.Login(context.Background(), "admin", "admin", pubkey, true)
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp.Token, c.ShouldNotBeEmpty)

			req, err := http.NewRequest(http.MethodGet, apiAddr+"/v1/gateway/info", nil)
			c.So(err, c.ShouldBeNil)
			xtoken = rsp.Token
			req.Header.Set("Authorization", "Bearer "+rsp.Token)
			req.Header.Set("accept", "application/json")
			resp, err := http.DefaultClient.Do(req)
			c.So(err, c.ShouldBeNil)
			defer resp.Body.Close()
			c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)

			apiRsp := &common.HttpResponse{}
			data, _ := io.ReadAll(resp.Body)
			err = protojson.Unmarshal(data, apiRsp)

			c.So(err, c.ShouldBeNil)
			c.So(apiRsp.Code, c.ShouldEqual, common.Code_OK)
			bData,err:=apiRsp.Data.MarshalJSON()
			c.So(err,c.ShouldBeNil)
			info:= &sys.InfoResponse{}
			err=protojson.Unmarshal(bData,info)
			c.So(err,c.ShouldBeNil)
			t.Log(info.Version,info.BuildTime,info.Commit)
			// c.So(info.Name,c.ShouldEqual,"gateway")

		})
}
func testLogout(t *testing.T) {
	c.Convey(
		"test logout",
		t,
		func() {
			apiClient := client.NewAuthzAPI(apiAddr, accessKey, secret)
			rsp, err := apiClient.Logout(context.Background(), xtoken)
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)

			req, err := http.NewRequest(http.MethodGet, apiAddr+"/v1/gateway/info", nil)
			c.So(err, c.ShouldBeNil)
			req.Header.Set("Authorization", "Bearer "+xtoken)
			req.Header.Set("accept", "application/json")
			resp, err := http.DefaultClient.Do(req)
			c.So(err, c.ShouldBeNil)
			defer resp.Body.Close()
			c.So(resp.StatusCode, c.ShouldEqual, http.StatusUnauthorized)

			apiRsp := &common.HttpResponse{}
			data, _ := io.ReadAll(resp.Body)
			err = protojson.Unmarshal(data, apiRsp)

			c.So(err, c.ShouldBeNil)
			c.So(apiRsp.Code, c.ShouldNotEqual, int32(common.Code_OK))
		},
	)
}
func TestAuth(t *testing.T) {
	t.Run("login", loginTest)
	t.Run("logout", testLogout)
}
