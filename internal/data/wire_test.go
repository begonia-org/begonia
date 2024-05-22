package data

import (
	"testing"

	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	c "github.com/smartystreets/goconvey/convey"
)

func TestNewDataRepo(t *testing.T) {
	c.Convey("test new data repo", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		v := NewDataRepo(conf, gateway.Log)
		c.So(v, c.ShouldNotBeNil)
	})
}

func TestNewLocker(t *testing.T) {
	c.Convey("test new locker", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		v := NewLocker(conf, gateway.Log, "test-test-test", 3, 0)
		c.So(v, c.ShouldNotBeNil)
	})
}
