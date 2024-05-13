package service

import (
	"context"
	"time"

	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EndpointsService struct {
	biz    *endpoint.EndpointUsecase
	log    logger.Logger
	config *config.Config
	api.UnimplementedEndpointServiceServer
}

func NewEndpointsService(biz *endpoint.EndpointUsecase, log logger.Logger, config *config.Config) *EndpointsService {
	return &EndpointsService{biz: biz, log: log, config: config}
}


func (e *EndpointsService) Update(ctx context.Context, in *api.EndpointSrvUpdateRequest) (*api.UpdateEndpointResponse, error) {
	timestamp, err := e.biz.Patch(ctx, in)
	if err != nil {
		return nil, err

	}
	tm, _ := time.Parse(time.RFC3339, timestamp)
	return &api.UpdateEndpointResponse{UpdatedAt: timestamppb.New(tm)}, nil
}
func (e *EndpointsService) Put(ctx context.Context, in *api.EndpointSrvConfig) (*api.AddEndpointResponse, error) {
	id, err := e.biz.AddConfig(ctx, in)
	if err != nil {
		return nil, err
	}
	return &api.AddEndpointResponse{UniqueKey: id}, nil
}

func (e *EndpointsService) Desc() *grpc.ServiceDesc {
	return &api.EndpointService_ServiceDesc
}

func (e *EndpointsService) List(ctx context.Context, in *api.ListEndpointRequest) (*api.ListEndpointResponse, error) {
	endpoints, err := e.biz.List(ctx, in)
	if err != nil {
		return nil, err

	}
	return &api.ListEndpointResponse{Endpoints: endpoints}, nil
}

func (e *EndpointsService) Delete(ctx context.Context, in *api.DeleteEndpointRequest) (*api.DeleteEndpointResponse, error) {
	err := e.biz.Delete(ctx, in.UniqueKey)
	if err != nil {
		return nil, err
	}
	return &api.DeleteEndpointResponse{}, nil
}

func (e *EndpointsService) Details(ctx context.Context, in *api.DetailsEndpointRequest) (*api.DetailsEndpointResponse, error) {
	endpoint, err := e.biz.Get(ctx, in.UniqueKey)
	if err != nil {
		return nil, err
	}
	return &api.DetailsEndpointResponse{Endpoints: endpoint}, nil
}
