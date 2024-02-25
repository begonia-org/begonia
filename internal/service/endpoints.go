package service

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/web"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type EndpointsService struct {
	biz    *biz.EndpointUsecase
	log    *logrus.Logger
	config *config.Config
	api.UnimplementedEndpointServiceServer
}

func NewEndpointsService(biz *biz.EndpointUsecase, log *logrus.Logger, config *config.Config) *EndpointsService {
	return &EndpointsService{biz: biz, log: log, config: config}
}
func (e *EndpointsService) Create(ctx context.Context, in *api.EndpointRequest) (*common.APIResponse, error) {
	err := e.biz.AddEndpoint(ctx, in.Endpoints)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(nil, nil)
}

func (e *EndpointsService) Update(ctx context.Context, in *api.EndpointRequest) (*common.APIResponse, error) {
	err := e.biz.UpdateEndpoint(ctx, in.Endpoints)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(nil, nil)
}
func (e *EndpointsService) Delete(ctx context.Context, in *api.EndpointRequest) (*common.APIResponse, error) {
	err := e.biz.DeleteEndpoint(ctx, in.Endpoints)
	if err != nil {
		return web.MakeResponse(nil, err)
	}
	return web.MakeResponse(nil, nil)
}

func (e *EndpointsService) List(ctx context.Context, in *api.EndpointRequest) (*common.APIResponse, error) {
	return web.MakeResponse(nil, nil)
}
func (e *EndpointsService) Desc() *grpc.ServiceDesc {
	return &api.EndpointService_ServiceDesc
}

// service EndpointService{
//     rpc Create(EndpointRequest) returns (.begonia.org.begonia.common.api.v1.APIResponse){
//         option (google.api.http) = {
//             post: "/api/v1/endpoint/create"
//             body: "*"
//           };

//     };
//     rpc Update(EndpointRequest) returns (.begonia.org.begonia.common.api.v1.APIResponse){
//         option (google.api.http) = {
//             post: "/api/v1/endpoint/update"
//             body: "*"
//           };
//     };
//     rpc Delete(EndpointRequest) returns (.begonia.org.begonia.common.api.v1.APIResponse){
//         option (google.api.http) = {
//             post: "/api/v1/endpoint/delete"
//             body: "*"
//           };
//     };
//     rpc List(EndpointRequest) returns (.begonia.org.begonia.common.api.v1.APIResponse){
//         option (google.api.http) = {
//             post: "/api/v1/endpoint/list"
//             body: "*"
//           };
//     };
// }
