package gateway

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/transport"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc/codes"
)

func newEndpoint(lb loadbalance.BalanceType, endpoints []*api.EndpointMeta) ([]loadbalance.Endpoint, error) {
	eps := make([]loadbalance.Endpoint, 0)
	gw := gateway.Get()

	opts := gw.GetOptions()
	for _, ep := range endpoints {
		pool := transport.NewGrpcConnPool(ep.GetAddr(), opts.PoolOptions...)
		eps = append(eps, transport.NewGrpcEndpoint(ep.GetAddr(), pool))
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

func deleteAll(ctx context.Context, pd transport.ProtobufDescription) error {
	routersList := routers.Get()
	routersList.DeleteRouters(pd)
	gw := gateway.Get()
	gw.DeleteLoadBalance(pd)
	err := gateway.Get().DeleteHandlerClient(ctx, pd)
	return err
}

func getDescriptorSet(config *config.Config, key string, value []byte) (transport.ProtobufDescription, error) {
	key = getEndpointId(config, key)
	outDir := config.GetGatewayDescriptionOut()
	pd, err := transport.NewDescriptionFromBinary(value, filepath.Join(outDir, key))
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_description_from_binary")
	}
	return pd, nil
}
func getEndpointId(config *config.Config, key string) string {
	prefix := config.GetEndpointsPrefix()
	key = strings.TrimPrefix(key, prefix)
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimSuffix(key, "/descriptor_set")
	return key
}

func getServiceKey(config *config.Config, id string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/service/%s", prefix, id)
}
func getServiceNameKey(config *config.Config, name string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/service_name/%s", prefix, name)
}
func getDetailsKey(config *config.Config, id string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/details/%s", prefix, id)
}
func getTagsKey(config *config.Config, tag, id string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/tags/%s/%s", prefix, tag, id)

}
