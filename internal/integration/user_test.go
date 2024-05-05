package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
			name:=fmt.Sprintf("user-%s", time.Now().Format("20060102150405"))
			rsp, err := apiClient.PostUser(context.Background(), &api.PostUserRequest{
				Name:     name,
				Password: "123456",
				Email:    fmt.Sprintf("%s@example.com",name),
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
			rsp, err := apiClient.PatchUser(context.Background(), uid, map[string]interface{}{
				"password": "123456ecfasddccddd",
				"email":    "begonia01@example.com"})
			c.So(err, c.ShouldBeNil)
			c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		})
}
func TestUser(t *testing.T) {
	t.Run("add user", addUser)
	t.Run("get user", getUser)
	// uid = "442210231930327040"
	t.Run("patch user", patchUser)
	t.Run("delete user", deleteUser)
}
