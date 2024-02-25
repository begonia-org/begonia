package biz

import (
	"context"
	"fmt"
	"path/filepath"

	api "github.com/begonia-org/begonia/api/v1"
	dp "github.com/begonia-org/dynamic-proto"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/routers"
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
}

func NewEndpointUsecase(repo EndpointRepo) *EndpointUsecase {
	return &EndpointUsecase{repo: repo}
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
func (u *EndpointUsecase) AddEndpoint(ctx context.Context, endpoints []*api.Endpoints) error {
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
	routersList:=routers.Get()
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
