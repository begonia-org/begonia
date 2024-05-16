package biz_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/transport"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var uid = ""
var username = ""

var uid2 = ""

func newUserBiz() *biz.UserUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewUserRepo(config, transport.Log)
	cnf := cfg.NewConfig(config)
	return biz.NewUserUsecase(repo, cnf)
}

func testPutUser(t *testing.T) {
	userBiz := newUserBiz()
	snk, _ := tiga.NewSnowflake(1)
	uid = snk.GenerateIDString()
	c.Convey("test user put success", t, func() {
		username = "user-biz-" + time.Now().Format("20060102150405")
		username2 := "user2-biz-" + time.Now().Format("20060102150405")
		uid2 = snk.GenerateIDString()
		u1 := &api.Users{
			Name:      username,
			Dept:      "dev",
			Email:     fmt.Sprintf("user-biz%s@example.com", time.Now().Format("20060102150405")),
			Phone:     time.Now().Format("20060102150405"),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		}

		err := userBiz.Add(context.TODO(), u1)
		c.So(err, c.ShouldBeNil)
		uid = u1.Uid
		time.Sleep(2 * time.Second)
		u2 := &api.Users{
			Name:      username2,
			Dept:      "dev",
			Email:     fmt.Sprintf("user2-biz%s@example.com", time.Now().Format("20060102150405")),
			Phone:     time.Now().Format("20060102150405"),
			Role:      api.Role_ADMIN,
			Avatar:    "https://www.example.com/avatar.jpg",
			Owner:     "test-user-01",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    api.USER_STATUS_ACTIVE,
		}
		err = userBiz.Add(context.TODO(), u2)
		c.So(err, c.ShouldBeNil)
		uid2 = u2.Uid

	})
	c.Convey("test user put failed", t, func() {
		uid3 := snk.GenerateIDString()
		err := userBiz.Add(context.TODO(), &api.Users{
			Uid:       uid3,
			Name:      username,
			Dept:      "dev",
			Email:     fmt.Sprintf("user-biz%s@example.com", time.Now().Format("20060102150405")),
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
	userBiz := newUserBiz()
	c.Convey("test user get success", t, func() {
		user, err := userBiz.Get(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		c.So(user.Name, c.ShouldEqual, username)

	})
	c.Convey("test user get failed", t, func() {
		_, err := userBiz.Get(context.TODO(), "not-exist")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

	})
}

func testPatchUser(t *testing.T) {
	userBiz := newUserBiz()
	c.Convey("test user patch success", t, func() {
		user, err := userBiz.Get(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
		user.Email = fmt.Sprintf("user-biz-2%s@example.com", time.Now().Format("20060102150405"))
		user.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"email"}}
		err = userBiz.Update(context.TODO(), user)
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("test user patch failed", t, func() {
		user, err := userBiz.Get(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)

		user.Uid = uid2
		user.ID = 0
		user.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"name"}}

		err = userBiz.Update(context.TODO(), user)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Duplicate entry")

		err = userBiz.Update(context.TODO(), &api.Users{
			Uid:        "not-exist",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "error in your SQL syntax")

	})
}

func testListUser(t *testing.T) {
	userBiz := newUserBiz()
	snk, _ := tiga.NewSnowflake(1)
	// rand.Seed(time.Now().UnixNano())
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	status := []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE, api.USER_STATUS_LOCKED}
	depts := [3]string{"dev", "test"}
	for i := 0; i < 20; i++ {
		err := userBiz.Add(context.TODO(), &api.Users{
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
	c.Convey("test user list success", t, func() {

		users, err := userBiz.List(context.TODO(), nil, nil, 1, 20)
		c.So(err, c.ShouldBeNil)
		c.So(len(users), c.ShouldBeGreaterThan, 0)
	})
	c.Convey("test user list failed", t, func() {
		users, err := userBiz.List(context.TODO(), []string{"unknown"}, []api.USER_STATUS{api.USER_STATUS_DELETED}, 1, 20)
		c.So(len(users), c.ShouldEqual, 0)
		c.So(err, c.ShouldBeNil)
		// c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}

func testDeleteUser(t *testing.T) {
	userBiz := newUserBiz()
	c.Convey("test user delete success", t, func() {
		err := userBiz.Delete(context.TODO(), uid)
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("test user delete failed", t, func() {
		snk, _ := tiga.NewSnowflake(1)

		err := userBiz.Delete(context.TODO(), snk.GenerateIDString())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")
	})
}

func TestUser(t *testing.T) {
	t.Run("testPutUser", testPutUser)
	t.Run("testGetUser", testGetUser)
	t.Run("testPatchUser", testPatchUser)
	t.Run("testListUser", testListUser)
	t.Run("testDeleteUser", testDeleteUser)
}
