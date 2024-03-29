package gateway

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	dp "github.com/begonia-org/dynamic-proto"
	api "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc/codes"

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
}

type EndpointUsecase struct {
	repo   EndpointRepo
	config *config.Config
	file   *file.FileUsecase
}

func NewEndpointUsecase(repo EndpointRepo, file *file.FileUsecase, config *config.Config) *EndpointUsecase {
	return &EndpointUsecase{repo: repo, file: file, config: config}
}
func (u *EndpointUsecase) newEndpoint(lb loadbalance.BalanceType, endpoints []*api.EndpointMeta) ([]loadbalance.Endpoint, error) {
	eps := make([]loadbalance.Endpoint, 0)
	gw := gateway.Get()

	opts := gw.GetOptions()
	for _, ep := range endpoints {
		pool := dp.NewGrpcConnPool(ep.GetAddr(), opts.PoolOptions...)
		eps = append(eps, dp.NewGrpcEndpoint(ep.GetAddr(), pool))
	}
	switch lb {
	case loadbalance.RRBalanceType:
		return eps, nil
	case loadbalance.WRRBalanceType:
		wrrEndpoints := make([]loadbalance.Endpoint, 0)
		for index, ep := range eps {
			wrrEndpoints = append(wrrEndpoints, loadbalance.NewWRREndpointImpl(ep, int(endpoints[index].GetWeight())))
		}
		return wrrEndpoints, nil
	case loadbalance.ConsistentHashBalanceType:
		return eps, nil
	case loadbalance.LCBalanceType:
		lcEndpoints := make([]loadbalance.Endpoint, 0)
		for _, ep := range eps {
			lcEndpoints = append(lcEndpoints, loadbalance.NewLCEndpointImpl(ep))
		}
		return lcEndpoints, nil
	case loadbalance.SEDBalanceType:
		sedEndpoints := make([]loadbalance.Endpoint, 0)
		for index, ep := range eps {
			sedEndpoints = append(sedEndpoints, loadbalance.NewSedEndpointImpl(ep, int(endpoints[index].GetWeight())))
		}
		return sedEndpoints, nil
	case loadbalance.WLCBalanceType:
		wlcEndpoints := make([]loadbalance.Endpoint, 0)
		for index, ep := range eps {
			wlcEndpoints = append(wlcEndpoints, loadbalance.NewWLCEndpointImpl(ep, int(endpoints[index].GetWeight())))
		}
		return wlcEndpoints, nil
	case loadbalance.NQBalanceType:
		nqEndpoints := make([]loadbalance.Endpoint, 0)
		for index, ep := range eps {
			nqEndpoints = append(nqEndpoints, loadbalance.NewSedEndpointImpl(ep, int(endpoints[index].GetWeight())))
		}
		return nqEndpoints, nil
	default:
		return nil, fmt.Errorf("Unknown load balance type")

	}
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
func (u *EndpointUsecase) CreateEndpoint(ctx context.Context, endpoint *api.AddEndpointRequest, author string) (string, error) {
	// destDir := u.config.GetProtosDir()

	// destDir = filepath.Join(destDir, "endpoints", endpoint.GetName(), endpoint.GetVersion())
	protoFile, err := u.file.Download(ctx, &api.DownloadRequest{Key: endpoint.ProtoPath, Version: endpoint.ProtoVersion}, author)
	if err != nil {
		return "", err
	}
	tmp, err := u.tmpProtoFile(protoFile)
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

	eps, err := u.newEndpoint(loadbalance.BalanceType(endpoint.Balance), endpoint.GetEndpoints())
	if err != nil {
		return "", errors.New(errors.ErrUnknownLoadBalancer, int32(api.EndpointSvrStatus_NOT_SUPPORT_BALANCE), codes.InvalidArgument, "new_endpoint")
	}
	lb, err := loadbalance.New(loadbalance.BalanceType(endpoint.Balance), eps)
	if err != nil {
		return "", errors.New(fmt.Errorf("new loadbalance error: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_loadbalance")
	}
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
		}

		return nil
	})


	if err != nil {
		return "", err
	}
	return "", nil
}
func (u *EndpointUsecase) AddEndpoints(ctx context.Context, endpoints []*api.Endpoints) error {
	pds := make([]dp.ProtobufDescription, 0)
	var err error
	gw := gateway.Get()

	defer func() {
		if err != nil {
			for _, pd := range pds {
				gw.DeleteService(pd)
			}
		}
	}()
	routersList := routers.Get()
	for _, endpoint := range endpoints {
		destDir := u.config.GetProtosDir()

		destDir = filepath.Join(destDir, "endpoints", endpoint.GetName(), endpoint.GetVersion())
		eps, err := u.newEndpoint(loadbalance.BalanceType(endpoint.Balance), endpoint.GetEndpoint())
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
