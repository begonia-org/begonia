package data

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slices"
)

type endpointRepoImpl struct {
	data *Data
	cfg  *config.Config
}

func NewEndpointRepoImpl(data *Data, cfg *config.Config) endpoint.EndpointRepo {
	return &endpointRepoImpl{data: data, cfg: cfg}
}

func (r *endpointRepoImpl) Get(ctx context.Context, key string) (string, error) {
	return r.data.etcd.GetString(ctx, key)
}
func (e *endpointRepoImpl) Del(ctx context.Context, id string) error {
	srvKey := e.cfg.GetServiceKey(id)
	ops := make([]clientv3.Op, 0)
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
	if err != nil || !ok {
		return fmt.Errorf("delete endpoint fail: %w", err)

	}
	return nil
}
func (e *endpointRepoImpl) Put(ctx context.Context, endpoint *api.Endpoints) error {
	ops := make([]clientv3.Op, 0)
	srvKey := e.cfg.GetServiceKey(endpoint.Key)
	for _, tag := range endpoint.Tags {
		tagKey := e.cfg.GetTagsKey(tag, endpoint.Key)
		ops = append(ops, clientv3.OpPut(tagKey, srvKey))
	}

	details, _ := json.Marshal(endpoint)
	ops = append(ops, clientv3.OpPut(srvKey, string(details)))
	ok, err := e.data.PutEtcdWithTxn(ctx, ops)
	if err != nil || !ok {
		// log.Printf("put endpoint fail: %s", err.Error())
		return fmt.Errorf("put endpoint fail: %w", err)
	}
	return nil
}
func (e *endpointRepoImpl) List(ctx context.Context, keys []string) ([]*api.Endpoints, error) {
	if len(keys) > 0 {
		ops := make([]clientv3.Op, 0)
		prefix := e.cfg.GetServicePrefix()

		for _, key := range keys {
			if !strings.HasPrefix(key, prefix) {
				key = e.cfg.GetServiceKey(key)
			}
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
func (e *endpointRepoImpl) patchTags(oldTags []interface{}, newTags []interface{}, id string) []clientv3.Op {
	ops := make([]clientv3.Op, 0)
	for _, tag := range oldTags {
		if val, ok := tag.(string); ok {
			// Del old tags if not in new tags
			if !slices.Contains(newTags, tag) {
				tagKey := e.cfg.GetTagsKey(val, id)
				ops = append(ops, clientv3.OpDelete(tagKey))
			}
		}

	}
	for _, tag := range newTags {
		if val, ok := tag.(string); ok {
			tagKey := e.cfg.GetTagsKey(val, id)
			ops = append(ops, clientv3.OpPut(tagKey, e.cfg.GetServiceKey(id)))
		}

	}
	return ops
}
func (e *endpointRepoImpl) getTags(v interface{}) ([]interface{}, error) {
	tags := make([]interface{}, 0)
	if val, ok := v.([]interface{}); ok {
		tags = val
	} else if val, ok := v.([]string); ok {
		for _, tag := range val {
			tags = append(tags, tag)
		}
	} else {
		return nil, fmt.Errorf("tags type error")
	}
	return tags, nil
}
func (e *endpointRepoImpl) Patch(ctx context.Context, id string, patch map[string]interface{}) error {
	origin, err := e.Get(ctx, e.cfg.GetServiceKey(id))
	if err != nil || origin == "" {
		return fmt.Errorf("get old endpoint error: %w or %w", err, errors.ErrEndpointNotExists)
	}
	originConfig := make(map[string]interface{})
	err = json.Unmarshal([]byte(origin), &originConfig)
	if err != nil {
		return err
	}
	ops := make([]clientv3.Op, 0)
	// 更新tags
	if tags, ok := patch["tags"]; ok && tags != nil {
		oldTags, err := e.getTags(originConfig["tags"])
		if err != nil {
			return fmt.Errorf("get old tags error: %w", err)
		}
		newTags, err := e.getTags(tags)
		if err != nil {
			return fmt.Errorf("get new tags error: %w", err)
		}

		ops = append(ops, e.patchTags(oldTags, newTags, id)...)

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
	if err != nil || !ok {
		return fmt.Errorf("patch endpoint fail: %w", err)
	}

	return nil
}

func (e *endpointRepoImpl) PutTags(ctx context.Context, id string, tags []string) error {
	origin, err := e.Get(ctx, e.cfg.GetServiceKey(id))
	if err != nil || origin == "" {
		return fmt.Errorf("get old endpoint error: %w or %w", err, errors.ErrEndpointNotExists)
	}
	endpoint := &api.Endpoints{}
	err = json.Unmarshal([]byte(origin), endpoint)
	if err != nil {
		return err
	}

	ops := make([]clientv3.Op, 0)
	srvKey := e.cfg.GetServiceKey(id)
	oldTags := make([]interface{}, 0)
	newTags := make([]interface{}, 0)
	for _, tag := range endpoint.Tags {
		oldTags = append(oldTags, tag)
	}
	for _, tag := range tags {
		newTags = append(newTags, tag)

	}
	ops = append(ops, e.patchTags(oldTags, newTags, id)...)
	endpoint.Tags = tags
	updated, err := json.Marshal(endpoint)
	if err != nil {
		return fmt.Errorf("marshal endpoint fail when update  tags: %w", err)
	}
	ops = append(ops, clientv3.OpPut(srvKey, string(updated)))

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
