package middleware_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/middleware"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type testStream struct {
	ctx context.Context
}

func (t *testStream) SetHeader(metadata.MD) error {
	return nil
}
func (t *testStream) SendHeader(metadata.MD) error {
	return nil
}
func (t *testStream) SetTrailer(metadata.MD) {

}
func (t *testStream) Context() context.Context {
	return t.ctx

}
func (t *testStream) SendMsg(m interface{}) error {
	return nil
}
func (t *testStream) RecvMsg(m interface{}) error {
	return nil
}

type greeterSayHelloWebsocketServer struct {
	grpc.ServerStream
}

func (x *greeterSayHelloWebsocketServer) Send(m *hello.HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greeterSayHelloWebsocketServer) Recv() (*hello.HelloRequest, error) {
	m := new(hello.HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
func (x *greeterSayHelloWebsocketServer) Context() context.Context {
	return x.ServerStream.Context()

}
func TestStreamInterceptor(t *testing.T) {
	c.Convey("test stream interceptor", t, func() {
		mid := middleware.NewHttp()
		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata")

		pd, err := gateway.NewDescription(pbFile)
		c.So(err, c.ShouldBeNil)
		R.LoadAllRouters(pd)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("grpcgateway-accept", "application/json"))
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv any, ss grpc.ServerStream) error {
			return ss.SendMsg(srv)
		})
		c.So(err, c.ShouldBeNil)
	})
}
func TestUnaryInterceptor(t *testing.T) {
	c.Convey("test unary interceptor", t, func() {
		mid := middleware.NewHttp()
		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata")

		pd, err := gateway.NewDescription(pbFile)
		c.So(err, c.ShouldBeNil)
		R.LoadAllRouters(pd)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("grpcgateway-accept", "application/json"))
		req, err := mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return req, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(req, c.ShouldNotBeNil)
		mid.SetPriority(1)
		c.So(mid.Priority(), c.ShouldEqual, 1)
		c.So(mid.Name(), c.ShouldEqual, "http")
		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, gosdk.NewError(pkg.ErrAPIKeyNotMatch, int32(user.UserSvrCode_USER_ACCOUNT_ERR), codes.Internal, "test", gosdk.WithClientMessage("test message"))
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(req, c.ShouldNotBeNil)

		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, status.Error(codes.NotFound, "test")
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(req, c.ShouldNotBeNil)

		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, status.Error(codes.Unimplemented, "test")
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(req, c.ShouldNotBeNil)

		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("test")
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(req, c.ShouldNotBeNil)

		patch := gomonkey.ApplyFuncReturn(protojson.Marshal, nil, fmt.Errorf("test"))
		defer patch.Reset()
		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &hello.HelloReply{}, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test")
		c.So(req, c.ShouldBeNil)
		patch.Reset()

		patch1 := gomonkey.ApplyFuncReturn(protojson.Unmarshal, fmt.Errorf("unmarshal test"))
		defer patch1.Reset()
		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &hello.HelloReply{}, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "unmarshal test")
		c.So(req, c.ShouldBeNil)
		patch1.Reset()

		req, err = mid.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &hello.HelloReply{}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(req, c.ShouldNotBeNil)

		req, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("grpcgateway-accept", "test")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &hello.HelloReply{}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(req, c.ShouldNotBeNil)

		req, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &httpbody.HttpBody{}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(req, c.ShouldNotBeNil)

		req, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("grpcgateway-test", "test")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &hello.HelloReply{}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(req, c.ShouldNotBeNil)

		stream := &middleware.HttpStream{ServerStream: &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: context.Background()}}, FullMethod: "/INTEGRATION.TESTSERVICE/GET"}
		err = stream.SendMsg(&hello.HelloReply{})
		c.So(err, c.ShouldBeNil)

		stream = &middleware.HttpStream{ServerStream: &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("grpcgateway-test", "test"))}}, FullMethod: "/INTEGRATION.TESTSERVICE/GET"}
		err = stream.SendMsg(&hello.HelloReply{})
		c.So(err, c.ShouldBeNil)

		stream = &middleware.HttpStream{ServerStream: &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("grpcgateway-accept", "test"))}}, FullMethod: "/INTEGRATION.TESTSERVICE/GET"}
		err = stream.SendMsg(&hello.HelloReply{})
		c.So(err, c.ShouldBeNil)

	})
}
