package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/gateway"
	dp "github.com/begonia-org/dynamic-proto"
	api "github.com/begonia-org/go-sdk/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/spark-lence/tiga/loadbalance"
	"github.com/spark-lence/tiga/pool"
)

func initTestCase() *EndpointUsecase {
	config2 := cfg.ReadConfig("dev")
	configConfig := config.NewConfig(config2)
	// mySQLDao := data.NewMySQL(config2)
	// redisDao := data.NewRDB(config2)
	// etcdDao := data.NewEtcd(config2)
	// dataData := data.NewData(mySQLDao, redisDao, etcdDao)
	// fileRepo := data.NewFileRepoImpl(dataData)
	fileUsecase := file.NewFileUsecase(nil, configConfig)
	endpoint := NewEndpointUsecase(nil, fileUsecase, configConfig)
	gateway.New(&dp.GatewayConfig{GatewayAddr: "127.0.0.1:12138", GrpcProxyAddr: "127.0.0.1:12139"}, &dp.GrpcServerOptions{
		Middlewares: make([]dp.GrpcProxyMiddleware, 0),
		Options:     make([]grpc.ServerOption, 0),
		PoolOptions: make([]pool.PoolOptionsBuildOption, 0),
	})

	return endpoint
	// endpointRepo := data.NewEndpointRepoImpl(dataData)
	// return fileUsecase
}
func TestCreateEndpoint(t *testing.T) {
	c.Convey("Given a list of endpoints and a load balance type", t, func() {
		endpoint := initTestCase()
		ctx := context.Background()
		// basePath, _ := os.Getwd()
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("Failed to retrieve current file path")
		}

		// 计算目标文件的路径
		// 例如，如果当前文件在 internal/biz/gateway 下，
		// 而目标文件在 example 目录下，你需要相应地调整路径
		targetFilePath := filepath.Join(filepath.Dir(currentFile), "../../../example/protos.tar.gz")

		// filePath := filepath.Join(basePath, "../example/protos.tar.gz")

		data, err := os.ReadFile(targetFilePath)
		c.So(err, c.ShouldBeNil)
		hasher := sha256.New()
		hasher.Write(data)
		hash := hasher.Sum(nil)
		hashStr := hex.EncodeToString(hash)
		contentType := http.DetectContentType(data)
		rsp, err := endpoint.file.Upload(ctx, &api.UploadFileRequest{
			Key:         "endpoints/protos.tar.gz",
			Content:     data,
			Sha256:      hashStr,
			ContentType: contentType,
		}, "tester")
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		endpointRsp, err := endpoint.CreateEndpoint(ctx, &api.AddEndpointRequest{
			Name:        "test",
			ServiceName: "test",
			Description: "test endpoint",
			ProtoPath:   rsp.Uri,
			Endpoints: []*api.EndpointMeta{
				{
					Addr: "127.0.0.1:8081",
				},
				{
					Addr: "127.0.0.1:1082",
				},
			},
			Balance: string(loadbalance.RRBalanceType),
		}, "tester")
		c.So(err, c.ShouldBeNil)
		c.So(endpointRsp, c.ShouldNotBeNil)
		_, err = protoregistry.GlobalTypes.FindMessageByName("helloworld.HelloRequest")
		c.So(err, c.ShouldBeNil)
		enum, err := protoregistry.GlobalTypes.FindEnumByName("begonia.org.begonia.example.common.Code")
		c.So(err, c.ShouldBeNil)
		c.So(int32(enum.Descriptor().Values().ByName("OK").Number()), c.ShouldEqual, 2000)
	})
}
