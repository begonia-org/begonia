package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia/internal/service"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

var tid = ""
var tn = ""
var tbn = ""
var tbid = ""

func testAddTenant(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	snk, _ := tiga.NewSnowflake(2)

	c.Convey("test add tenant", t, func() {
		tn = fmt.Sprintf("test-add-tenant-%s", snk.GenerateIDString())
		rsp, err := apiClient.RegisterTenant(context.Background(), tn, "test tenant", fmt.Sprintf("%s@example.com", tn), []string{"test"})
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_OK))
		c.So(rsp.TenantId, c.ShouldNotBeEmpty)
		tid = rsp.TenantId

	})
	c.Convey("test add tenant no creator", t, func() {
		patch := gomonkey.ApplyFuncReturn(service.GetIdentity, "")
		defer patch.Reset()
		rsp, err := apiClient.RegisterTenant(context.Background(), tn, "test tenant", fmt.Sprintf("%s@example.com", tn), []string{"test"})
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(api.UserSvrCode_USER_IDENTITY_MISSING_ERR))

	})
}
func testUpdateTenant(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	c.Convey("test update tenant", t, func() {
		rsp, err := apiClient.PatchTenant(context.Background(), tid, client.WithPatchParams("description", "update tenant"))
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_OK))
		c.So(rsp.TenantName, c.ShouldEqual, tn)
		c.So(rsp.Description, c.ShouldEqual, "update tenant")

	})

}

func testGetTenant(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	c.Convey("test get tenant", t, func() {
		rsp, err := apiClient.GetTenant(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_OK))
		c.So(rsp.TenantName, c.ShouldEqual, tn)
		c.So(rsp.Description, c.ShouldEqual, "update tenant")
	})

}
func testListTenant(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	c.Convey("test list tenant", t, func() {
		rsp, err := apiClient.ListTenants(context.Background(), 1, 10, []string{"test"}, []string{api.TENANTS_STATUS_TENANTS_ACTIVE.String()})
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Tenants, c.ShouldNotBeEmpty)
		c.So(len(rsp.Tenants), c.ShouldBeGreaterThanOrEqualTo, 1)
		snk, _ := tiga.NewSnowflake(1)
		rsp2, err2 := apiClient.ListTenants(context.Background(), 1, 10, []string{snk.GenerateIDString()}, []string{api.TENANTS_STATUS_TENANTS_ACTIVE.String()})
		c.So(err2, c.ShouldBeNil)
		c.So(rsp2.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp2.Tenants, c.ShouldBeEmpty)
		c.So(len(rsp2.Tenants), c.ShouldEqual, 0)

		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination, fmt.Errorf("pagination error"))
		defer patch.Reset()
		rsp3, err3 := apiClient.ListTenants(context.Background(), 1, 10, []string{"test"}, []string{api.TENANTS_STATUS_TENANTS_ACTIVE.String()})
		c.So(err3, c.ShouldNotBeNil)
		c.So(rsp3.StatusCode, c.ShouldEqual, int(common.Code_INTERNAL_ERROR))
	})

}

func testAddTenantBusiness(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	businessApi := client.NewBusinessAPI(apiAddr, accessKey, secret)
	snk, _ := tiga.NewSnowflake(1)
	c.Convey("test add tenant business", t, func() {
		bRsp, err := businessApi.PostBusiness(context.Background(), fmt.Sprintf("test-business-%s", snk.GenerateIDString()), "test-business", []string{"test-plan"})
		c.So(err, c.ShouldBeNil)
		tbid = bRsp.BusinessId
		rsp, err := apiClient.AddTenantBusiness(context.Background(), tid, tbid, "test-plan")
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_OK))
		c.So(rsp.TenantId, c.ShouldEqual, tid)
		c.So(rsp.BusinessId, c.ShouldNotBeEmpty)
	})
	c.Convey("test add tenant business no creator", t, func() {
		patch := gomonkey.ApplyFuncReturn(service.GetIdentity, "")
		defer patch.Reset()
		rsp, err := apiClient.AddTenantBusiness(context.Background(), tid, tbid, "test-plan")
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(api.UserSvrCode_USER_IDENTITY_MISSING_ERR))
	})
}
func testListTenantBusiness(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	c.Convey("test list tenant business", t, func() {
		rsp, err := apiClient.ListTenantBusiness(context.Background(), tid, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.TenantsBusiness, c.ShouldNotBeEmpty)
		c.So(len(rsp.TenantsBusiness), c.ShouldBeGreaterThanOrEqualTo, 1)
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination, fmt.Errorf("pagination error"))
		defer patch.Reset()
		rsp2, err2 := apiClient.ListTenantBusiness(context.Background(), tid, 1, 10)
		c.So(err2, c.ShouldNotBeNil)
		c.So(rsp2.StatusCode, c.ShouldEqual, int(common.Code_INTERNAL_ERROR))
	})
}

func testDeleteTenantBusiness(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	c.Convey("test delete tenant business", t, func() {
		rsp, err := apiClient.DeleteTenantBusiness(context.Background(), tid, tbid)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
	})
	c.Convey("test delete tenant error", t, func() {
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.UpdateSelectColumns, fmt.Errorf("update delete error"))
		defer patch.Reset()
		rsp, err := apiClient.DeleteTenantBusiness(context.Background(), tid, tbid)
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_INTERNAL_ERROR))
	})
}
func testDeleteTenant(t *testing.T) {
	apiClient := client.NewTenantAPI(apiAddr, accessKey, secret)
	c.Convey("test delete tenant", t, func() {
		rsp, err := apiClient.DeleteTenant(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
	})
	c.Convey("test delete tenant error", t, func() {
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.UpdateSelectColumns, fmt.Errorf("update delete error"))
		defer patch.Reset()
		rsp, err := apiClient.DeleteTenant(context.Background(), tid)
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, int(common.Code_INTERNAL_ERROR))
	})
}

func TestTenant(t *testing.T) {
	t.Run("test add tenant", testAddTenant)
	t.Run("test update tenant", testUpdateTenant)
	t.Run("test get tenant", testGetTenant)
	t.Run("test list tenant", testListTenant)
	t.Run("test add tenant business", testAddTenantBusiness)
	t.Run("test list tenant business", testListTenantBusiness)
	t.Run("test delete tenant business", testDeleteTenantBusiness)
	t.Run("test delete tenant", testDeleteTenant)
}
