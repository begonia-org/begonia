package endpoint

// import (
// 	"fmt"
// 	"net/http"
// 	"reflect"
// 	"testing"

// 	"github.com/begonia-org/begonia/gateway"
// 	loadbalance "github.com/begonia-org/go-loadbalancer"
// 	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
// 	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
// 	c "github.com/smartystreets/goconvey/convey"
// 	"google.golang.org/grpc"
// )

// func TestNewEndpoint(t *testing.T) {
// 	opts := &gateway.GrpcServerOptions{
// 		Middlewares:     make([]gateway.GrpcProxyMiddleware, 0),
// 		Options:         make([]grpc.ServerOption, 0),
// 		PoolOptions:     make([]loadbalance.PoolOptionsBuildOption, 0),
// 		HttpMiddlewares: make([]gwRuntime.ServeMuxOption, 0),
// 		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
// 	}
// 	gwCnf := &gateway.GatewayConfig{
// 		GatewayAddr:   "127.0.0.1:1949",
// 		GrpcProxyAddr: "127.0.0.1:12148",
// 	}
// 	gateway.New(gwCnf, opts)
// 	meta := []*api.EndpointMeta{{
// 		Addr:   "127.0.0.1:12138",
// 		Weight: 0,
// 	},
// 		{
// 			Addr:   "127.0.0.1:12138",
// 			Weight: 1,
// 		},
// 		{
// 			Addr:   "127.0.0.1:12138",
// 			Weight: 2,
// 		},
// 	}
// 	cases := []struct {
// 		name               string
// 		endpoints          []*api.EndpointMeta
// 		exceptErr          error
// 		exceptEndpointType reflect.Type
// 	}{
// 		{
// 			name:               string(loadbalance.RRBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(gateway.NewGrpcEndpoint("", nil)).Elem(),
// 		},
// 		{
// 			name:               string(loadbalance.WRRBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(loadbalance.NewWRREndpointImpl(nil, 0)).Elem(),
// 		},
// 		{
// 			name:               string(loadbalance.WLCBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(loadbalance.NewWLCEndpointImpl(nil, 0)).Elem(),
// 		},
// 		{
// 			name:               string(loadbalance.ConsistentHashBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(gateway.NewGrpcEndpoint("", nil)).Elem(),
// 		},
// 		{
// 			name:               string(loadbalance.LCBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(loadbalance.NewLCEndpointImpl(nil)).Elem(),
// 		},
// 		{
// 			name:               string(loadbalance.SEDBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(loadbalance.NewSedEndpointImpl(nil, 0)).Elem(),
// 		},
// 		{
// 			name:               string(loadbalance.NQBalanceType),
// 			endpoints:          meta,
// 			exceptErr:          nil,
// 			exceptEndpointType: reflect.TypeOf(loadbalance.NewSedEndpointImpl(nil, 0)).Elem(),
// 		},
// 		{
// 			name:               "unknown",
// 			endpoints:          meta,
// 			exceptErr:          fmt.Errorf("unknown balance type"),
// 			exceptEndpointType: nil,
// 		},
// 	}
// 	c.Convey("TestNewEndpoint", t, func() {
// 		for _, testCase := range cases {
// 			enps, err := newEndpoint(loadbalance.BalanceType(testCase.name), testCase.endpoints)
// 			if testCase.exceptErr != nil {
// 				c.So(err, c.ShouldNotBeNil)
// 			} else {
// 				c.So(reflect.TypeOf(enps[0]).Elem(), c.ShouldEqual, testCase.exceptEndpointType)

// 			}
// 		}
// 	})

// }
