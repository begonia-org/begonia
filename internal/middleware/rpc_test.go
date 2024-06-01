package middleware_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia/internal/middleware"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	"github.com/begonia-org/go-sdk/example"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestPluginUnaryInterceptor(t *testing.T) {
	c.Convey("test plugin unary interceptor", t, func() {
		go example.RunPlugins(":9850")
		time.Sleep(2 * time.Second)
		lb := goloadbalancer.NewGrpcLoadBalance(&goloadbalancer.Server{
			Name: "test",
			Endpoints: []goloadbalancer.EndpointServer{
				{
					Addr: "127.0.0.1:9850",
				},
			},
			Pool: &goloadbalancer.PoolConfig{
				MaxOpenConns:   10,
				MaxIdleConns:   5,
				MaxActiveConns: 5,
			},
		})
		mid := middleware.NewPluginImpl(lb, "test", 3*time.Second)
		c.So(mid.Name(), c.ShouldEqual, "test")
		mid.SetPriority(3)
		c.So(mid.Priority(), c.ShouldEqual, 3)
		addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9090")
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(peer.FromContext, &peer.Peer{Addr: addr}, true)
		defer patch.Reset()
		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "127.0.0.1:9090")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)

		_, err = mid.Info(metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test")), &emptypb.Empty{})
		c.So(err, c.ShouldBeNil)

		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldBeNil)
		patch2 := gomonkey.ApplyFuncSeq(metadata.FromIncomingContext, []gomonkey.OutputCell{{
			Values: gomonkey.Params{metadata.New(map[string]string{"test": "test"}), true},
			Times:  2,
		},
			{
				Values: gomonkey.Params{nil, false},
			},
		})
		defer patch2.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		patch2.Reset()
		c.So(err, c.ShouldBeNil)

	})
}

func TestPluginUnaryInterceptorErr(t *testing.T) {
	c.Convey("test plugin unary interceptor", t, func() {
		go example.RunPlugins(":9850")
		time.Sleep(2 * time.Second)
		lb := goloadbalancer.NewGrpcLoadBalance(&goloadbalancer.Server{
			Name: "test",
			Endpoints: []goloadbalancer.EndpointServer{
				{
					Addr: "127.0.0.1:9850",
				},
			},
			Pool: &goloadbalancer.PoolConfig{
				MaxOpenConns:   10,
				MaxIdleConns:   5,
				MaxActiveConns: 5,
			},
		})
		mid := middleware.NewPluginImpl(lb, "test", 3*time.Second)
		c.So(mid.Name(), c.ShouldEqual, "test")
		mid.SetPriority(3)
		c.So(mid.Priority(), c.ShouldEqual, 3)
		_, err := mid.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get metadata from context error")

		patch := gomonkey.ApplyMethodReturn(lb, "Select", nil, fmt.Errorf("select endpoint error"))
		defer patch.Reset()
		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "select endpoint error")
		_, err = mid.Info(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "")), &emptypb.Empty{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "select endpoint error")
		patch.Reset()
		ep := goloadbalancer.NewGrpcEndpoint("", nil)
		patch2 := gomonkey.ApplyMethodReturn(ep, "Get", nil, fmt.Errorf("get connection error"))
		defer patch2.Reset()
		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get connection error")
		_, err = mid.Info(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "")), &emptypb.Empty{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get connection error")
		patch2.Reset()
		patch3 := gomonkey.ApplyFuncReturn(anypb.New, nil, fmt.Errorf("new any to plugin error"))
		defer patch3.Reset()
		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		patch3.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "new any to plugin error")

		patch4 := gomonkey.ApplyFuncReturn((*anypb.Any).UnmarshalTo, fmt.Errorf("unmarshal to request error"))
		defer patch4.Reset()
		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "127.0.0.1:9090")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		patch4.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "unmarshal to request error")

	})

}

func TestPluginStreamInterceptorErr(t *testing.T) {
	c.Convey("test plugin unary interceptor", t, func() {
		go example.RunPlugins(":9850")
		time.Sleep(2 * time.Second)
		lb := goloadbalancer.NewGrpcLoadBalance(&goloadbalancer.Server{
			Name: "test",
			Endpoints: []goloadbalancer.EndpointServer{
				{
					Addr: "127.0.0.1:9850",
				},
			},
			Pool: &goloadbalancer.PoolConfig{
				MaxOpenConns:   10,
				MaxIdleConns:   5,
				MaxActiveConns: 5,
			},
		})
		mid := middleware.NewPluginImpl(lb, "test", 3*time.Second)
		c.So(mid.Name(), c.ShouldEqual, "test")
		mid.SetPriority(3)
		c.So(mid.Priority(), c.ShouldEqual, 3)
		patch := gomonkey.ApplyFuncReturn((*testStream).RecvMsg, fmt.Errorf("recv msg error"))
		defer patch.Reset()
		err := mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "recv msg error")
		patch1 := gomonkey.ApplyMethodReturn(mid, "Apply", nil, fmt.Errorf("call test plugin error"))
		defer patch1.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		patch1.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "call test plugin error")

		patch2 := gomonkey.ApplyFuncReturn((*anypb.Any).UnmarshalTo, fmt.Errorf("unmarshal to request error"))
		defer patch2.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		patch2.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "unmarshal to request error")

	})

}
