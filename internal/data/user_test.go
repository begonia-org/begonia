package data

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
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
		repo := NewUserRepo(cfg.ReadConfig(env), gateway.Log)
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
		pipe,err := repo.Cache(context.TODO(), "test:user:cache", []*api.Users{{Uid: uid, Name: user1, Dept: "dev"}}, 3*time.Second, func(user *api.Users) ([]byte, interface{}) {
			return []byte(user.Uid), user.Uid
		})
		c.So(err, c.ShouldBeNil)
		c.So(pipe, c.ShouldNotBeNil)
		cmds, err := pipe.Exec(context.Background())
		c.So(err, c.ShouldBeNil)
		c.So(len(cmds), c.ShouldBeGreaterThan, 0)
		for _, cmd := range cmds {
			c.So(cmd.Err(), c.ShouldBeNil)
		}
		// c.So(pipe.Exec())
		val, err := repo.(*userRepoImpl).local.Get(context.Background(), fmt.Sprintf("test:user:cache:%s", uid))
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldNotBeNil)
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

		patch := gomonkey.ApplyFunc((*curdImpl).SetDatetimeAt, func(_ *curdImpl, _ biz.Model, name string) error {
			if name == "created_at" {
				return fmt.Errorf("SetDatetimeAt error")
			}
			if name == "updated_at" {
				return fmt.Errorf("SetDatetimeAt error")
			}
			return nil
		})

		defer patch.Reset()
		uid4 := snk.GenerateIDString()
		u := &api.Users{
			Uid:       uid4,
			Name:      fmt.Sprintf("user4-%s", time.Now().Format("20060102150405")),
			Dept:      "dev",
			Email:     fmt.Sprintf("user2%s@example.com", time.Now().Format("20060102150405")),
			Phone:     fmt.Sprintf("123%s", time.Now().Format("20060102150405")),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		}
		err = repo.Add(context.TODO(), u)
		patch.Reset()

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "SetDatetimeAt error")

		patch2 := gomonkey.ApplyFunc((*curdImpl).SetDatetimeAt, func(c *curdImpl, model biz.Model, name string) error {
			if name == "created_at" {
				return nil
			}
			if name == "updated_at" {
				return fmt.Errorf("updateAt error")
			}
			return nil
		})
		defer patch2.Reset()
		uid5 := snk.GenerateIDString()
		u.Uid = uid5
		err = repo.Add(context.TODO(), u)
		patch2.Reset()

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "updateAt error")

		patch3 := gomonkey.ApplyFuncReturn(tiga.EncryptStructAES, fmt.Errorf("encrypt error"))
		defer patch3.Reset()
		err = repo.Add(context.TODO(), u)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "encrypt error")

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
		repo := NewUserRepo(conf, gateway.Log)
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
		c.Convey("test user get by decrypt error", func() {
			patch := gomonkey.ApplyFuncReturn(tiga.DecryptStructAES, fmt.Errorf("decrypt error"))
			defer patch.Reset()
			_, err = repo.Get(context.TODO(), uid)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, "decrypt error")
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
		repo := NewUserRepo(cfg.ReadConfig(env), gateway.Log)
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
		repo := NewUserRepo(cfg.ReadConfig(env), gateway.Log)
		err := repo.Del(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		_, err = repo.Get(context.TODO(), uid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "record not found")

		// patch:=gomonkey.ApplyFuncReturn((*userRepoImpl).Get,nil,fmt.Errorf("record not found"))
		err = repo.Del(context.TODO(), uid)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "record not found")
	})
}

func testListUser(t *testing.T) {
	c.Convey("test user list success", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(1)
		// rand.Seed(time.Now().UnixNano())
		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		status := []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE, api.USER_STATUS_LOCKED}
		depts := [3]string{"dev", "test"}

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

		user3, err := repo.List(context.TODO(), []string{"unknown"}, []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE}, 1, 5)
		c.So(err, c.ShouldBeNil)
		c.So(len(user3), c.ShouldEqual, 0)

		patch := gomonkey.ApplyFuncReturn((*curdImpl).List, fmt.Errorf("list user error"))
		defer patch.Reset()
		_, err = repo.List(context.TODO(), []string{"dev", "test"}, []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE}, 1, 5)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "list user error")
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn(tiga.DecryptStructAES, fmt.Errorf("decrypt user error"))
		defer patch2.Reset()
		_, err = repo.List(context.TODO(), []string{"dev", "test"}, []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE}, 1, 5)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "decrypt user error")
		patch2.Reset()

	})
}
func testCacheError(t *testing.T) {
	c.Convey("test cache error", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(1)
		users := []*api.Users{
			{
				Uid:      snk.GenerateIDString(),
				Name:     fmt.Sprintf("user-cache-%s", time.Now().Format("20060102150405")),
				Dept:     "dev",
				Password: "123456",
				Phone:    fmt.Sprintf("%d%s", 1, time.Now().Format("20060102150405")),
				Role:     api.Role_ADMIN,
				Avatar:   "https://www.example.com/avatar.jpg",
			},
		}
		patch := gomonkey.ApplyFuncReturn((*LayeredCache).Set, fmt.Errorf("set cache error"))
		defer patch.Reset()
		_,err := repo.Cache(context.TODO(), "test:user:cache", users, 3*time.Second, func(user *api.Users) ([]byte, interface{}) {
			return []byte(user.Uid), user.Uid
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "set cache error")
		patch.Reset()
	})
}
func TestUser(t *testing.T) {
	t.Run("testAddUser", testAddUser)
	t.Run("testCacheError", testCacheError)
	t.Run("testGetUser", testGetUser)
	t.Run("testUpdateUser", testUpdateUser)
	t.Run("testDelUser", testDelUser)
	t.Run("testListUser", testListUser)
}
