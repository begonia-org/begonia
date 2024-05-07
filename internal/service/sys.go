package service

import (
	"context"
	"log"

	"github.com/begonia-org/begonia"
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

func (s *SysService) Get(ctx context.Context, in *api.InfoRequest) (*api.InfoResponse, error) {
	log.Printf("Version: %v", begonia.Version)
	log.Printf("commit: %v", begonia.Commit)
	return &api.InfoResponse{
		Version:   begonia.Version,
		BuildTime: begonia.BuildTime,
		Commit:    begonia.Commit,
	}, nil
}
