package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
	glc "github.com/begonia-org/go-layered-cache"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

func TestGetFromLocalErr(t *testing.T) {
	c.Convey("test get from local err", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cache := NewLayered(cfg.ReadConfig(env), gateway.Log)
		_, err := cache.GetFromLocal(context.Background(), fmt.Sprintf("test-local:%s", time.Now().Format("20060102150405")))
		c.So(err, c.ShouldNotBeNil)

		patch := gomonkey.ApplyFuncReturn((*glc.LayeredKeyValueCacheImpl).GetFromLocal, nil, nil)
		defer patch.Reset()
		_, err = cache.GetFromLocal(context.Background(), fmt.Sprintf("test-local:%s", time.Now().Format("20060102150405")))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "local cache value is not found")

		patch.Reset()
		fk := fmt.Sprintf("test-local:%s", time.Now().Format("20060102150405"))
		err = cache.AddToFilter(context.Background(), fk, []byte(fk))
		c.So(err, c.ShouldBeNil)
		ok, err := cache.CheckInFilter(context.Background(), fk, []byte(fk))
		c.So(err, c.ShouldBeNil)
		c.So(ok, c.ShouldBeTrue)
		err = cache.DelInFilter(context.Background(), fk, []byte(fk))
		c.So(err, c.ShouldBeNil)

		ok, err = cache.CheckInFilter(context.Background(), fk, []byte(fk))
		c.So(err, c.ShouldBeNil)
		c.So(ok, c.ShouldBeFalse)
		// patch:=gomonkey.ApplyFuncReturn((*l), []byte(fk), nil)

	})
}
func TestCacheWatch(t *testing.T) {
	c.Convey("test cache watch", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cache := NewLayered(cfg.ReadConfig(env), gateway.Log)
		go cache.Watch(context.Background())
		conf := cfg.ReadConfig(env)
		cnf := config.NewConfig(conf)
		key := fmt.Sprintf("%s:%s", cnf.GetKeyValuePrefix(), time.Now().Format("20060102150405"))
		// _ = rdb.GetClient().Set(context.Background(), key, key, 10*time.Second)
		cache2 := newCache(tiga.NewRedisDao(conf), cnf, gateway.Log)

		err := cache2.Set(context.Background(), key, []byte(key), 10*time.Second)
		c.So(err, c.ShouldBeNil)
		time.Sleep(2 * time.Second)
		val, err := cache.GetFromLocal(context.Background(), key)
		c.So(err, c.ShouldBeNil)
		c.So(string(val), c.ShouldEqual, key)

	})
}
