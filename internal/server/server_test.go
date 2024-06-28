package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/middleware"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/service"
	gosdk "github.com/begonia-org/go-sdk"
	c "github.com/smartystreets/goconvey/convey"
)

func TestReadDescErr(t *testing.T) {
	c.Convey("TestReadDescErr", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := config.ReadConfig(env)
		cnf := cfg.NewConfig(conf)
		patch := gomonkey.ApplyFuncReturn(os.ReadFile, []byte{}, fmt.Errorf("read file error"))
		defer patch.Reset()
		_, err := readDesc(cnf)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "read file error")
		patch1 := gomonkey.ApplyFuncReturn(gateway.NewDescriptionFromBinary, nil, fmt.Errorf("new desc error"))
		defer patch1.Reset()
		_, err = readDesc(cnf)
		patch1.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "new desc error")
		desc := cnf.GetLocalAPIDesc()
		bin, _ := os.ReadFile(desc)
		pd, _ := gateway.NewDescriptionFromBinary(bin, filepath.Dir(desc))
		patch2 := gomonkey.ApplyMethodReturn(pd, "SetHttpResponse", fmt.Errorf("set http response error"))
		defer patch2.Reset()
		_, err = readDesc(cnf)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "set http response error")

		worker := func() {
			NewGateway(NewGatewayConfig(":127.0.0.1:12138"), cnf, []service.Service{&service.SysService{}}, &middleware.PluginsApply{Plugins: make(gosdk.Plugins, 0)})
		}
		c.So(worker, c.ShouldPanic)
		patch2.Reset()

		patch3 := gomonkey.ApplyFuncReturn((*gateway.GatewayServer).RegisterLocalService, fmt.Errorf("register local service error"))
		defer patch3.Reset()
		c.So(worker, c.ShouldPanicWith, fmt.Errorf("register local service error"))
		patch3.Reset()

		c.So(worker, c.ShouldNotPanic)
	})
}
