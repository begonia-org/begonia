package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spark-lence/tiga/errors"
	api "github.com/wetrycode/begonia/api/v1"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/service"
)

func initGrpcSvr(ctx context.Context, mux *runtime.ServeMux, user *service.UsersService, config *config.Config) error {
	// opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// err := api.RegisterManagerServiceHandlerFromEndpoint(ctx, mux, config.GetString("rpc.scheduler"), opts)

	// if err != nil {
	// 	return errors.Wrap(err, "注册调度服务失败")
	// }
	// err = api.RegisterTaskServiceHandlerFromEndpoint(ctx, mux, config.GetString("rpc.scheduler"), opts)
	// if err != nil {
	// 	return errors.Wrap(err, "注册任务服务失败")
	// }
	err := api.RegisterAuthServiceHandlerServer(ctx, mux, user)
	if err != nil {
		return errors.Wrap(err, "注册用户服务失败")
	}
	return nil
}
