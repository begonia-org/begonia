package gateway

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	dp "github.com/begonia-org/dynamic-proto"
	api "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/spark-lence/tiga"
	"github.com/spark-lence/tiga/loadbalance"
)

type EndpointRepo interface {
	// mysql
	AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
	DeleteEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
	UpdateEndpoint(ctx context.Context, endpoints []*api.Endpoints) error
	GetEndpoint(ctx context.Context, pluginId string) (*api.Endpoints, error)
	ListEndpoint(ctx context.Context, plugins []string) ([]*api.Endpoints, error)
	PutConfig(ctx context.Context, key string, value string) error
	PutEndpoint(ctx context.Context, ops []clientv3.Op) (bool, error)
	GetConfig(ctx context.Context, key string) (string, error)
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

func (e *EndpointUsecase) tmpProtoFile(data []byte) (string, error) {
	tempFile, err := os.CreateTemp("", "begonia-endpoint-proto-")
	if err != nil {
		return "", fmt.Errorf("Failed to create temp file:%w", err)
	}
	defer tempFile.Close()

	// 写入数据到临时文件
	if _, err := tempFile.Write(data); err != nil {
		// fmt.Println("Failed to write to temp file:", err)
		return "", fmt.Errorf("Failed to write to temp file:%w", err)
	}
	return tempFile.Name(), nil

}

func (e *EndpointUsecase) CreateEndpoint(ctx context.Context, endpoint *api.AddEndpointRequest, author string) (string, error) {
	// destDir := u.config.GetProtosDir()

	// destDir = filepath.Join(destDir, "endpoints", endpoint.GetName(), endpoint.GetVersion())
	protoFile, err := e.file.Download(ctx, &api.DownloadRequest{Key: endpoint.ProtoPath, Version: endpoint.ProtoVersion}, author)
	if err != nil {
		return "", err
	}
	tmp, err := e.tmpProtoFile(protoFile)
	if err != nil {
		return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_tmp_file")

	}
	defer os.Remove(tmp)
	// destDir := filepath.Join(u.config.GetProtosDir(), "endpoints", endpoint.GetName(), endpoint.ProtoVersion)
	destDir, err := os.MkdirTemp("", "endpoints")
	if err != nil {
		return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_tmp_dir")

	}
	defer os.RemoveAll(destDir)

	err = tiga.Decompress(tmp, destDir)
	if err != nil {
		return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "decompress_proto_file")
	}

	eps, err := newEndpoint(loadbalance.BalanceType(endpoint.Balance), endpoint.GetEndpoints())
	if err != nil {
		return "", errors.New(errors.ErrUnknownLoadBalancer, int32(api.EndpointSvrStatus_NOT_SUPPORT_BALANCE), codes.InvalidArgument, "new_endpoint")
	}
	lb, err := loadbalance.New(loadbalance.BalanceType(endpoint.Balance), eps)
	if err != nil {
		return "", errors.New(fmt.Errorf("new loadbalance error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_loadbalance")
	}
	id := e.snk.GenerateIDString()
	err = filepath.WalkDir(destDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 检查是否为目录且确保只获取一级子目录
		if d.IsDir() && filepath.Dir(path) == destDir {
			pd, err := dp.NewDescription(path)
			if err != nil {
				return errors.New(fmt.Errorf("new description error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_description")
			}
			routersList := routers.Get()
			routersList.LoadAllRouters(pd)
			err = gateway.Get().RegisterService(ctx, pd, lb)
			if err != nil {
				return errors.New(fmt.Errorf("register service error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "register_service")
			}
			id, err = e.AddConfig(ctx, &api.EndpointSrvConfig{
				Name:          endpoint.Name,
				Description:   endpoint.Description,
				Tags:          endpoint.Tags,
				DescriptorSet: pd.GetDescription(),
				Endpoints:     endpoint.Endpoints,
				Balance:       endpoint.Balance,
				ServiceName:   endpoint.ServiceName,
			})
			if err != nil {
				return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "add_config")
			}

		}

		return nil
	})

	if err != nil {
		return "", err
	}
	return id, nil
}

func (e *EndpointUsecase) AddConfig(ctx context.Context, srvConfig *api.EndpointSrvConfig) (string, error) {
	// prefix := e.config.GetEndpointsPrefix()
	exists, err := e.repo.GetConfig(ctx, getServiceNameKey(e.config, srvConfig.Name))
	if err != nil {
		return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_config")

	}
	if exists != "" {
		return "", errors.New(errors.ErrEndpointExists, int32(api.EndpointSvrStatus_SERVICE_NAME_DUPLICATE), codes.AlreadyExists, "endpoint_exists")

	}
	id := e.snk.GenerateIDString()
	srvKey := getServiceKey(e.config, id)
	endpoint := &api.Endpoints{
		Name:          srvConfig.Name,
		Description:   srvConfig.Description,
		Tags:          srvConfig.Tags,
		Version:       fmt.Sprintf("%d", time.Now().UnixMilli()),
		CreatedAt:     timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		UpdatedAt:     timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
		UniqueKey:     id,
		Endpoints:     srvConfig.Endpoints,
		Balance:       srvConfig.Balance,
		ServiceName:   srvConfig.ServiceName,
		DescriptorSet: srvConfig.DescriptorSet,
	}

	ops := make([]clientv3.Op, 0)
	// Add service key to tag list
	for _, tag := range srvConfig.Tags {
		tagKey := getTagsKey(e.config, tag, id)
		ops = append(ops, clientv3.OpPut(tagKey, srvKey))
	}
	ops = append(ops, clientv3.OpPut(getServiceNameKey(e.config, endpoint.Name), srvKey))
	details, _ := protojson.Marshal(endpoint)
	ops = append(ops, clientv3.OpPut(getDetailsKey(e.config, id), string(details)))
	ok, err := e.repo.PutEndpoint(ctx, ops)
	if err != nil {
		return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "put_endpoint")
	}
	if !ok {
		return "", errors.New(fmt.Errorf("put config fail"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "put_endpoint")
	}
	return id, nil
}
func (u *EndpointUsecase) AddEndpoints(ctx context.Context, endpoints []*api.Endpoints) error {
	pds := make([]dp.ProtobufDescription, 0)
	var err error
	gw := gateway.Get()

	defer func() {
		if err != nil {
			for _, pd := range pds {
				gw.DeleteLoadBalance(pd)
			}
		}
	}()
	routersList := routers.Get()
	for _, endpoint := range endpoints {
		destDir := u.config.GetProtosDir()

		destDir = filepath.Join(destDir, "endpoints", endpoint.GetName(), endpoint.GetVersion())
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("create dir error: %w", err)
		}
		eps, err := newEndpoint(loadbalance.BalanceType(endpoint.Balance), endpoint.GetEndpoints())
		if err != nil {
			return fmt.Errorf("new endpoint error: %w", err)
		}
		lb, err := loadbalance.New(loadbalance.BalanceType(endpoint.Balance), eps)
		if err != nil {
			return fmt.Errorf("new loadbalance error: %w", err)
		}
		pd, err := dp.NewDescription(destDir)
		pds = append(pds, pd)
		if err != nil {
			return fmt.Errorf("new description error: %w", err)
		}
		routersList.LoadAllRouters(pd)
		err = gw.RegisterService(ctx, pd, lb)
		if err != nil {
			return fmt.Errorf("register service error: %w", err)
		}

	}
	err = u.repo.AddEndpoint(ctx, endpoints)
	return err
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
