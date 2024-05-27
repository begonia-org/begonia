package auth_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/middleware/auth"
	"github.com/begonia-org/begonia/internal/pkg"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAPIKeyUnaryInterceptor(t *testing.T) {
	c.Convey("TestAPIKeyUnaryInterceptor", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		apikey := auth.NewApiKeyAuth(cnf)
		apikey.SetPriority(1)
		c.So(apikey.Name(), c.ShouldEqual, "api_key_auth")
		c.So(apikey.Priority(), c.ShouldEqual, 1)

		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

		pd, _ := gateway.NewDescription(pbFile)
		R.LoadAllRouters(pd)

		_, err := apikey.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/example.v1/TEST",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("dont need to call validator")

		})
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("dont need to call validator").Error())

		_, err = apikey.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("metadata not exists in context").Error())

		_, err = apikey.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("app", "test")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("apikey not exists in context").Error())

		_, err = apikey.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", "test")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrAPIKeyNotMatch.Error())

		_, err = apikey.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", cnf.GetAdminAPIKey())), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err, c.ShouldBeNil)
	})
}

func TestApiKeyStreamInterceptor(t *testing.T) {
	c.Convey("TestApiKeyStreamInterceptor", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		apikey := auth.NewApiKeyAuth(cnf)
		apikey.SetPriority(1)
		c.So(apikey.Name(), c.ShouldEqual, "api_key_auth")
		c.So(apikey.Priority(), c.ShouldEqual, 1)

		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

		pd, _ := gateway.NewDescription(pbFile)
		R.LoadAllRouters(pd)
		err := apikey.StreamInterceptor(&hello.HelloRequest{}, &testStream{}, &grpc.StreamServerInfo{
			FullMethod: "TEST/TETS",
		}, func(srv any, stream grpc.ServerStream) error {
			return nil
		})
		c.So(err, c.ShouldBeNil)

		err = apikey.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", cnf.GetAdminAPIKey())),
		}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldBeNil)
	})
}
