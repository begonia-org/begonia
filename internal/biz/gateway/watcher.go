package gateway

import (
	"context"
	"fmt"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	api "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/spark-lence/tiga/loadbalance"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
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
func (g *GatewayWatcher) update(ctx context.Context, key string, value string) error {
	endpoint := &api.Endpoints{}
	err := protojson.Unmarshal([]byte(value), endpoint)
	if err != nil {
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_endpoint")
	}
	pd, err := getDescriptorSet(g.config, key, endpoint.DescriptorSet)
	if err != nil {
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
	err = g.updateTags(ctx, endpoint.UniqueKey, endpoint.Tags)
	return err
}
func (g *GatewayWatcher) del(ctx context.Context, key string, value string) error {
	endpoint := &api.Endpoints{}
	err := protojson.Unmarshal([]byte(value), endpoint)
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
		return g.update(ctx, key, value)
	case mvccpb.DELETE:
		return g.del(ctx, key,value)
	default:
		return errors.New(fmt.Errorf("unknown operation"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "unknown_operation")
	}
}
func (g *GatewayWatcher) updateTags(ctx context.Context, id string, tags []string) error {
	ops := make([]clientv3.Op, 0)
	for _, tag := range tags {
		tagKey := getTagsKey(g.config, tag, id)
		srvKey := getServiceKey(g.config, id)
		ops = append(ops, clientv3.OpPut(tagKey, srvKey))
	}
	ok, err := g.repo.PutEndpoint(ctx, ops)
	if err != nil {
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "put_endpoint")

	}
	if !ok {
		return errors.New(fmt.Errorf("put config fail"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "put_endpoint")
	}
	return nil
}
func NewWatcher(config *config.Config, repo EndpointRepo) *GatewayWatcher {
	return &GatewayWatcher{
		config: config,
		repo:   repo,
	}
}
