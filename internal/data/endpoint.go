package data

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz/gateway"
	api "github.com/begonia-org/go-sdk/api/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type endpointRepoImpl struct {
	data *Data
}

func NewEndpointRepoImpl(data *Data) gateway.EndpointRepo {
	return &endpointRepoImpl{data: data}
}

// AddEndpoint(ctx context.Context,endpoints []*api.Endpoints) error
// DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
// UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
// GetEndpoint(ctx context.Context, pluginId string) (error, []*api.Endpoints)
// ListEndpoint(ctx context.Context, plugins[]string) (error, []*api.Endpoints)
func (r *endpointRepoImpl) AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	// return r.data.CreateInBatches(endpoints)
	sources := NewSourceTypeArray(endpoints)
	return r.data.CreateInBatches(sources)
}

func (r *endpointRepoImpl) DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	// return r.data.BatchDelete(endpoints, &api.Endpoints{})
	return nil
}

func (r *endpointRepoImpl) UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	// return r.data.BatchUpdates(endpoints, &api.Endpoints{})
	return nil
}

func (r *endpointRepoImpl) GetEndpoint(ctx context.Context, pluginId string) (*api.Endpoints, error) {
	// var endpoint *api.Endpoints
	// err := r.data.Get(&endpoint, pluginId)
	// return err, endpoint
	return nil, nil
}

func (r *endpointRepoImpl) ListEndpoint(ctx context.Context, plugins []string) ([]*api.Endpoints, error) {
	// var endpoints []*api.Endpoints
	// err := r.data.List(&api.Endpoints{}, &endpoints, "plugin_id in (?)", plugins)
	// return err, endpoints
	return nil, nil
}

func (r *endpointRepoImpl) PutConfig(ctx context.Context, key string, value string) (error) {
	return r.data.EtcdPut(ctx,key, value)
	// return "", nil
}
func (e *endpointRepoImpl)PutEndpoint(ctx context.Context, ops []clientv3.Op) (bool,error) {
	return e.data.PutEtcdWithTxn(ctx, ops)
}
func (e *endpointRepoImpl)GetConfig(ctx context.Context, key string) (string,error) {
	return e.data.EtcdGet(ctx, key)
}
