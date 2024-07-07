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
	api "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var bid = ""
var bn = ""

func testAddBusiness(t *testing.T) {
	c.Convey("test add business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(2)

		bs := biz.NewBusinessUsecase(repo)
		bn = fmt.Sprintf("test-%s", bid)
		in := &api.PostBusinessRequest{
			BusinessName: fmt.Sprintf("test-%s", snk.GenerateIDString()),
			Description:  "test",
			Tags:         []string{"test"},
		}
		business, err := bs.Add(context.Background(), in, snk.GenerateIDString())
		c.So(err, c.ShouldBeNil)
		bid = business.BusinessId
		bn = business.BusinessName
		_, err = bs.Add(context.Background(), in, snk.GenerateIDString())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")
		in2 := &api.PostBusinessRequest{
			BusinessName: fmt.Sprintf("test-1-%s", bid),
		}

		patch := gomonkey.ApplyMethodReturn(repo, "Add", fmt.Errorf("too long"))
		defer patch.Reset()
		_, err = bs.Add(context.Background(), in2, snk.GenerateIDString())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "too long")
	})
}

func testGetBusiness(t *testing.T) {
	c.Convey("test get business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		bs := biz.NewBusinessUsecase(repo)
		business, err := bs.Get(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		c.So(business, c.ShouldNotBeNil)
		c.So(business.BusinessName, c.ShouldEqual, bn)
	})
}
func testUpdateBusiness(t *testing.T) {
	c.Convey("test update business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		bs := biz.NewBusinessUsecase(repo)
		business, err := bs.Get(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		c.So(business, c.ShouldNotBeNil)
		business.Description = "test update"
		business.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"description"}}
		err = bs.Patch(context.Background(), business)
		c.So(err, c.ShouldBeNil)
		business, err = bs.Get(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		c.So(business, c.ShouldNotBeNil)
		c.So(business.Description, c.ShouldEqual, "test update")

		snk, _ := tiga.NewSnowflake(2)
		in := &api.PostBusinessRequest{
			BusinessName: fmt.Sprintf("test-%s", snk.GenerateIDString()),
			Description:  "test",
			Tags:         []string{"test"},
		}
		_, err = bs.Add(context.Background(), in, snk.GenerateIDString())
		c.So(err, c.ShouldBeNil)
		bn2 := in.BusinessName
		business.BusinessName = bn2
		business.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"business_name"}}
		err = bs.Patch(context.Background(), business)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")

		business.BusinessName = bn
		business.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"business_name"}}
		business.BusinessId = snk.GenerateIDString()
		err = bs.Patch(context.Background(), business)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

	})

}
func testListBusiness(t *testing.T) {
	c.Convey("test list business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		bs := biz.NewBusinessUsecase(repo)
		business, err := bs.List(context.Background(), []string{"test"}, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(business, c.ShouldNotBeEmpty)

		patch := gomonkey.ApplyMethodReturn(repo, "List", nil, fmt.Errorf("list error"))
		defer patch.Reset()
		_, err = bs.List(context.Background(), []string{"test"}, 1, 10)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "list error")

	})
}
func testDeleteBusiness(t *testing.T) {
	c.Convey("test delete business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := data.NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		bs := biz.NewBusinessUsecase(repo)
		err := bs.Del(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		_, err = bs.Get(context.Background(), bid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}

func TestBusinessBiz(t *testing.T) {
	t.Run("test add business", testAddBusiness)
	t.Run("test get business", testGetBusiness)
	t.Run("test update business", testUpdateBusiness)
	t.Run("test list business", testListBusiness)
	t.Run("test delete business", testDeleteBusiness)
}
