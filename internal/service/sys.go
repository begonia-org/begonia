package service

import (
	"context"

	api "github.com/begonia-org/go-sdk/api/sys/v1"
	"google.golang.org/grpc"
)

type SysService struct {
	api.UnimplementedSystemServiceServer
}

func (s *SysService) Desc() *grpc.ServiceDesc {
	return &api.SystemService_ServiceDesc
}

func NewSysService() *SysService {
	return &SysService{}
}

func (s *SysService) GetInfo(ctx context.Context, in *api.InfoRequest) (*api.InfoResponse, error) {
	return &api.InfoResponse{
		Version:   "v1.0.0",
		BuildTime: "2021-01-01",
		Commit:    "xxxxxx",
	}, nil
}
