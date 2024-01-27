package data

import (
	"context"

	api "github.com/wetrycode/begonia/api/v1"
	"github.com/wetrycode/begonia/internal/biz"
)

type endpointRepoImpl struct {
	data *Data
}

func NewEndpointRepoImpl(data *Data) biz.EndpointRepo {
	return &endpointRepoImpl{data: data}
}

// AddEndpoint(ctx context.Context,endpoints []*api.Endpoints) error
// DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
// UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
// GetEndpoint(ctx context.Context, pluginId string) (error, []*api.Endpoints)
// ListEndpoint(ctx context.Context, plugins[]string) (error, []*api.Endpoints)
func (r *endpointRepoImpl) AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	// return r.data.CreateInBatches(endpoints)
	return nil
}

func (r *endpointRepoImpl) DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	// return r.data.BatchDelete(endpoints, &api.Endpoints{})
	return nil
}

func (r *endpointRepoImpl) UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	// return r.data.BatchUpdates(endpoints, &api.Endpoints{})
	return nil
}

func (r *endpointRepoImpl) GetEndpoint(ctx context.Context, pluginId string) (*api.Endpoints,error) {
	// var endpoint *api.Endpoints
	// err := r.data.Get(&endpoint, pluginId)
	// return err, endpoint
	return nil, nil
}

func (r *endpointRepoImpl) ListEndpoint(ctx context.Context, plugins []string) ([]*api.Endpoints,error) {
	// var endpoints []*api.Endpoints
	// err := r.data.List(&api.Endpoints{}, &endpoints, "plugin_id in (?)", plugins)
	// return err, endpoints
	return nil, nil
}