package biz

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
)

type EndpointRepo interface {
	// mysql
	AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
	DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
	UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
	GetEndpoint(ctx context.Context, pluginId string) (*api.Endpoints, error)
	ListEndpoint(ctx context.Context, plugins []string) ([]*api.Endpoints, error)
}

type EndpointUsecase struct {
	repo EndpointRepo
}

func NewEndpointUsecase(repo EndpointRepo) *EndpointUsecase {
	return &EndpointUsecase{repo: repo}
}

func (u *EndpointUsecase) AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	return u.repo.AddEndpoint(ctx, endpoints)
}

func (u *EndpointUsecase) DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	return u.repo.DeleteEndpoint(ctx, endpoints)
}

func (u *EndpointUsecase) UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
	return u.repo.UpdateEndpoint(ctx, endpoints)
}

func (u *EndpointUsecase) GetEndpoint(ctx context.Context, pluginId string) (*api.Endpoints, error) {
	return u.repo.GetEndpoint(ctx, pluginId)
}

func (u *EndpointUsecase) ListEndpoint(ctx context.Context, plugins []string) ([]*api.Endpoints, error) {
	return u.repo.ListEndpoint(ctx, plugins)
}
