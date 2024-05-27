package routers_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	c "github.com/smartystreets/goconvey/convey"
)

func TestLoadAllRouters(t *testing.T) {
	c.Convey("TestLoadAllRouters", t, func() {
		R := routers.NewHttpURIRouteToSrvMethod()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")
		pd, err := gateway.NewDescription(pbFile)
		c.So(err, c.ShouldBeNil)
		R.LoadAllRouters(pd)
		d := R.GetRoute("/test/get")
		c.So(d, c.ShouldNotBeNil)
		c.So(d.ServiceName, c.ShouldEqual, "/INTEGRATION.TESTSERVICE/GET")
		R.AddLocalSrv("/INTEGRATION.TESTSERVICE/TEST")
		c.So(R.IsLocalSrv("/INTEGRATION.TESTSERVICE/TEST"), c.ShouldBeTrue)

		d = R.GetRouteByGrpcMethod("/INTEGRATION.TESTSERVICE/GET")
		c.So(d, c.ShouldNotBeNil)

		rs:=R.GetAllRoutes()
		c.So(len(rs), c.ShouldBeGreaterThan, 0)
		d,ok:=rs["/test/custom"]
		c.So(ok, c.ShouldBeTrue)
		c.So(d.ServiceName, c.ShouldEqual, "/INTEGRATION.TESTSERVICE/CUSTOM")

	})
}
func TestDeleteRouters(t *testing.T) {
	c.Convey("TestDeleteRouters", t, func() {
		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

		pd, err := gateway.NewDescription(pbFile)
		c.So(err, c.ShouldBeNil)
		R.LoadAllRouters(pd)
		R.DeleteRouters(pd)
		d := R.GetRoute("/test/get")
		c.So(d, c.ShouldBeNil)
	})
}
