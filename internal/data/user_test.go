package data

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var uid = ""
var uid2 = ""
var uid3 = ""
var user1 = fmt.Sprintf("user1-%s", time.Now().Format("20060102150405"))
var user2 = fmt.Sprintf("user2-%s", time.Now().Format("20060102150405"))

// var user3 = fmt.Sprintf("user3-%s", time.Now().Format("20060102150405"))
func testAddUser(t *testing.T) {
	c.Convey("test user add success", t, func() {
		t.Log("add test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), logger.Log)
		snk, _ := tiga.NewSnowflake(1)
		uid = snk.GenerateIDString()
		err := repo.Add(context.TODO(), &api.Users{
			Uid:       uid,
			Name:      user1,
			Dept:      "dev",
			Email:     fmt.Sprintf("%s%s@example.com", uid, time.Now().Format("20060102150405")),
			Phone:     time.Now().Format("20060102150405"),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		})
		c.So(err, c.ShouldBeNil)
		time.Sleep(2 * time.Second)
		uid2 = snk.GenerateIDString()
		err = repo.Add(context.TODO(), &api.Users{
			Uid:       uid2,
			Name:      user2,
			Dept:      "dev",
			Email:     fmt.Sprintf("user2%s@example.com", time.Now().Format("20060102150405")),
			Phone:     time.Now().Format("20060102150405"),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		})
		c.So(err, c.ShouldBeNil)
		uid3 = snk.GenerateIDString()
		err = repo.Add(context.TODO(), &api.Users{
			Uid:       uid3,
			Name:      user2,
			Dept:      "dev",
			Email:     fmt.Sprintf("user2%s@example.com", time.Now().Format("20060102150405")),
			Phone:     time.Now().Format("20060102150405"),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")

	})
}
func testGetUser(t *testing.T) {
	c.Convey("test user get success", t, func() {
		t.Log("get test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewUserRepo(conf, logger.Log)
		user, err := repo.Get(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		c.So(user, c.ShouldNotBeNil)
		c.So(user.Uid, c.ShouldEqual, uid)
		c.So(user.Name, c.ShouldEqual, user1)

		c.Convey("test user get by account", func() {
			cnf := config.NewConfig(conf)
			key, iv := cnf.GetAesConfig()
			account, err := tiga.EncryptAES([]byte(key), user.Name, iv)
			t.Log("account:", account)
			c.So(err, c.ShouldBeNil)
			userByName, err := repo.Get(context.TODO(), account)
			c.So(err, c.ShouldBeNil)
			c.So(userByName, c.ShouldNotBeNil)
			c.So(userByName.Uid, c.ShouldEqual, uid)
			c.So(userByName.Name, c.ShouldEqual, user1)
		})
	})
}

func testUpdateUser(t *testing.T) {
	c.Convey("test user update success", t, func() {
		t.Log("update test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), logger.Log)
		user, err := repo.Get(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		c.So(user, c.ShouldNotBeNil)
		oldPhone := user.Phone
		oldName := user.Name
		lastedUpdateAt := user.UpdatedAt
		createdAt := user.CreatedAt
		user.Name = fmt.Sprintf("user-update-%s", time.Now().Format("20060102150405"))
		time.Sleep(1 * time.Second)
		user.Phone = time.Now().Format("20060102150405")
		user.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"name", "phone"}}
		err = repo.Patch(context.TODO(), user)

		c.So(err, c.ShouldBeNil)
		updated, err := repo.Get(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		c.So(updated, c.ShouldNotBeNil)
		c.So(updated.Uid, c.ShouldEqual, uid)
		c.So(updated.Name, c.ShouldNotEqual, oldName)
		c.So(updated.Phone, c.ShouldNotEqual, oldPhone)
		c.So(updated.UpdatedAt.Seconds, c.ShouldBeGreaterThanOrEqualTo, lastedUpdateAt.Seconds)
		c.So(updated.CreatedAt.Seconds, c.ShouldEqual, createdAt.Seconds)
		updated.Uid = uid2
		updated.ID = 0
		updated.Name = fmt.Sprintf("user4-update-%s", time.Now().Format("20060102150405"))
		updated.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"name", "phone", "email"}}
		err = repo.Patch(context.TODO(), updated)
		c.So(err, c.ShouldNotBeNil)

		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")

	})
}

func testDelUser(t *testing.T) {
	c.Convey("test user delete success", t, func() {
		t.Log("delete test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), logger.Log)
		err := repo.Del(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		_, err = repo.Get(context.TODO(), uid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "record not found")
	})
}

func testListUser(t *testing.T) {
	c.Convey("test user list success", t, func() {
		t.Log("list test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), logger.Log)
		snk, _ := tiga.NewSnowflake(1)
		// rand.Seed(time.Now().UnixNano())
		rand:=rand.New(rand.NewSource(time.Now().UnixNano()))
		status := []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE, api.USER_STATUS_LOCKED}
		depts := [3]string{"dev", "test", "prd"}
		for i := 0; i < 20; i++ {
			err := repo.Add(context.TODO(), &api.Users{
				Uid:       snk.GenerateIDString(),
				Name:      fmt.Sprintf("user-%d-%s@example.com", i, time.Now().Format("20060102150405")),
				Dept:      depts[rand.Intn(len(depts))],
				Email:     fmt.Sprintf("user-%d-%s@example.com", i, time.Now().Format("20060102150405")),
				Phone:     fmt.Sprintf("%d%s", i, time.Now().Format("20060102150405")),
				Role:      api.Role_ADMIN,
				Avatar:    "https://www.example.com/avatar.jpg",
				Owner:     "test-user-01",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
				Status:    status[rand.Intn(len(status))],
			})
			if err != nil {
				t.Errorf("add user error:%v", err)
			}
		}
		users1, err := repo.List(context.TODO(), []string{"dev", "test"}, []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE}, 1, 5)
		c.So(err, c.ShouldBeNil)
		c.So(users1, c.ShouldNotBeEmpty)
		users2, err := repo.List(context.TODO(), []string{"dev", "test"}, []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE}, 2, 5)
		c.So(err, c.ShouldBeNil)
		c.So(users2, c.ShouldNotBeEmpty)
		c.So(users1[0].Uid, c.ShouldNotEqual, users2[0].Uid)
		c.So(users1[4].Uid, c.ShouldBeLessThan, users2[0].Uid)

	})
}
func TestUser(t *testing.T) {
	t.Run("testAddUser", testAddUser)
	t.Run("testGetUser", testGetUser)
	t.Run("testUpdateUser", testUpdateUser)
	t.Run("testDelUser", testDelUser)
	t.Run("testListUser", testListUser)
}
