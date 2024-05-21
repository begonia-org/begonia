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
