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
	glc "github.com/begonia-org/go-layered-cache"
	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
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

		patch := gomonkey.ApplyFuncReturn((*userRepoImpl).List, nil, fmt.Errorf("list error"))
		defer patch.Reset()
		_, err = operator.GetAllForbiddenUsers(context.Background())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "list error")
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn((*userRepoImpl).List, nil, gorm.ErrRecordNotFound)
		defer patch2.Reset()
		val, err := operator.GetAllForbiddenUsers(context.Background())
		c.So(err, c.ShouldBeNil)
		c.So(len(val), c.ShouldEqual, 0)
		patch2.Reset()

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
		err = lock.Lock(context.Background())
		c.So(err, c.ShouldBeNil)
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
		_ = lock.UnLock(context.TODO())
		c.So(isOK, c.ShouldBeTrue)
		patch := gomonkey.ApplyFuncReturn((*redislock.Client).Obtain, nil, fmt.Errorf("lock error"))
		defer patch.Reset()
		err = lock.Lock(context.TODO())
		c.So(err, c.ShouldNotBeNil)
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

		patch := gomonkey.ApplyFuncReturn((*appRepoImpl).List, nil, fmt.Errorf("list app error"))
		defer patch.Reset()
		_, err = operator.GetAllApps(context.Background())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "list app error")
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn((*appRepoImpl).List, nil, gorm.ErrRecordNotFound)
		defer patch2.Reset()
		val, err := operator.GetAllApps(context.Background())
		c.So(err, c.ShouldBeNil)
		c.So(len(val), c.ShouldEqual, 0)
		patch2.Reset()

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

		patch := gomonkey.ApplyFuncReturn((*redis.StringCmd).Int64, int64(0), redis.Nil)
		defer patch.Reset()
		tm, err := operator.LastUpdated(context.Background(), "test:user:blacklist")
		c.So(err, c.ShouldBeNil)
		c.So(tm.IsZero(), c.ShouldBeTrue)
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn((*redis.StringCmd).Int64, int64(0), fmt.Errorf("get last updated error"))
		defer patch2.Reset()
		_, err = operator.LastUpdated(context.Background(), "test:user:blacklist")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get last updated error")
		patch2.Reset()

	})
}
func testWatcher(t *testing.T) {
	c.Convey("test watcher", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		updated := make(chan string,1)
		deleted := make(chan string,1)
		ctx,cancel:=context.WithCancel(context.Background())
		defer cancel()
		defer close(updated)
		defer close(deleted)
		go func(ctx context.Context) {
			err := operator.Watcher(ctx, "/test/watcher/user/info", func(ctx context.Context, op mvccpb.Event_EventType, key, value string) error {
				if op == mvccpb.PUT {
					updated<-key
				} else if op == mvccpb.DELETE {
					deleted<-key
				}
				return nil
			})
			if err != nil {
				t.Errorf("watcher error:%v", err)
			}
		}(ctx)
		snk, _ := tiga.NewSnowflake(1)
		uid := snk.GenerateIDString()
		err := operator.(*dataOperatorRepo).data.etcd.Put(context.Background(), fmt.Sprintf("/test/watcher/user/info/%s", uid), fmt.Sprintf("user-%s", time.Now().Format("20060102150405")))
		c.So(err, c.ShouldBeNil)
		up:=<-updated
		c.So(up, c.ShouldEqual, fmt.Sprintf("/test/watcher/user/info/%s", uid))

		err = operator.(*dataOperatorRepo).data.etcd.Delete(context.Background(), fmt.Sprintf("/test/watcher/user/info/%s", uid))
		c.So(err, c.ShouldBeNil)
		d:=<-deleted
		c.So(d, c.ShouldEqual, fmt.Sprintf("/test/watcher/user/info/%s", uid))

		cancel()
		
	})

}
func testLoadRemoteCache(t *testing.T) {
	c.Convey("test load remote cache", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		cnf := config.NewConfig(conf)
		rdb := NewRDB(cfg.ReadConfig(env))
		key := fmt.Sprintf("%s:%s", cnf.GetKeyValuePrefix(), time.Now().Format("20060102150405"))
		_ = rdb.GetClient().Set(context.Background(), key, key, 10*time.Second)
		filterKey := fmt.Sprintf("%s:%s", cnf.GetFilterPrefix(), time.Now().Format("20060102150405"))
		rdb.GetClient().CFAdd(context.Background(), filterKey, filterKey)
		operator := NewOperator(cfg.ReadConfig(env), gateway.Log)
		// operator.(*dataOperatorRepo).LoadRemoteCache(context.Background())
		operator.(*dataOperatorRepo).Sync(2 * time.Second)
		time.Sleep(4 * time.Second)
		val, err := operator.(*dataOperatorRepo).local.Get(context.Background(), key)
		c.So(err, c.ShouldBeNil)
		c.So(val, c.ShouldNotBeNil)
		patch := gomonkey.ApplyFuncReturn((*glc.LayeredKeyValueCacheImpl).LoadDump, fmt.Errorf("error"))
		defer patch.Reset()
		err = operator.(*dataOperatorRepo).LoadRemoteCache(context.Background())
		c.So(err, c.ShouldNotBeNil)
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn((*glc.LayeredCuckooFilterImpl).LoadDump, fmt.Errorf("loaddump error"))
		defer patch2.Reset()
		err = operator.(*dataOperatorRepo).LoadRemoteCache(context.Background())
		c.So(err, c.ShouldNotBeNil)
		patch2.Reset()
		filterErrChan := make(chan error, 1)
		filterErrChan <- fmt.Errorf("filter watch error")

		kvErrChan := make(chan error, 1)
		kvErrChan <- fmt.Errorf("kv watch error")
		patch3 := gomonkey.ApplyFuncReturn((*glc.LayeredCuckooFilterImpl).Watch, filterErrChan)
		defer patch3.Reset()
		operator.(*dataOperatorRepo).onStartOperator()
		time.Sleep(2 * time.Second)
		patch3.Reset()

		patch4 := gomonkey.ApplyFuncReturn((*glc.LayeredKeyValueCacheImpl).Watch, kvErrChan)
		defer patch4.Reset()
		operator2 := NewOperator(cfg.ReadConfig(env), gateway.Log)

		operator2.(*dataOperatorRepo).onStartOperator()
		time.Sleep(2 * time.Second)
		patch4.Reset()
		// val, err = operator.(*dataOperatorRepo).local.GetFromLocal(context.Background(), filterKey)

	})
}
func TestOperator(t *testing.T) {
	t.Run("testGetAllForbiddenUsers", testGetAllForbiddenUsers)
	t.Run("testFlashUsersCache", testFlashUsersCache)
	t.Run("testGetAllApp", testGetAllApp)
	t.Run("testFlashAppsCache", testFlashAppsCache)
	t.Run("testLastUpdated", testLastUpdated)
	t.Run("testLoadRemoteCache", testLoadRemoteCache)
	t.Run("testWatcher", testWatcher)
}
