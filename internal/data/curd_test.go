package data

import (
	"testing"

	api "github.com/begonia-org/go-sdk/api/app/v1"
	c "github.com/smartystreets/goconvey/convey"
)

func TestAssertDeletedModel(t *testing.T) {
	c.Convey("test assert deleted model", t, func() {
		curd := &curdImpl{}
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
	})
}
func TestGetPrimaryColumnValueErr(t *testing.T) {
	c.Convey("test get primary column value err", t, func() {
		_, _, err := getPrimaryColumnValue(make(map[string]interface{}), "primary")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not a struct type")

		_, _, err = getPrimaryColumnValue(&struct {
			Primary string
			Name    string
		}{
			Primary: "primary",
			Name:    "name",
		}, "primary")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found primary column")
	})
}
