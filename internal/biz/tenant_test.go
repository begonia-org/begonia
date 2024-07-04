package biz_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var tid = ""
var tn = ""
var tid2 = ""
var tn2 = ""
var tenantBusinessId = ""

func testAddTenant(t *testing.T) {
	c.Convey("test tenant add", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(3)

		bs := biz.NewBusinessUsecase(bRepo)

		tbiz := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		in := &api.PostTenantRequest{
			TenantName:  fmt.Sprintf("test-%s", snk.GenerateIDString()),
			Description: "test tenant",
			Tags:        []string{"test"},
			Email:       fmt.Sprintf("%s@example.com", snk.GenerateIDString()),
		}
		uid:=snk.GenerateIDString()
		tenant, err := tbiz.Add(context.Background(), in,uid)
		c.So(err, c.ShouldBeNil)
		c.So(tenant.TenantId, c.ShouldNotBeEmpty)
		tid = tenant.TenantId
		tn = tenant.TenantName
		_, err = tbiz.Add(context.Background(), in,uid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")
		patch := gomonkey.ApplyMethodReturn(repo, "Add", fmt.Errorf("too long"))
		defer patch.Reset()
		_, err = tbiz.Add(context.Background(), in,uid)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "too long")

		in2 := &api.PostTenantRequest{
			TenantName:  fmt.Sprintf("test-%s", snk.GenerateIDString()),
			Description: "test tenant2",
			Tags:        []string{"test"},
			Email:       fmt.Sprintf("%s@example.com", snk.GenerateIDString()),
		}
		tenant, err = tbiz.Add(context.Background(), in2,uid)
		c.So(err, c.ShouldBeNil)
		tid2 = tenant.TenantId
		tn2 = tenant.TenantName

	})
}
func testGetTenant(t *testing.T) {
	c.Convey("test tenant get", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		// snk, _ := tiga.NewSnowflake(3)

		bs := biz.NewBusinessUsecase(bRepo)

		tbiz := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		tenant, err := tbiz.Get(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		c.So(tenant.TenantId, c.ShouldEqual, tid)
		_, err = tbiz.Get(context.Background(), "not-exist")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}
func testPatchTenant(t *testing.T) {
	c.Convey("test tenant patch", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(3)

		bs := biz.NewBusinessUsecase(bRepo)

		tbiz := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		in := &api.PatchTenantRequest{
			TenantName:  fmt.Sprintf("test-%s", snk.GenerateIDString()),
			Description: "test tenant patch",
			Tags:        []string{"test"},
			Email:       fmt.Sprintf("%s@example.com", snk.GenerateIDString()),
			TenantId:    tid,
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{"description", "tags", "email"}},
		}
		_, err := tbiz.Update(context.Background(), in)
		c.So(err, c.ShouldBeNil)
		tenant, err := tbiz.Get(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		c.So(tenant.TenantName, c.ShouldEqual, tn)

		in = &api.PatchTenantRequest{
			TenantName:  fmt.Sprintf("test-%s", snk.GenerateIDString()),
			Description: "test tenant patch",
			Tags:        []string{"test"},
			Email:       fmt.Sprintf("%s@example.com", snk.GenerateIDString()),
			TenantId:    snk.GenerateIDString(),
			UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{"description", "tags", "email"}},
		}
		_, err = tbiz.Update(context.Background(), in)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

		patch := gomonkey.ApplyMethodReturn(repo, "Patch", fmt.Errorf("too long"))
		defer patch.Reset()
		in.TenantId = tid
		_, err = tbiz.Update(context.Background(), in)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "too long")
		in.TenantName = tn
		in.TenantId = tid2
		in.UpdateMask.Paths = []string{"tenant_name"}
		t.Logf("tid:%s,tid2:%s,tn:%s,tn2:%s", tid, tid2, tn, tn2)
		_, err = tbiz.Update(context.Background(), in)
		c.So(err, c.ShouldNotBeNil)
		t.Logf("update tenant err:%s", err.Error())
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")

	},
	)
}
func testListTenant(t *testing.T) {
	c.Convey("test tenant list", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		// snk, _ := tiga.NewSnowflake(3)

		bs := biz.NewBusinessUsecase(bRepo)

		tbiz := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		tenants, err := tbiz.List(context.Background(), []string{"test"}, []api.TENANTS_STATUS{api.TENANTS_STATUS_TENANTS_ACTIVE}, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(tenants, c.ShouldNotBeEmpty)
		c.So(len(tenants), c.ShouldBeGreaterThan, 0)

		patch := gomonkey.ApplyMethodReturn(repo, "List", nil, fmt.Errorf("list error"))
		defer patch.Reset()
		_, err = tbiz.List(context.Background(), []string{"test"}, []api.TENANTS_STATUS{api.TENANTS_STATUS_TENANTS_ACTIVE}, 1, 10)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "list error")

	})
}
func testDelTenant(t *testing.T) {
	c.Convey("test tenant delete", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		// snk, _ := tiga.NewSnowflake(3)

		bs := biz.NewBusinessUsecase(bRepo)

		tbiz := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		err := tbiz.Delete(context.Background(), tid)
		c.So(err, c.ShouldBeNil)
		_, err = tbiz.Get(context.Background(), tid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
		patch:=gomonkey.ApplyMethodReturn(repo,"Del",fmt.Errorf("del error"))
		defer patch.Reset()
		err = tbiz.Delete(context.Background(), tid)
		// patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "del error")

	})

}

func testAddTenantBusiness(t *testing.T) {
	c.Convey("test add tenant business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(2)
		bs := biz.NewBusinessUsecase(bRepo)
		tr := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		in := &api.PostBusinessRequest{
			BusinessName: fmt.Sprintf("test-data-%s", snk.GenerateIDString()),
			Description:  "test business",
		}
		business, err := bs.Add(context.Background(), in,snk.GenerateIDString())
		c.So(err, c.ShouldBeNil)
		tb, err := tr.AddTenantBusiness(context.Background(), tid, business.BusinessId, "FREE",uid)
		c.So(err, c.ShouldBeNil)
		tenantBusinessId = business.BusinessId
		c.So(tb.TenantId, c.ShouldEqual, tid)
		c.So(tb.BusinessId, c.ShouldEqual, business.BusinessId)

		_, err = tr.AddTenantBusiness(context.Background(), tid, snk.GenerateIDString(), "FREE",uid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found business")

		_, err = tr.AddTenantBusiness(context.Background(), snk.GenerateIDString(), business.BusinessId, "FREE",uid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found tenant")

		patch := gomonkey.ApplyMethodReturn(repo, "AddBusiness", fmt.Errorf("add error"))
		defer patch.Reset()
		_, err = tr.AddTenantBusiness(context.Background(), tid, business.BusinessId, "FREE",uid)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)

	})
}
func testGetTenantBusiness(t *testing.T) {
	c.Convey("test get tenant business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(2)
		bs := biz.NewBusinessUsecase(bRepo)
		tr := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))

		tbs, err := tr.GetTenantBusiness(context.Background(), tid, tenantBusinessId)
		c.So(err, c.ShouldBeNil)
		c.So(tbs, c.ShouldNotBeEmpty)
		c.So(tbs.BusinessId, c.ShouldEqual, tenantBusinessId)

		_, err = tr.GetTenantBusiness(context.Background(), snk.GenerateIDString(), tenantBusinessId)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}
func testListTenantBusiness(t *testing.T) {
	c.Convey("test list tenant business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(2)
		bs := biz.NewBusinessUsecase(bRepo)
		tr := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))
		tbs, err := tr.ListTenantBusiness(context.Background(), tid, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(tbs, c.ShouldNotBeEmpty)
		c.So(len(tbs), c.ShouldBeGreaterThan, 0)

		patch := gomonkey.ApplyMethodReturn(repo, "TenantBusinessList", nil, fmt.Errorf("list error"))
		defer patch.Reset()
		_, err = tr.ListTenantBusiness(context.Background(), snk.GenerateIDString(), 1, 10)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "list error")
	})
}
func testDelTenantBusiness(t *testing.T) {
	c.Convey("test del tenant business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewTenantRepo(cfg.ReadConfig(env), gateway.Log)
		bRepo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		// snk, _ := tiga.NewSnowflake(2)
		bs := biz.NewBusinessUsecase(bRepo)
		tr := biz.NewTenantUsecase(repo, bs, config.NewConfig(cfg.ReadConfig(env)))

		err := tr.DelTenantBusiness(context.Background(), tid, tenantBusinessId)
		c.So(err, c.ShouldBeNil)
		_, err = tr.GetTenantBusiness(context.Background(), tid, tenantBusinessId)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

		patch := gomonkey.ApplyMethodReturn(repo, "DelTenantBusiness", fmt.Errorf("del error"))
		defer patch.Reset()
		err = tr.DelTenantBusiness(context.Background(), tid, tenantBusinessId)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "del error")
	})
}

func TestTenant(t *testing.T) {
	t.Run("test add tenant", testAddTenant)
	t.Run("test get tenant", testGetTenant)
	t.Run("test patch tenant", testPatchTenant)
	t.Run("test list tenant", testListTenant)
	t.Run("test add tenant business", testAddTenantBusiness)
	t.Run("test get tenant business", testGetTenantBusiness)
	t.Run("test list tenant business", testListTenantBusiness)
	t.Run("test del tenant business", testDelTenantBusiness)
	t.Run("test del tenant", testDelTenant)

}
