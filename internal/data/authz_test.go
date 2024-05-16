package data

import (
	"context"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/transport"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

var token = ""

func testCacheToken(t *testing.T) {

	c.Convey("test cache token", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAuthzRepo(cfg.ReadConfig(env), transport.Log)
		snk, _ := tiga.NewSnowflake(1)
		token = snk.GenerateIDString()
		err := repo.CacheToken(context.TODO(), "test:token", token, 5*time.Second)
		c.So(err, c.ShouldBeNil)
	})
}
func testGetToken(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	repo := NewAuthzRepo(cfg.ReadConfig(env), transport.Log)
	c.Convey("test get token", t, func() {

		tk := repo.GetToken(context.TODO(), "test:token")
		c.So(tk, c.ShouldNotBeEmpty)
		c.So(tk, c.ShouldEqual, token)
	})
	c.Convey("test token expiration", t, func() {
		time.Sleep(7 * time.Second)
		tk := repo.GetToken(context.TODO(), "test:token")
		c.So(tk, c.ShouldBeEmpty)
	})
}

func deleteToken(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	repo := NewAuthzRepo(cfg.ReadConfig(env), transport.Log)
	c.Convey("test delete exp token", t, func() {
		err := repo.DelToken(context.TODO(), "test:token")
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("test token delete", t, func() {
		snk, _ := tiga.NewSnowflake(1)
		token = snk.GenerateIDString()
		err := repo.CacheToken(context.TODO(), "test:token2", token, 5*time.Second)
		c.So(err, c.ShouldBeNil)
		err = repo.DelToken(context.TODO(), "test:token2")
		c.So(err, c.ShouldBeNil)
		tk := repo.GetToken(context.TODO(), "test:token2")
		c.So(tk, c.ShouldBeEmpty)
	})
}
func testPutBlacklist(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	repo := NewAuthzRepo(cfg.ReadConfig(env), transport.Log)
	c.Convey("test put blacklist", t, func() {
		err := repo.PutBlackList(context.TODO(), token)
		c.So(err, c.ShouldBeNil)
	})
}
func testCheckInBlackList(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	repo := NewAuthzRepo(cfg.ReadConfig(env), transport.Log)
	c.Convey("test check in blacklist", t, func() {

		b, err := repo.CheckInBlackList(context.TODO(), token)
		c.So(err, c.ShouldBeNil)
		c.So(b, c.ShouldBeTrue)
		snk, _ := tiga.NewSnowflake(1)

		b, err = repo.CheckInBlackList(context.TODO(), snk.GenerateIDString())
		c.So(err, c.ShouldBeNil)
		c.So(b, c.ShouldBeFalse)
	})
}
func TestAuthzRepo(t *testing.T) {
	t.Run("testCacheToken", testCacheToken)
	t.Run("testGetToken", testGetToken)
	t.Run("deleteToken", deleteToken)
	t.Run("testPutBlacklist", testPutBlacklist)
	t.Run("testCheckInBlackList", testCheckInBlackList)
}
