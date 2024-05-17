package gateway

import (
	"fmt"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
)

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}
func JSONCamelCase(s string) string {
	var b []byte
	var wasUnderscore bool
	for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
		c := s[i]
		if c != '_' {
			if wasUnderscore && isASCIILower(c) {
				c -= 'a' - 'A' // convert to uppercase
			}
			b = append(b, c)
		}
		wasUnderscore = c == '_'
	}
	return string(b)
}

func NewLoadBalanceEndpoint(lb loadbalance.BalanceType, endpoints []*api.EndpointMeta) ([]loadbalance.Endpoint, error) {
	eps := make([]loadbalance.Endpoint, 0)
	gw := Get()

	opts := gw.GetOptions()
	for _, ep := range endpoints {
		pool := NewGrpcConnPool(ep.GetAddr(), opts.PoolOptions...)
		eps = append(eps, NewGrpcEndpoint(ep.GetAddr(), pool))
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
