package auth_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/middleware/auth"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAccessKeyAuthMiddleware(t *testing.T) {
	c.Convey("TestAccessKeyAuthMiddleware", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		repo := data.NewAppRepo(config, gateway.Log)
		cnf := cfg.NewConfig(config)

		akBiz := biz.NewAccessKeyAuth(repo, cnf, gateway.Log)
		ak := auth.NewAccessKeyAuth(akBiz, cnf, gateway.Log)
		ak.SetPriority(1)
		c.So(ak.Name(), c.ShouldEqual, "ak_auth")
		c.So(ak.Priority(), c.ShouldEqual, 1)
		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

		pd, _ := gateway.NewDescription(pbFile)
		R.LoadAllRouters(pd)
		_, err := ak.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/example.v1.HelloService/TEST",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("metadata not exists in context")

		})
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("metadata not exists in context").Error())

		err = ak.StreamInterceptor(context.Background(), &testStream{ctx: context.Background()}, &grpc.StreamServerInfo{
			FullMethod: "/example.v1.HelloService/TEST",
		}, func(srv any, stream grpc.ServerStream) error {
			return fmt.Errorf("metadata not exists in context")
		})
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("metadata not exists in context").Error())

		patch := gomonkey.ApplyFuncReturn((*auth.AccessKeyAuthMiddleware).StreamRequestBefore, nil, nil)
		patch = patch.ApplyFuncReturn((*auth.AccessKeyAuthMiddleware).StreamResponseAfter, fmt.Errorf("StreamResponseAfter err"))
		defer patch.Reset()
		err = ak.StreamInterceptor(context.Background(), &testStream{ctx: context.Background()}, &grpc.StreamServerInfo{FullMethod: "/integration.TestService/Get"}, func(srv any, stream grpc.ServerStream) error {
			return nil

		})
		c.So(err, c.ShouldBeNil)
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn((*auth.AccessKeyAuthMiddleware).StreamRequestBefore, nil, fmt.Errorf("StreamRequestBefore err"))
		defer patch2.Reset()
		err = ak.StreamInterceptor(context.Background(), &testStream{ctx: context.Background()}, &grpc.StreamServerInfo{FullMethod: "/integration.TestService/Get"}, func(srv any, stream grpc.ServerStream) error {
			return nil

		})
		patch2.Reset()
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("StreamRequestBefore err").Error())
	})
}
func TestRequestBeforeErr(t *testing.T) {
	c.Convey("TestRequestBeforeErr", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		repo := data.NewAppRepo(config, gateway.Log)
		cnf := cfg.NewConfig(config)

		akBiz := biz.NewAccessKeyAuth(repo, cnf, gateway.Log)
		ak := auth.NewAccessKeyAuth(akBiz, cnf, gateway.Log)
		patch := gomonkey.ApplyFuncReturn(gosdk.NewGatewayRequestFromGrpc, nil, fmt.Errorf("NewGatewayRequestFromGrpc err"))
		defer patch.Reset()
		_, err := ak.RequestBefore(context.TODO(), &grpc.UnaryServerInfo{}, nil)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "NewGatewayRequestFromGrpc err")
		patch1 := gomonkey.ApplyFuncReturn((*biz.AccessKeyAuth).AppValidator, "dxadada", nil)
		patch1 = patch1.ApplyFuncReturn((*biz.AccessKeyAuth).GetAppOwner, "", fmt.Errorf("owner is empty"))
		defer patch1.Reset()
		_, err1 := ak.RequestBefore(context.TODO(), &grpc.UnaryServerInfo{}, nil)
		patch1.Reset()
		c.So(err1, c.ShouldNotBeNil)
		c.So(err1.Error(), c.ShouldContainSubstring, "owner is empty")

		patch2 := gomonkey.ApplyFuncReturn(gosdk.NewGatewayRequestFromGrpc, nil, nil)
		patch2 = patch2.ApplyFuncReturn((*biz.AccessKeyAuth).AppValidator, "dxadada", nil)
		patch2 = patch2.ApplyFuncReturn((*biz.AccessKeyAuth).GetAppOwner, "dadad", nil)
		defer patch2.Reset()
		ctx, err2 := ak.RequestBefore(context.TODO(), &grpc.UnaryServerInfo{}, nil)
		patch2.Reset()
		c.So(err2, c.ShouldBeNil)

		md, ok := metadata.FromIncomingContext(ctx)
		c.So(ok, c.ShouldBeTrue)
		c.So(md.Get("x-identity"), c.ShouldResemble, []string{"dadad"})

	})
}

func TestValidateStream(t *testing.T) {
	c.Convey("TestValidateStream", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		repo := data.NewAppRepo(config, gateway.Log)
		cnf := cfg.NewConfig(config)

		akBiz := biz.NewAccessKeyAuth(repo, cnf, gateway.Log)
		ak := auth.NewAccessKeyAuth(akBiz, cnf, gateway.Log)
		patch := gomonkey.ApplyFuncReturn(gosdk.NewGatewayRequestFromGrpc, nil, fmt.Errorf("NewGatewayRequestFromGrpc err"))
		defer patch.Reset()
		_, err := ak.ValidateStream(context.TODO(), nil, "", nil)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "NewGatewayRequestFromGrpc err")

		stream := auth.NewGrpcStream(&testStream{}, "", context.TODO(), nil)
		err = ak.StreamResponseAfter(context.TODO(), stream, nil)
		c.So(err, c.ShouldBeNil)
	})
}
