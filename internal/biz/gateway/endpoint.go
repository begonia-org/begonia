package gateway

import (
    "context"
    "encoding/json"
    "fmt"
    "io/fs"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/begonia-org/begonia/internal/biz/file"
    "github.com/begonia-org/begonia/internal/pkg/config"
    "github.com/begonia-org/begonia/internal/pkg/errors"
    "github.com/begonia-org/begonia/transport"
    loadbalance "github.com/begonia-org/go-loadbalancer"
    api "github.com/begonia-org/go-sdk/api/endpoint/v1"
    fapi "github.com/begonia-org/go-sdk/api/file/v1"
    common "github.com/begonia-org/go-sdk/common/api/v1"
    "google.golang.org/grpc/codes"
    "google.golang.org/protobuf/types/known/timestamppb"

    "github.com/begonia-org/begonia/internal/pkg/gateway"
    "github.com/begonia-org/begonia/internal/pkg/routers"
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
    protoFile, err := e.file.Download(ctx, &fapi.DownloadRequest{Key: endpoint.ProtoPath, Version: endpoint.ProtoVersion}, author)
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
            pd, err := transport.NewDescription(path)
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
    if !loadbalance.CheckBalanceType(srvConfig.Balance) {
        return "", errors.New(errors.ErrUnknownLoadBalancer, int32(api.EndpointSvrStatus_NOT_SUPPORT_BALANCE), codes.InvalidArgument, "balance_type")
    }
    id := e.snk.GenerateIDString()

    endpoint := &api.Endpoints{
        Name:          srvConfig.Name,
        Description:   srvConfig.Description,
        Tags:          srvConfig.Tags,
        Version:       fmt.Sprintf("%d", time.Now().UnixMilli()),
        CreatedAt:     timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
        UpdatedAt:     timestamppb.New(time.Now()).AsTime().Format(time.RFC3339),
        Key:     id,
        Endpoints:     srvConfig.Endpoints,
        Balance:       srvConfig.Balance,
        ServiceName:   srvConfig.ServiceName,
        DescriptorSet: srvConfig.DescriptorSet,
    }
    log.Printf("endpoint add tags :%v", srvConfig.Tags)
    err := e.repo.Put(ctx, endpoint)
    if err != nil {
        return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "put_endpoint")

    }
    return id, nil

}
func (e *EndpointUsecase) Patch(ctx context.Context, srvConfig *api.EndpointSrvUpdateRequest) (string, error) {
    patch := make(map[string]interface{})
    bSrvConfig, err := json.Marshal(srvConfig)
    if err != nil {
        return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "marshal_config")
    }
    svrConfigPatch := make(map[string]interface{})
    err = json.Unmarshal(bSrvConfig, &svrConfigPatch)
    if err != nil {
        return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_config")

    }
    // 过滤掉不允许修改的字段
    for _, field := range srvConfig.Mask.Paths {
        patch[field] = svrConfigPatch[field]

    }

    updated_at := timestamppb.New(time.Now()).AsTime().Format(time.RFC3339)
    patch["updated_at"] = updated_at
    err = e.repo.Patch(ctx, srvConfig.UniqueKey, patch)
    if err != nil {
        return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "patch_config")
    }
    return updated_at, nil
}

func (u *EndpointUsecase) Delete(ctx context.Context, uniqueKey string) error {
    return u.repo.Del(ctx, uniqueKey)
}

func (u *EndpointUsecase) Get(ctx context.Context, uniqueKey string) (*api.Endpoints, error) {
    detailsKey := u.config.GetServiceKey(uniqueKey)
    value, err := u.repo.Get(ctx, detailsKey)
    if value == "" {
        return nil, errors.New(errors.ErrEndpointNotExists, int32(common.Code_NOT_FOUND), codes.NotFound, "get_endpoint")

    }
    if err != nil {
        return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_endpoint")
    }
    endpoint := &api.Endpoints{}
    err = json.Unmarshal([]byte(value), endpoint)
    if err != nil {
        return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "unmarshal_endpoint")
    }
    return endpoint, nil

}

func (u *EndpointUsecase) List(ctx context.Context, in *api.ListEndpointRequest) ([]*api.Endpoints, error) {
    keys := make([]string, 0)
    if len(in.Tags) > 0 {
        ks, err := u.repo.GetKeysByTags(ctx, in.Tags)
        if err != nil {
            return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_keys_by_tags")
        }
        keys = append(keys, ks...)

    }
    if len(in.UniqueKeys) > 0 {
        keys = append(keys, in.UniqueKeys...)
    }
    return u.repo.List(ctx, keys)
}
