package integration_test

import (
	"context"
	"testing"

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
			rsp, err := apiClient.PostUser(context.Background(), &api.PostUserRequest{
				Name:     "test-user-01",
				Password: "123456",
				Email:    "begonia@example.com",
				Role:     api.Role_ADMIN,
				Dept:     "development",
				Avatar:   "https://www.example.com/avatar.jpg",
				Owner:    "test-user-01",
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

			_,err=apiClient.GetUser(context.Background(),uid)
			c.So(err,c.ShouldNotBeNil)

		})
	}
func TestUser(t *testing.T) {
	t.Run("add user", addUser)
	t.Run("get user", getUser)
	t.Run("delete user", deleteUser)
}