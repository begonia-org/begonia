package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/spark-lence/tiga"
)

type EndpointRepo interface {
	// mysql
	Del(ctx context.Context, id string) error
	Get(ctx context.Context, key string) (string, error)
	List(ctx context.Context, keys []string) ([]*api.Endpoints, error)
	Put(ctx context.Context, endpoint *api.Endpoints) error
	Patch(ctx context.Context, id string, patch map[string]interface{}) error
	PutTags(ctx context.Context, id string, tags []string) error
	GetKeysByTags(ctx context.Context, tags []string) ([]string, error)
}

type EndpointUsecase struct {
	repo    EndpointRepo
	config  *config.Config
	snk     *tiga.Snowflake
	watcher *EndpointWatcher
}

func NewEndpointUsecase(repo EndpointRepo, config *config.Config) *EndpointUsecase {
	snk, _ := tiga.NewSnowflake(1)
	return &EndpointUsecase{repo: repo, config: config, snk: snk, watcher: NewWatcher(config, repo)}
}

func (e *EndpointUsecase) AddConfig(ctx context.Context, srvConfig *api.EndpointSrvConfig) (string, error) {
	if !loadbalance.CheckBalanceType(srvConfig.Balance) {
		return "", gosdk.NewError(pkg.ErrUnknownLoadBalancer, int32(api.EndpointSvrStatus_NOT_SUPPORT_BALANCE), codes.InvalidArgument, "balance_type")
	}
	id := e.snk.GenerateIDString()

	endpoint := &api.Endpoints{
		Name:          srvConfig.Name,
		Description:   srvConfig.Description,
		Tags:          srvConfig.Tags,
		Version:       fmt.Sprintf("%d", time.Now().UnixMilli()),
		CreatedAt:     timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		UpdatedAt:     timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		Key:           id,
		Endpoints:     srvConfig.Endpoints,
		Balance:       srvConfig.Balance,
		ServiceName:   srvConfig.ServiceName,
		DescriptorSet: srvConfig.DescriptorSet,
	}
	err := e.repo.Put(ctx, endpoint)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "put_endpoint")

	}
	data, _ := json.Marshal(endpoint)
	err = e.watcher.Update(ctx, id, string(data))
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "watcher_update")

	}
	return id, err

}
func (e *EndpointUsecase) Patch(ctx context.Context, srvConfig *api.EndpointSrvUpdateRequest) (string, error) {
	patch := make(map[string]interface{})
	bSrvConfig, err := json.Marshal(srvConfig)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "marshal_config")
	}
	svrConfigPatch := make(map[string]interface{})
	err = json.Unmarshal(bSrvConfig, &svrConfigPatch)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_config")

	}
	if srvConfig.UpdateMask != nil {
		// 过滤掉不允许修改的字段
		for _, field := range srvConfig.UpdateMask.Paths {
			patch[field] = svrConfigPatch[field]

		}
	}

	updated_at := timestamppb.New(time.Now()).AsTime().Format(time.RFC3339)
	patch["updated_at"] = updated_at
	err = e.repo.Patch(ctx, srvConfig.UniqueKey, patch)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "patch_config")
	}
	detailsKey := e.config.GetServiceKey(srvConfig.UniqueKey)

	newVal, err := e.repo.Get(ctx, detailsKey)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_endpoint")
	}
	err = e.watcher.Update(ctx, srvConfig.UniqueKey, newVal)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "watcher_update")
	}
	return updated_at, err
}

func (u *EndpointUsecase) Delete(ctx context.Context, uniqueKey string) error {
	detailsKey := u.config.GetServiceKey(uniqueKey)

	origin, _ := u.repo.Get(ctx, detailsKey)
	err := u.repo.Del(ctx, uniqueKey)
	if err != nil {
		return gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "del_endpoint")
	}
	err = u.watcher.Del(ctx, uniqueKey, origin)
	if err != nil {
		return gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "watcher_del")
	}
	return nil
}

func (u *EndpointUsecase) Get(ctx context.Context, uniqueKey string) (*api.Endpoints, error) {
	detailsKey := u.config.GetServiceKey(uniqueKey)
	value, err := u.repo.Get(ctx, detailsKey)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("%s:%w", pkg.ErrEndpointNotExists.Error(), err), int32(common.Code_NOT_FOUND), codes.NotFound, "get_endpoint")

	}
	if value == "" {
		return nil, gosdk.NewError(pkg.ErrEndpointNotExists, int32(common.Code_NOT_FOUND), codes.NotFound, "get_endpoint")
	}
	// log.Printf("get endpoint value:%s", value)
	endpoint := &api.Endpoints{}
	err = json.Unmarshal([]byte(value), endpoint)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_endpoint")
	}
	return endpoint, nil

}

func (u *EndpointUsecase) List(ctx context.Context, in *api.ListEndpointRequest) ([]*api.Endpoints, error) {
	keys := make([]string, 0)
	if len(in.Tags) > 0 {
		// log.Printf("list tags:%v", in.Tags)
		ks, err := u.repo.GetKeysByTags(ctx, in.Tags)
		if err != nil {
			return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_keys_by_tags")
		}
		keys = append(keys, ks...)

	}

	if len(in.UniqueKeys) > 0 {
		keys = append(keys, in.UniqueKeys...)
	}
	list := make([]*api.Endpoints, 0)
	page := 1
	pageSize := 20
	number := int(math.Ceil(float64(len(keys)) / float64(pageSize)))
	for page <= number {
		start := (page - 1) * pageSize
		end := page * pageSize
		if end > len(keys) {
			end = len(keys)
		}
		ks := keys[start:end]
		eps, err := u.repo.List(ctx, ks)
		if err != nil {
			return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "list_endpoint")
		}
		list = append(list, eps...)
		page++

	}
	// log.Printf("list keys:%v", keys)
	return list, nil
}
