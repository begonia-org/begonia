package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia/internal/service"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

var bn = ""
var bid = ""

func testAddBusinessService(t *testing.T) {
	c.Convey("test add business", t, func() {
		apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)
		snk, _ := tiga.NewSnowflake(2)
		bn = fmt.Sprintf("test-service-%s", snk.GenerateIDString())
		rsp, err := apiClient.PostBusiness(context.Background(), bn, "test", []string{"test"})
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.BusinessId, c.ShouldNotBeEmpty)
		bid = rsp.BusinessId
	})
	c.Convey("test add business duplicate", t, func() {
		apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)
		_, err := apiClient.PostBusiness(context.Background(), bn, "test", []string{"test"})
		c.So(err, c.ShouldNotBeNil)
		t.Log(err.Error())
		// c.So(rsp.StatusCode, c.ShouldEqual, common.Code_CONFLICT)
		// c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")
	})
	c.Convey("test no id", t, func() {
		patch := gomonkey.ApplyFuncReturn(service.GetIdentity, "")
		defer patch.Reset()
		apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)
		_, err := apiClient.PostBusiness(context.Background(), bn, "test", []string{"test"})
		c.So(err, c.ShouldNotBeNil)
	})
}

func testGetBusinessService(t *testing.T) {
	apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)

	c.Convey("test get business by id", t, func() {
		rsp, err := apiClient.GetBusiness(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.BusinessName, c.ShouldEqual, bn)
	})
	c.Convey("test get business by name", t, func() {
		rsp, err := apiClient.GetBusiness(context.Background(), bn)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.BusinessName, c.ShouldEqual, bn)
	})
	c.Convey("test get business id not found", t, func() {
		rsp, err := apiClient.GetBusiness(context.Background(), "not found")
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_NOT_FOUND))
	})
}

func testUpdateBusinessService(t *testing.T) {
	apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)

	c.Convey("test update business", t, func() {
		rsp, err := apiClient.PatchBusiness(context.Background(), bid, client.WithPatchParams("description", "update desc"))
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Description, c.ShouldEqual, "update desc")
		c.So(rsp.BusinessName, c.ShouldEqual, bn)
		c.So(rsp.Tags[0], c.ShouldEqual, "test")
	})
	c.Convey("test update with duplicate name", t, func() {
		snk, _ := tiga.NewSnowflake(2)
		bn2 := fmt.Sprintf("test-service2-%s", snk.GenerateIDString())
		rsp, err := apiClient.PostBusiness(context.Background(), bn2, "test", []string{"test"})
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		// bu := &api.PatchBusinessRequest{}
		rsp2, err := apiClient.PatchBusiness(context.Background(), bid, client.WithPatchParams("business_name", bn2), client.WithPatchParams("name", bn2), client.WithPatchParams("description", "update desc"))
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp2.StatusCode, c.ShouldEqual, int(common.Code_CONFLICT))
	})
	c.Convey("test update with not found id", t, func() {
		snk, _ := tiga.NewSnowflake(2)

		rsp, err := apiClient.PatchBusiness(context.Background(), snk.GenerateIDString(), client.WithPatchParams("description", "update desc"))
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_NOT_FOUND))
	})
}

func testListBusinessService(t *testing.T) {
	apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)

	c.Convey("test list business", t, func() {
		rsp, err := apiClient.ListBusiness(context.Background(), []string{"test"}, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(len(rsp.Business), c.ShouldBeGreaterThanOrEqualTo, 1)

		rsp, err = apiClient.ListBusiness(context.Background(), []string{"test2"}, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(len(rsp.Business), c.ShouldEqual, 0)

	})
	c.Convey("test list business with error", t, func() {
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination, fmt.Errorf("pagination error"))
		defer patch.Reset()
		rsp, err := apiClient.ListBusiness(context.Background(), []string{"test"}, 1, 20)
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_INTERNAL_ERROR))
	})
}
func testDeleteBusinessService(t *testing.T) {
	apiClient := client.NewBusinessAPI(apiAddr, accessKey, secret)

	c.Convey("test delete business", t, func() {
		rsp, err := apiClient.DeleteBusiness(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
	})
	c.Convey("test delete business with not found id", t, func() {
		rsp, err := apiClient.DeleteBusiness(context.Background(), bid)
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_NOT_FOUND))
	})
}

func TestBusinessService(t *testing.T) {
	t.Run("test add business", testAddBusinessService)
	t.Run("test get business", testGetBusinessService)
	t.Run("test update business", testUpdateBusinessService)
	t.Run("test list business", testListBusinessService)
	t.Run("test delete business", testDeleteBusinessService)
}
