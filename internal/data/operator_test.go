package data

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	appApi "github.com/begonia-org/go-sdk/api/app/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var users []*api.Users
var apps []*appApi.Apps

func testGetAllForbiddenUsers(t *testing.T) {
	c.Convey("test get all forbidden users", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewUserRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(1)
		// rand.Seed(time.Now().UnixNano())
		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		status := []api.USER_STATUS{api.USER_STATUS_ACTIVE, api.USER_STATUS_INACTIVE, api.USER_STATUS_LOCKED, api.USER_STATUS_DELETED}
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
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		var err error
		users, err = operator.GetAllForbiddenUsers(context.Background())
		c.So(err, c.ShouldBeNil)
		c.So(len(users), c.ShouldNotEqual, 0)
		hasActive := false
		for _, user := range users {
			if user.Status == api.USER_STATUS_ACTIVE {
				hasActive = true
				break
			}
		}
		c.So(hasActive, c.ShouldBeFalse)
	})
}
func testFlashUsersCache(t *testing.T) {
	c.Convey("test flash users cache", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		// operator.(*dataOperatorRepo).local.OnStart()
		operator.OnStart()
		lock, err := operator.Locker(context.Background(), "test:user:blacklist:lock", 3*time.Second)
		c.So(err, c.ShouldBeNil)
		defer func() {
			_ = lock.UnLock(context.TODO())
		}()
		err = operator.FlashUsersCache(context.Background(), "test:user:blacklist", users, 10*time.Second)
		c.So(err, c.ShouldBeNil)
		isOK := true
		time.Sleep(1 * time.Second)
		for _, user := range users {
			usersCache, err := operator.(*dataOperatorRepo).local.Get(context.Background(), fmt.Sprintf("test:user:blacklist:%s", user.Uid))
			if err != nil || usersCache == nil {
				t.Errorf("get user cache error:%v", err)
				isOK = false
				break
			}
			item := &api.Users{}
			err = protojson.Unmarshal(usersCache, item)
			if err != nil || item.Uid != user.Uid {
				t.Errorf("unmarshal user cache error:%v", err)
				isOK = false
				break
			}
		}
		c.So(isOK, c.ShouldBeTrue)
	})
}
func testGetAllApp(t *testing.T) {
	c.Convey("test get all app", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		repo := NewAppRepo(cfg.ReadConfig(env), gateway.Log)
		snk, _ := tiga.NewSnowflake(1)
		for i := 0; i < 10; i++ {
			appAccess, _ := generateRandomString(32)
			appSecret, _ := generateRandomString(64)
			err := repo.Add(context.TODO(), &appApi.Apps{
				Appid:       snk.GenerateIDString(),
				AccessKey:   appAccess,
				Secret:      appSecret,
				Status:      appApi.APPStatus_APP_ENABLED,
				IsDeleted:   false,
				Name:        fmt.Sprintf("app-%d-%s", i, time.Now().Format("20060102150405")),
				Description: "test",
				CreatedAt:   timestamppb.New(time.Now()),
				UpdatedAt:   timestamppb.New(time.Now()),
			})

			c.So(err, c.ShouldBeNil)
		}
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		var err error
		apps, err = operator.GetAllApps(context.Background())
		c.So(err, c.ShouldBeNil)
		c.So(len(apps), c.ShouldNotEqual, 0)

	})
}
func testFlashAppsCache(t *testing.T) {
	c.Convey("test flash apps cache", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		err := operator.FlashAppsCache(context.Background(), "test:app:blacklist", apps, 10*time.Second)
		c.So(err, c.ShouldBeNil)
		isOK := true
		for _, app := range apps {
			appCache, err := operator.(*dataOperatorRepo).local.Get(context.Background(), fmt.Sprintf("test:app:blacklist:access_key:%s", app.AccessKey))
			if err != nil || appCache == nil {
				t.Errorf("get app cache error:%v", err)
				isOK = false
				break
			}

		}
		c.So(isOK, c.ShouldBeTrue)
	})
}
func testLastUpdated(t *testing.T) {
	c.Convey("test last updated", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)

		t, err := operator.LastUpdated(context.Background(), "test:user:blacklist")
		c.So(err, c.ShouldBeNil)
		c.So(t, c.ShouldHappenOnOrBefore, time.Now())
	})
}
func testWatcher(t *testing.T) {
	c.Convey("test watcher", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		updated := ""
		deleted := ""
		go func() {
			err := operator.Watcher(context.Background(), "/test/user/info", func(ctx context.Context, op mvccpb.Event_EventType, key, value string) error {
				if op == mvccpb.PUT {
					updated = key
				} else if op == mvccpb.DELETE {
					deleted = key
				}
				return nil
			})
			if err != nil {
				t.Errorf("watcher error:%v", err)
			}
		}()
		snk, _ := tiga.NewSnowflake(1)
		uid := snk.GenerateIDString()
		err := operator.(*dataOperatorRepo).data.etcd.Put(context.Background(), fmt.Sprintf("/test/user/info/%s", uid), fmt.Sprintf("user-%s", time.Now().Format("20060102150405")))
		c.So(err, c.ShouldBeNil)
		time.Sleep(2 * time.Second)

		err = operator.(*dataOperatorRepo).data.etcd.Delete(context.Background(), fmt.Sprintf("/test/user/info/%s", uid))
		c.So(err, c.ShouldBeNil)
		time.Sleep(3 * time.Second)
		c.So(updated, c.ShouldEqual, fmt.Sprintf("/test/user/info/%s", uid))
		c.So(deleted, c.ShouldEqual, fmt.Sprintf("/test/user/info/%s", uid))
	})
}
func TestOperator(t *testing.T) {
	t.Run("testGetAllForbiddenUsers", testGetAllForbiddenUsers)
	t.Run("testFlashUsersCache", testFlashUsersCache)
	t.Run("testGetAllApp", testGetAllApp)
	t.Run("testFlashAppsCache", testFlashAppsCache)
	t.Run("testLastUpdated", testLastUpdated)
	t.Run("testWatcher", testWatcher)
}
