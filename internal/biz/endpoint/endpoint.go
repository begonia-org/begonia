package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
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
	repo   EndpointRepo
	config *config.Config
	file   *file.FileUsecase
	snk    *tiga.Snowflake
}

func NewEndpointUsecase(repo EndpointRepo, file *file.FileUsecase, config *config.Config) *EndpointUsecase {
	snk, _ := tiga.NewSnowflake(1)
	return &EndpointUsecase{repo: repo, file: file, config: config, snk: snk}
}

func (e *EndpointUsecase) AddConfig(ctx context.Context, srvConfig *api.EndpointSrvConfig) (string, error) {
	if !loadbalance.CheckBalanceType(srvConfig.Balance) {
		return "", gosdk.NewError(errors.ErrUnknownLoadBalancer, int32(api.EndpointSvrStatus_NOT_SUPPORT_BALANCE), codes.InvalidArgument, "balance_type")
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
	return id, nil

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
	return updated_at, nil
}

func (u *EndpointUsecase) Delete(ctx context.Context, uniqueKey string) error {
	return u.repo.Del(ctx, uniqueKey)
}

func (u *EndpointUsecase) Get(ctx context.Context, uniqueKey string) (*api.Endpoints, error) {
	detailsKey := u.config.GetServiceKey(uniqueKey)
	value, err := u.repo.Get(ctx, detailsKey)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("%s:%w", errors.ErrEndpointNotExists.Error(), err), int32(common.Code_NOT_FOUND), codes.NotFound, "get_endpoint")

	}
	if value == "" {
		return nil, gosdk.NewError(errors.ErrEndpointNotExists, int32(common.Code_NOT_FOUND), codes.NotFound, "get_endpoint")
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
		log.Printf("list tags:%v", in.Tags)
		ks, err := u.repo.GetKeysByTags(ctx, in.Tags)
		if err != nil {
			return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_keys_by_tags")
		}
		keys = append(keys, ks...)

	}
	if len(in.UniqueKeys) > 0 {
		keys = append(keys, in.UniqueKeys...)
	}
	return u.repo.List(ctx, keys)
}
