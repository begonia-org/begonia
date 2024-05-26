package pkg_test

import (
	"testing"

	"github.com/begonia-org/begonia/internal/pkg"
	c "github.com/smartystreets/goconvey/convey"
)
func TestConstant(t *testing.T) {
	t.Log("TestConstant")
	c.Convey("TestConstant", t, func() {
		c.So(pkg.ErrAPIKeyNotMatch, c.ShouldNotBeNil)
	})
}
