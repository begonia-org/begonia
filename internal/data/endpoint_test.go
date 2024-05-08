package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var endpointId = ""
var tag = fmt.Sprintf("test-%s", time.Now().Format("20060102150405"))
var tag3 = fmt.Sprintf("test3-%s", time.Now().Format("20060102150405"))
func putTest(t *testing.T) {
	c.Convey("test app add success", t, func() {
		t.Log("add test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "integration", "testdata", "helloworld.pb")
		pb, _ := os.ReadFile(pbFile)
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, logger.Log)
		snk, _ := tiga.NewSnowflake(1)
		endpointId = snk.GenerateIDString()
		err := repo.Put(context.Background(), &api.Endpoints{
			Key:           endpointId,
			DescriptorSet: pb,
			Name:          "test",
			ServiceName:   "test",
			Description:   "test",
			Balance:       string(goloadbalancer.RRBalanceType),
			Endpoints: []*api.EndpointMeta{
				{
					Addr:   "127.0.0.1:21213",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21214",
					Weight: 0,
				},
				{
					Addr:   "127.0.0.1:21215",
					Weight: 0,
				},
			},
			Tags:      []string{tag,tag3},
			Version:   fmt.Sprintf("%d", time.Now().UnixMilli()),
			CreatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
			UpdatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		})
		c.So(err, c.ShouldBeNil)
	})
}
func getEndpointTest(t *testing.T) {
	c.Convey("test endpoint get success", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, logger.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		data, err := repo.Get(context.Background(), endpointKey)
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldNotBeEmpty)
		added := &api.Endpoints{}
		err = json.Unmarshal([]byte(data), added)
		c.So(err, c.ShouldBeNil)
		c.So(added.Key, c.ShouldEqual, endpointId)

	})
}

func getKeysByTagsTest(t *testing.T) {
	c.Convey("test get keys by tags success", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, logger.Log)
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldNotBeEmpty)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		c.So(keys, c.ShouldContain, endpointKey)
	})
}
func patchEndpointTest(t *testing.T) {
	c.Convey("test endpoint patch success", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, logger.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		tag1:=fmt.Sprintf("test-patch-%s", time.Now().Format("20060102150405"))
		err := repo.Patch(context.Background(), endpointId, map[string]interface{}{
			"description": "test description",
			"balance":     string(goloadbalancer.WRRBalanceType),
			"tags":		[]string{tag1,tag3},
		})
		c.So(err, c.ShouldBeNil)

		data, err := repo.Get(context.Background(), endpointKey)
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldNotBeEmpty)
		updated := &api.Endpoints{}
		err = json.Unmarshal([]byte(data), updated)
		c.So(err, c.ShouldBeNil)
		c.So(updated.Description, c.ShouldEqual, "test description")
		c.So(updated.Balance, c.ShouldEqual, string(goloadbalancer.WRRBalanceType))
		c.So(updated.Name, c.ShouldEqual, "test")
		// Tags should be updated
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldBeEmpty)
		keys, err = repo.GetKeysByTags(context.Background(), []string{tag1})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldContain, endpointKey)

	})
}
func delEndpointTest(t *testing.T){
	c.Convey("test endpoint delete", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, logger.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		err := repo.Del(context.Background(), endpointId)
		c.So(err, c.ShouldBeNil)
		data, err := repo.Get(context.Background(), endpointKey)
		// t.Logf("err:%v", err)
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldBeEmpty)

		// Tags should be del
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag3})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldBeEmpty)
		// c.So(keys, c.ShouldContain, endpointKey)

	})
}
func putTagsTest(t *testing.T){
	c.Convey("test endpoint add tags", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, logger.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		tag1:=fmt.Sprintf("test1-%s", time.Now().Format("20060102150405"))
		tag2:=fmt.Sprintf("test2-%s", time.Now().Format("20060102150405"))
		err := repo.PutTags(context.Background(), endpointId, []string{tag1,tag2,tag3})
		c.So(err, c.ShouldBeNil)

		data, err := repo.Get(context.Background(), endpointKey)
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldNotBeEmpty)
		updated := &api.Endpoints{}
		err = json.Unmarshal([]byte(data), updated)
		c.So(err, c.ShouldBeNil)

		// Tags should be updated
		c.So(updated.Tags, c.ShouldContain, tag1)
		c.So(updated.Tags, c.ShouldContain, tag2)
		c.So(updated.Tags, c.ShouldContain, tag3)
		c.So(updated.Tags, c.ShouldNotContain, tag)

		// Tags will be covered
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldBeEmpty)
		keys, err = repo.GetKeysByTags(context.Background(), []string{tag1})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldContain, endpointKey)

	})
}


func TestEndpoint(t *testing.T) {
	t.Run("put", putTest)
	t.Run("get", getEndpointTest)
	t.Run("getKeysByTags", getKeysByTagsTest)
	t.Run("patch", patchEndpointTest)
	t.Run("putTags", putTagsTest)
	t.Run("del", delEndpointTest)
}
