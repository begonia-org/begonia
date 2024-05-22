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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var endpointId = ""
var tag = fmt.Sprintf("test-endpoint-data-%s", time.Now().Format("20060102150405"))
var tag3 = fmt.Sprintf("test3-%s", time.Now().Format("20060102150405"))

func putTest(t *testing.T) {
	c.Convey("test app add success", t, func() {
		t.Log("add test")
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata", "helloworld.pb")
		pb, _ := os.ReadFile(pbFile)
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, gateway.Log)
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
			Tags:      []string{tag, tag3},
			Version:   fmt.Sprintf("%d", time.Now().UnixMilli()),
			CreatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
			UpdatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		})
		c.So(err, c.ShouldBeNil)

		patch := gomonkey.ApplyFuncReturn((*Data).PutEtcdWithTxn, false, fmt.Errorf("put endpoint fail"))
		defer patch.Reset()
		err = repo.Put(context.Background(), &api.Endpoints{
			Key:           snk.GenerateIDString(),
			DescriptorSet: pb,
			Name:          "test1",
			ServiceName:   "test3",
			Description:   "test4",
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
			Tags:      []string{tag, tag3},
			Version:   fmt.Sprintf("%d", time.Now().UnixMilli()),
			CreatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
			UpdatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "put endpoint fail")
		patch.Reset()

	})
}
func getEndpointTest(t *testing.T) {
	c.Convey("test endpoint get success", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, gateway.Log)
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
		repo := NewEndpointRepo(conf, gateway.Log)
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldNotBeEmpty)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		c.So(keys, c.ShouldContain, endpointKey)

		patch := gomonkey.ApplyFuncReturn((tiga.EtcdDao).BatchGet, nil, fmt.Errorf("get keys by tags fail"))
		defer patch.Reset()
		_, err = repo.GetKeysByTags(context.Background(), []string{tag})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get keys by tags fail")
		patch.Reset()
	})
}
func testList(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata", "helloworld.pb")
	pb, _ := os.ReadFile(pbFile)
	conf := cfg.ReadConfig(env)
	repo := NewEndpointRepo(conf, gateway.Log)
	snk, _ := tiga.NewSnowflake(1)
	enps := make([]string, 0)
	c.Convey("test list", t, func() {
		for i := 0; i < 10; i++ {
			epd := snk.GenerateIDString()
			enps = append(enps, epd)
			err := repo.Put(context.Background(), &api.Endpoints{
				Key:           epd,
				DescriptorSet: pb,
				Name:          fmt.Sprintf("test-data-%d-%s", i, time.Now().Format("20060102150405")),
				ServiceName:   fmt.Sprintf("test-data-%d-%s", i, time.Now().Format("20060102150405")),
				Description:   fmt.Sprintf("test-data-%d-%s", i, time.Now().Format("20060102150405")),
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
				Tags:      []string{fmt.Sprintf("test-list-%d-data-%s", i, time.Now().Format("20060102150405"))},
				Version:   fmt.Sprintf("%d", time.Now().UnixMilli()),
				CreatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
				UpdatedAt: timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
			})
			c.So(err, c.ShouldBeNil)
		}
		data, err := repo.List(context.Background(), enps)
		c.So(err, c.ShouldBeNil)
		c.So(len(data), c.ShouldEqual, 10)

		data, err = repo.List(context.Background(), []string{"not-exist"})
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldBeEmpty)
		data, err = repo.List(context.Background(), nil)
		c.So(err, c.ShouldBeNil)
		c.So(len(data), c.ShouldBeGreaterThan, 0)

		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  (tiga.EtcdDao).BatchGet,
				output: []interface{}{nil, fmt.Errorf("get keys by tags fail")},
				err:    fmt.Errorf("get keys by tags fail"),
			},
			{
				patch:  json.Unmarshal,
				output: []interface{}{fmt.Errorf("unmarshal fail")},
				err:    fmt.Errorf("unmarshal fail"),
			},
			{
				patch:  (tiga.EtcdDao).GetWithPrefix,
				output: []interface{}{nil, fmt.Errorf("get keys by tags fail")},
				err:    fmt.Errorf("get keys by tags fail"),
			},
		}
		for index, caseV := range cases {
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			if index == 2 {
				enps = []string{}
			}
			_, err = repo.List(context.Background(), enps)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
			patch.Reset()

		}
	})
}
func patchEndpointTest(t *testing.T) {
	c.Convey("test endpoint patch success", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, gateway.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		tag1 := fmt.Sprintf("test-data-patch-%s", time.Now().Format("20060102150405"))
		err := repo.Patch(context.Background(), endpointId, map[string]interface{}{
			"description": "test description",
			"balance":     string(goloadbalancer.WRRBalanceType),
			"tags":        []string{tag1, tag3},
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
		c.So(updated.Tags, c.ShouldNotContain, tag)
		c.So(updated.Tags, c.ShouldContain, tag1)
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldBeEmpty)
		keys, err = repo.GetKeysByTags(context.Background(), []string{tag1})
		c.So(err, c.ShouldBeNil)
		t.Logf("keys:%v", keys)
		c.So(keys, c.ShouldContain, endpointKey)
		invalidPatch := map[string]interface{}{
			"description": "test description",
			"balance":     string(goloadbalancer.WRRBalanceType),
			"tags":        tag1,
		}
		invalidPatchData, _ := json.Marshal(invalidPatch)
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{{
			patch:  (*endpointRepoImpl).Get,
			output: []interface{}{nil, fmt.Errorf("get endpoint fail")},
			err:    fmt.Errorf("get endpoint fail"),
		},
			{
				patch:  json.Unmarshal,
				output: []interface{}{fmt.Errorf("unmarshal fail")},
				err:    fmt.Errorf("unmarshal fail"),
			},
			{
				patch:  json.Marshal,
				output: []interface{}{nil, fmt.Errorf("marshal fail")},
				err:    fmt.Errorf("marshal fail"),
			},
			{
				patch:  (*Data).PutEtcdWithTxn,
				output: []interface{}{false, fmt.Errorf("put endpoint fail")},
				err:    fmt.Errorf("put endpoint fail"),
			},
			{
				patch:  (*endpointRepoImpl).Get,
				output: []interface{}{string(invalidPatchData), nil},
				err:    fmt.Errorf("tags type error"),
			},
		}
		for index, caseV := range cases {
			t.Logf("index:%d", index)
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			err = repo.Patch(context.Background(), endpointId, map[string]interface{}{
				"description": "test description",
				"balance":     string(goloadbalancer.WRRBalanceType),
				"tags":        []string{tag1, tag3},
			})
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
			patch.Reset()
		}

		err = repo.Patch(context.Background(), endpointId, map[string]interface{}{
			"description": "test description",
			"balance":     string(goloadbalancer.WRRBalanceType),
			"tags":        tag1,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "tags type error")
	})
}
func delEndpointTest(t *testing.T) {
	c.Convey("test endpoint delete", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, gateway.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)

		// c.So(keys, c.ShouldContain, endpointKey)

		patch := gomonkey.ApplyFuncReturn((tiga.EtcdDao).GetWithPrefix, nil, fmt.Errorf("del GetWithPrefix fail"))

		defer patch.Reset()
		err := repo.Del(context.Background(), endpointId)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "del GetWithPrefix fail")
		patch.Reset()

		patch2:=gomonkey.ApplyFuncReturn((*Data).PutEtcdWithTxn, false, fmt.Errorf("del PutEtcdWithTxn fail"))
		defer patch2.Reset()
		err = repo.Del(context.Background(), endpointId)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "del PutEtcdWithTxn fail")
		patch2.Reset()

		err = repo.Del(context.Background(), endpointId)
		c.So(err, c.ShouldBeNil)
		data, err := repo.Get(context.Background(), endpointKey)
		// t.Logf("err:%v", err)
		c.So(err, c.ShouldBeNil)
		c.So(data, c.ShouldBeEmpty)

		// Tags should be del
		keys, err := repo.GetKeysByTags(context.Background(), []string{tag3})
		c.So(err, c.ShouldBeNil)
		c.So(keys, c.ShouldBeEmpty)


	})
}
func putTagsTest(t *testing.T) {
	c.Convey("test endpoint add tags", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		repo := NewEndpointRepo(conf, gateway.Log)
		cnf := config.NewConfig(conf)
		endpointKey := cnf.GetServiceKey(endpointId)
		tag1 := fmt.Sprintf("test1-data-%s", time.Now().Format("20060102150405"))
		tag2 := fmt.Sprintf("test2-%s", time.Now().Format("20060102150405"))
		err := repo.PutTags(context.Background(), endpointId, []string{tag1, tag2, tag3})
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

		patch := gomonkey.ApplyFuncReturn((*Data).PutEtcdWithTxn, false, fmt.Errorf("put tags fail"))
		defer patch.Reset()
		err = repo.PutTags(context.Background(), endpointId, []string{tag1, tag2, tag3})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "put tags fail")

		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn(json.Marshal, nil, fmt.Errorf("marshal tags fail"))
		defer patch2.Reset()

		err = repo.PutTags(context.Background(), endpointId, []string{tag1, tag2, tag3})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "marshal tags fail")
		patch2.Reset()

		patch3 := gomonkey.ApplyFuncReturn((*endpointRepoImpl).Get,"", fmt.Errorf("get endpoint fail"))
		defer patch3.Reset()
		err = repo.PutTags(context.Background(), endpointId, []string{tag1, tag2, tag3})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get endpoint fail")
		patch3.Reset()

		patch4 := gomonkey.ApplyFuncReturn(json.Unmarshal, fmt.Errorf("unmarshal fail"))
		defer patch4.Reset()
		err = repo.PutTags(context.Background(), endpointId, []string{tag1, tag2, tag3})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "unmarshal fail")
		patch4.Reset()

	})
}

func TestEndpoint(t *testing.T) {
	t.Run("put", putTest)
	t.Run("get", getEndpointTest)
	t.Run("getKeysByTags", getKeysByTagsTest)
	t.Run("list", testList)
	t.Run("patch", patchEndpointTest)
	t.Run("putTags", putTagsTest)
	t.Run("del", delEndpointTest)
}
