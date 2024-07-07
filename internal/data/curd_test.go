package data

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

func TestAssertDeletedModel(t *testing.T) {
	c.Convey("test assert deleted model", t, func() {
		curd := &curdImpl{db: &tiga.MySQLDao{}}
		v, ok := curd.assertDeletedModel(&struct{}{})
		c.So(ok, c.ShouldBeFalse)
		c.So(v, c.ShouldBeNil)
		v, ok = curd.assertDeletedModel(&[]struct{}{})
		c.So(ok, c.ShouldBeFalse)
		c.So(v, c.ShouldBeNil)

		v, ok = curd.assertDeletedModel(nil)
		c.So(ok, c.ShouldBeFalse)
		c.So(v, c.ShouldBeNil)
		err := curd.SetBoolean(&api.Apps{}, "is_deleted_test")
		c.So(err, c.ShouldNotBeNil)
		err = curd.SetDatetimeAt(&api.Apps{}, "deleted_at_test")
		c.So(err, c.ShouldNotBeNil)
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Begin, nil)
		defer patch.Reset()
		curd.BeginTx(context.Background())
	})
}
func TestGetPrimaryColumnValueErr(t *testing.T) {
	c.Convey("test get primary column value err", t, func() {
		_, err := getPrimaryColumnValue(make(map[string]interface{}), "primary")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not a struct type")
	})
}
