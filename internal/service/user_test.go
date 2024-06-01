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
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
)

var uid = ""

func addUser(t *testing.T) {
	c.Convey(
		"test add user",
		t,
		func() {
			apiClient := client.NewUsersAPI(apiAddr, accessKey, secret)
			name := fmt.Sprintf("user-service-test-%s", time.Now().Format("20060102150405"))
			rsp, err := apiClient.PostUser(context.Background(), &api.PostUserRequest{
				Name:     name,
				Password: "123456",
				Email:    fmt.Sprintf("%s@example.com", name),
				Role:     api.Role_ADMIN,
				Dept:     "development",
				Avatar:   "https://www.example.com/avatar.jpg",
				Owner:    "test-user-01",
				Phone:    time.Now().Format("20060102150405"),
			})
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
			c.So(rsp.Uid, c.ShouldNotBeEmpty)
			uid = rsp.Uid

		})
}

func getUser(t *testing.T) {
	c.Convey(
		"test get user",
		t,
		func() {
			apiClient := client.NewUsersAPI(apiAddr, accessKey, secret)
			rsp, err := apiClient.GetUser(context.Background(), uid)
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		})
}
func deleteUser(t *testing.T) {
	c.Convey(
		"test delete user",
		t,
		func() {
			apiClient := client.NewUsersAPI(apiAddr, accessKey, secret)
			rsp, err := apiClient.DeleteUser(context.Background(), uid)
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)

			_, err = apiClient.GetUser(context.Background(), uid)
			c.So(err, c.ShouldNotBeNil)

		})
}
func patchUser(t *testing.T) {
	c.Convey(
		"test patch user",
		t,
		func() {
			apiClient := client.NewUsersAPI(apiAddr, accessKey, secret)
			rsp, err := apiClient.UpdateUser(context.Background(), uid, map[string]interface{}{
				"password": "123456ecfasddccddd",
				"email":    fmt.Sprintf("%s@example.com", time.Now().Format("20060102150405"))})
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		})
}
func testRegisterErr(t *testing.T) {
	c.Convey(
		"test register user error",
		t,
		func() {
			env := "dev"
			if begonia.Env != "" {
				env = begonia.Env
			}
			cnf := config.ReadConfig(env)
			srv := service.NewUserSvrForTest(cnf, gateway.Log)
			patch := gomonkey.ApplyFuncReturn((*biz.UserUsecase).Add, fmt.Errorf("test add user error"))
			defer patch.Reset()
			_, err := srv.Register(context.Background(), &api.PostUserRequest{})
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, "test add user error")

		})
}
func testUpdateErr(t *testing.T) {
	c.Convey("test update user error", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewUserSvrForTest(cnf, gateway.Log)
		patch := gomonkey.ApplyFuncReturn((*biz.UserUsecase).Update, fmt.Errorf("test update user error"))
		defer patch.Reset()
		_, err := srv.Update(context.Background(), &api.PatchUserRequest{Uid: "", Owner: "test-user-01"})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test update user error")
		patch.Reset()
	})
}
func testDelUserErr(t *testing.T) {
	c.Convey("test delete user error", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewUserSvrForTest(cnf, gateway.Log)
		patch := gomonkey.ApplyFuncReturn((*biz.UserUsecase).Delete, fmt.Errorf("test delete user error"))
		defer patch.Reset()
		_, err := srv.Delete(context.Background(), &api.DeleteUserRequest{Uid: ""})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test delete user error")
		patch.Reset()
	})
}
func TestUser(t *testing.T) {
	t.Run("add user", addUser)
	t.Run("get user", getUser)
	// uid = "442210231930327040"
	t.Run("patch user", patchUser)
	t.Run("delete user", deleteUser)
	t.Run("test register user error", testRegisterErr)
	t.Run("test update user error", testUpdateErr)
	t.Run("test delete user error", testDelUserErr)
}
