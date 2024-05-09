package gateway

import (
	"context"
	"fmt"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"encoding/json"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc/codes"
)

type GatewayWatcher struct {
	config *config.Config
	repo   EndpointRepo
}

// update
//
// Created or Update endpoint from etcd data
// It will delete all old endpoint and register new endpoint
// and then new endpoint will be registered to gateway
func (g *GatewayWatcher) Update(ctx context.Context, key string, value string) error {
	endpoint := &api.Endpoints{}
	err := json.Unmarshal([]byte(value), endpoint)
	if err != nil {
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_endpoint")
	}
	pd, err := getDescriptorSet(g.config, key, endpoint.DescriptorSet)
	if err != nil {
		logger.Log.Errorf("get descriptor set error: %s", err.Error())
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_descriptor_set")
	}
	err = deleteAll(ctx, pd)
	if err != nil {
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "delete_descriptor")
	}
	eps, err := newEndpoint(loadbalance.BalanceType(endpoint.Balance), endpoint.GetEndpoints())
	if err != nil {
		return errors.New(errors.ErrUnknownLoadBalancer, int32(api.EndpointSvrStatus_NOT_SUPPORT_BALANCE), codes.InvalidArgument, "new_endpoint")
	}
	lb, err := loadbalance.New(loadbalance.BalanceType(endpoint.Balance), eps)
	if err != nil {
		return errors.New(fmt.Errorf("new loadbalance error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_loadbalance")
	}
	gw := gateway.Get()
	err = gw.RegisterService(ctx, pd, lb)
	if err != nil {
		return errors.New(fmt.Errorf("register service error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "register_service")
	}
	// err = g.repo.PutTags(ctx, endpoint.Key, endpoint.Tags)
	return nil
}
func (g *GatewayWatcher) del(ctx context.Context, key string, value string) error {
	endpoint := &api.Endpoints{}
	err := json.Unmarshal([]byte(value), endpoint)
	if err != nil {
		return errors.New(err, int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "unmarshal_endpoint")
	}
	pd, err := getDescriptorSet(g.config, key, endpoint.DescriptorSet)
	if err != nil {
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_descriptor_set")
	}
	err = deleteAll(ctx, pd)
	if err != nil {
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "delete_descriptor")
	}
	return nil
}

func (g *GatewayWatcher) Handle(ctx context.Context, op mvccpb.Event_EventType, key, value string) error {
	switch op {
	case mvccpb.PUT:
		return g.Update(ctx, key, value)
	case mvccpb.DELETE:
		return g.del(ctx, key, value)
	default:
		return errors.New(fmt.Errorf("unknown operation"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "unknown_operation")
	}
}

func NewWatcher(config *config.Config, repo EndpointRepo) *GatewayWatcher {
	return &GatewayWatcher{
		config: config,
		repo:   repo,
	}
}
