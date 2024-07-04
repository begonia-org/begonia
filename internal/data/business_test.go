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

var bid = ""
var bn = ""

func testAddBusiness(t *testing.T) {
	c.Convey("test add business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(2)
		bid = snk.GenerateIDString()
		bn = fmt.Sprintf("test-%s", bid)
		business := &api.Business{
			BusinessId:   bid,
			BusinessName: fmt.Sprintf("test-%s", bid),
			Description:  "test",
			Tags:         []string{"test"},
		}
		err := bs.Add(context.Background(), business)
		c.So(err, c.ShouldBeNil)
	})
}
func testUpdateBusiness(t *testing.T) {
	c.Convey("test update business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		business, err := bs.Get(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		business.Description = "update description"
		business.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"description"}}
		err = bs.Patch(context.Background(), business)
		c.So(err, c.ShouldBeNil)
	})
}
func testGetBusiness(t *testing.T) {
	c.Convey("test get business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		business, err := bs.Get(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		c.So(business.BusinessName, c.ShouldEqual, bn)

		business, err = bs.Get(context.Background(), bn)
		c.So(err, c.ShouldBeNil)
		c.So(business.BusinessId, c.ShouldEqual, bid)

		_, err = bs.Get(context.Background(), "not found")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found business")
	})
}

func testListBusiness(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	bs := NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
	c.Convey("test list business", t, func() {

		businesses, err := bs.List(context.Background(), []string{"test","test2"}, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(len(businesses), c.ShouldBeGreaterThan, 0)
	})
	c.Convey("test list business fail", t, func() {
		patch:=gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination,fmt.Errorf("pagination error"))
		defer patch.Reset()
		_, err := bs.List(context.Background(), []string{"not found"}, 1, 10)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "pagination error")
	})
}
func testDelBusiness(t *testing.T) {
	c.Convey("test del business", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		bs := NewBusinessRepo(cfg.ReadConfig(env), gateway.Log)
		err := bs.Del(context.Background(), bid)
		c.So(err, c.ShouldBeNil)
		err = bs.Del(context.Background(), bn)
		c.So(err,c.ShouldNotBeNil)
	})
}
func TestBusiness(t *testing.T){
	t.Run("test add business",testAddBusiness)
	t.Run("test update business",testUpdateBusiness)
	t.Run("test get business",testGetBusiness)
	t.Run("test list business",testListBusiness)
	t.Run("test del business",testDelBusiness)

}
