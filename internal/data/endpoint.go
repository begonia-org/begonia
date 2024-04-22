package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/begonia-org/begonia/internal/biz/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type endpointRepoImpl struct {
	data *Data
	cfg  *config.Config
}

func NewEndpointRepoImpl(data *Data, cfg *config.Config) gateway.EndpointRepo {
	return &endpointRepoImpl{data: data, cfg: cfg}
}

func (r *endpointRepoImpl) AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	sources := NewSourceTypeArray(endpoints)
	return r.data.CreateInBatches(sources)
}

func (r *endpointRepoImpl) Get(ctx context.Context, key string) (string, error) {
	return r.data.etcd.GetString(ctx, key)
}
func (e *endpointRepoImpl) Del(ctx context.Context, id string) error {
	srvKey := e.cfg.GetServiceKey(id)
	ops := make([]clientv3.Op, 0)
	// details := getDetailsKey(e.cfg, id)
	// ops = append(ops, clientv3.OpDelete(details))
	ops = append(ops, clientv3.OpDelete(srvKey))
	kvs, err := e.data.etcd.GetWithPrefix(ctx, filepath.Join(e.cfg.GetEndpointsPrefix(), "tags"))
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		if string(kv.Value) == srvKey {
			ops = append(ops, clientv3.OpDelete(string(kv.Key)))
		}
	}
	ok, err := e.data.PutEtcdWithTxn(ctx, ops)
	if err != nil {
		return fmt.Errorf("delete endpoint fail: %w", err)

	}
	if !ok {
		return fmt.Errorf("delete endpoint fail")
	}
	return nil
}
func (e *endpointRepoImpl) Put(ctx context.Context, endpoint *api.Endpoints) error {
	ops := make([]clientv3.Op, 0)
	srvKey := e.cfg.GetServiceKey(endpoint.UniqueKey)
	for _, tag := range endpoint.Tags {
		tagKey := e.cfg.GetTagsKey(tag, endpoint.UniqueKey)
		ops = append(ops, clientv3.OpPut(tagKey, srvKey))
	}

	details, _ := json.Marshal(endpoint)
	ops = append(ops, clientv3.OpPut(srvKey, string(details)))
	ok, err := e.data.PutEtcdWithTxn(ctx, ops)
	if err != nil {
		log.Printf("put endpoint fail: %v", err)
		return fmt.Errorf("put endpoint fail: %w", err)
	}
	if !ok {
		return fmt.Errorf("put endpoint fail")
	}
	return nil
}
func (e *endpointRepoImpl) List(ctx context.Context, keys []string) ([]*api.Endpoints, error) {
	if len(keys) > 0 {
		ops := make([]clientv3.Op, 0)
		for _, key := range keys {
			ops = append(ops, clientv3.OpGet(key))
		}
		kvs, err := e.data.etcd.BatchGet(ctx, ops)
		if err != nil {
			return nil, err
		}
		endpoints := make([]*api.Endpoints, 0)
		for _, kv := range kvs {
			endpoint := &api.Endpoints{}
			err = json.Unmarshal(kv.Value, endpoint)
			if err != nil {
				return nil, err
			}
			endpoints = append(endpoints, endpoint)
		}
		return endpoints, nil

	}
	kvs, err := e.data.etcd.GetWithPrefix(ctx, e.cfg.GetServicePrefix())
	if err != nil {
		return nil, fmt.Errorf("get endpoints fail with prefix: %w", err)
	}
	endpoints := make([]*api.Endpoints, 0)
	ops := make([]clientv3.Op, 0)
	for _, kv := range kvs {
		ops = append(ops, clientv3.OpGet(string(kv.Value)))
	}
	kvs, err = e.data.etcd.BatchGet(ctx, ops)
	if err != nil {
		return nil, fmt.Errorf("get endpoints fail with prefix: %w", err)

	}
	for _, kv := range kvs {
		endpoint := &api.Endpoints{}
		err = json.Unmarshal(kv.Value, endpoint)
		if err != nil {
			return nil, fmt.Errorf("unmarshal endpoint fail: %w,%v", err, string(kv.Value))
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (e *endpointRepoImpl) Patch(ctx context.Context, id string, patch map[string]interface{}) error {
	origin, err := e.Get(ctx, e.cfg.GetServiceKey(id))
	if err != nil {
		return err
	}
	if origin == "" {
		return fmt.Errorf("config not found")
	}
	originConfig := make(map[string]interface{})
	err = json.Unmarshal([]byte(origin), &originConfig)
	if err != nil {
		return err
	}
	ops := make([]clientv3.Op, 0)
	// 更新tags
	if tags, ok := patch["tags"]; ok {
		// 先删除原有tags
		if originTags := originConfig["tags"]; originTags != nil {
			for _, tag := range originTags.([]string) {
				tagKey := e.cfg.GetTagsKey(tag, id)
				ops = append(ops, clientv3.OpDelete(tagKey))
			}
		}
		// 添加新tags
		for _, tag := range tags.([]string) {
			tagKey := e.cfg.GetTagsKey(tag, id)
			ops = append(ops, clientv3.OpPut(tagKey, e.cfg.GetServiceKey(id)))
		}

	}
	for k, v := range patch {
		originConfig[k] = v
	}
	newConfig, err := json.Marshal(originConfig)
	if err != nil {
		return fmt.Errorf("marshal new config error: %w", err)
	}
	ops = append(ops, clientv3.OpPut(e.cfg.GetServiceKey(id), string(newConfig)))
	ok, err := e.data.PutEtcdWithTxn(ctx, ops)
	if err != nil {
		return fmt.Errorf("patch endpoint fail: %w", err)
	}
	if !ok {
		return fmt.Errorf("patch endpoint fail")
	}
	return nil
}

func (e *endpointRepoImpl) PutTags(ctx context.Context, id string, tags []string) error {
	origin, err := e.Get(ctx, e.cfg.GetServiceKey(id))
	if err != nil {
		return err
	}
	if origin == "" {
		return fmt.Errorf("config not found")
	}
	endpoint := &api.Endpoints{}
	err = json.Unmarshal([]byte(origin), endpoint)
	if err != nil {
		return err
	}

	ops := make([]clientv3.Op, 0)
	filters := make(map[string]bool)

	// 先删除原有tags
	for _, tag := range endpoint.Tags {
		if _, ok := filters[tag]; ok {
			continue
		}
		filters[tag] = true
		tagKey := e.cfg.GetTagsKey(tag, id)
		ops = append(ops, clientv3.OpDelete(tagKey))
	}
	srvKey := e.cfg.GetServiceKey(id)
	for _, tag := range tags {
		if _, ok := filters[tag]; ok {
			continue
		}
		filters[tag] = true
		tagKey := e.cfg.GetTagsKey(tag, id)
		ops = append(ops, clientv3.OpPut(tagKey, srvKey))
	}
	ok, err := e.data.PutEtcdWithTxn(ctx, ops)
	if err != nil {
		return fmt.Errorf("put tags fail: %w", err)
	}
	if !ok {
		return fmt.Errorf("put tags fail")
	}
	return nil
}

func (e *endpointRepoImpl) GetKeysByTags(ctx context.Context, tags []string) ([]string, error) {
	ops := make([]clientv3.Op, 0)
	for _, tag := range tags {
		tagKey := e.cfg.GetTagsKey(tag, "")
		ops = append(ops, clientv3.OpGet(tagKey, clientv3.WithPrefix()))
	}
	kvs, err := e.data.etcd.BatchGet(ctx, ops)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0)
	for _, kv := range kvs {
		ids = append(ids, string(kv.Value))
	}
	return ids, nil
}
