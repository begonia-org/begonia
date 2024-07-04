package data

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var tid = ""
var tn = ""
var tsn=""

func testAddTenant(t *testing.T) {
	c.Convey("test add tenant", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(2)
		tid = snk.GenerateIDString()
		tn = fmt.Sprintf("test-%s", tid)
		tenant := &api.Tenants{
			TenantId:    tid,
			TenantName:  fmt.Sprintf("test-%s", tid),
			Description: "test tenant",
			Tags:        []string{"test"},
			Email: fmt.Sprintf("%s@example.com",tn),
		}
		err := bs.Add(context.Background(), tenant)
		c.So(err, c.ShouldBeNil)
	})
}
func testUpdateTenant(t *testing.T) {
	c.Convey("test update tenant", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		tenant, err := bs.Get(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		tenant.Description = "update description"
		tenant.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"description"}}
		err = bs.Patch(context.Background(), tenant)
		c.So(err, c.ShouldBeNil)
	})
}
func testGetTenant(t *testing.T) {
	c.Convey("test get tenant", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		tenant, err := bs.Get(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		c.So(tenant.TenantName, c.ShouldEqual, tn)

		tenant, err = bs.Get(context.Background(), tn)
		c.So(err, c.ShouldBeNil)
		c.So(tenant.TenantId, c.ShouldEqual, tid)

		_, err = bs.Get(context.Background(), "not found")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}

func testListTenant(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	bs := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
	c.Convey("test list tenant", t, func() {

		tenantes, err := bs.List(context.Background(), []string{"test", "test2"},[]api.TENANTS_STATUS{api.TENANTS_STATUS_TENANTS_ACTIVE}, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(len(tenantes), c.ShouldBeGreaterThan, 0)
	})
	c.Convey("test list tenant fail", t, func() {
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination, fmt.Errorf("pagination error"))
		defer patch.Reset()
		_, err := bs.List(context.Background(), []string{"not found"},[]api.TENANTS_STATUS{api.TENANTS_STATUS_TENANTS_DELETED}, 1, 10)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "pagination error")
	})
}
func testAddTenantBusiness(t *testing.T){
	c.Convey("test add tenant",t,func(){
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		tr := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bs:=NewBusinessRepo(cfg.ReadConfig(env),gateway.Log)
		snk, _ := tiga.NewSnowflake(2)
		business:=&api.Business{
			BusinessId:snk.GenerateIDString(),
			BusinessName:fmt.Sprintf("test-data-%s",snk.GenerateIDString()),
			Description:"test business",
		}
		tsn = business.BusinessName
		err:=bs.Add(context.Background(),business)
		c.So(err,c.ShouldBeNil)
		tenantBusiness := &api.TenantsBusiness{
			TenantId: tid,
			BusinessId: business.BusinessId,
			BusinessName: business.BusinessName,
			TenantName: tn,
			Plan: "Free",
		}
		err = tr.AddBusiness(context.Background(),tenantBusiness)
		c.So(err, c.ShouldBeNil)
		tb:=&api.TenantsBusiness{
			TenantId: snk.GenerateIDString(),
			BusinessId: business.BusinessId,
			BusinessName: business.BusinessName,
			TenantName: tn,
			Plan: "Free",
		}
		err = tr.AddBusiness(context.Background(),tb)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(),c.ShouldContainSubstring,"not found")

		tb2:=&api.TenantsBusiness{
			TenantId: tid,
			BusinessId: snk.GenerateIDString(),
			BusinessName: business.BusinessName,
			TenantName: tn,
			Plan: "Free",
		}
		err = tr.AddBusiness(context.Background(),tb2)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(),c.ShouldContainSubstring,"not found")
	})
}
func testGetTenantBusiness(t *testing.T){
	c.Convey("test get tenant business",t,func(){
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		snk,_:=tiga.NewSnowflake(2)
		tr := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		tb,err:=tr.GetTenantBusiness(context.Background(),tid,tsn)
		c.So(err,c.ShouldBeNil)
	
		c.So(tb.BusinessName,c.ShouldEqual,tsn)
		_,err=tr.GetTenantBusiness(context.Background(),snk.GenerateIDString(),snk.GenerateIDString())
		c.So(err,c.ShouldNotBeNil)
		c.So(err.Error(),c.ShouldContainSubstring,"not found")

	})
}
func testTenantBusinessList(t *testing.T){
	c.Convey("test tenant business list",t,func(){
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		tr := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		tbs,err:=tr.TenantBusinessList(context.Background(),tid,1,10)
		c.So(err,c.ShouldBeNil)
		c.So(len(tbs),c.ShouldBeGreaterThan,0)
		patch :=gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination,fmt.Errorf("pagination error"))
		defer patch.Reset()
		_,err=tr.TenantBusinessList(context.Background(),"not found",1,10)
		c.So(err,c.ShouldNotBeNil)
		c.So(err.Error(),c.ShouldContainSubstring,"pagination error")
	})

}
func testDelTenantBusiness(t *testing.T){
	c.Convey("test del tenant business",t,func(){
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		tr := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		err:=tr.DelTenantBusiness(context.Background(),tid,tsn)
		c.So(err,c.ShouldBeNil)
	})

}
func testDelTenant(t *testing.T) {
	c.Convey("test del tenant", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		err := bs.Del(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		err = bs.Del(context.Background(), tn)
		c.So(err, c.ShouldBeNil)

		err = bs.Del(context.Background(), "not found")
		c.So(err, c.ShouldBeNil)
		// c.So(err.Error(), c.ShouldContainSubstring, "not found")

	})
}
func TestTenant(t *testing.T) {
	t.Run("test add tenant", testAddTenant)
	t.Run("test update tenant", testUpdateTenant)
	t.Run("test get tenant", testGetTenant)
	t.Run("test list tenant", testListTenant)
	t.Run("test add tenant business",testAddTenantBusiness)
	t.Run("test get tenant business",testGetTenantBusiness)
	t.Run("test tenant business list",testTenantBusinessList)
	t.Run("test del tenant business",testDelTenantBusiness)
	t.Run("test del tenant", testDelTenant)

}
