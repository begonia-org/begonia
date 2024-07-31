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
	api "github.com/begonia-org/go-sdk/api/plugin/v1"
	"github.com/begonia-org/go-sdk/example"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type testClientStream struct {
	ctx context.Context
	grpc.ClientStream
}

func (t *testClientStream) Context() context.Context {
	return t.ctx
}
func (t *testClientStream) SendMsg(m interface{}) error {
	return nil
}
func (t *testClientStream) RecvMsg(m interface{}) error {
	return nil

}
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
		// patch2 := gomonkey.ApplyFuncSeq(metadata.FromIncomingContext, []gomonkey.OutputCell{{
		// 	Values: gomonkey.Params{metadata.New(map[string]string{"test": "test"}), true},
		// 	Times:  4,
		// },
		// 	{
		// 		Values: gomonkey.Params{nil, false},
		// 	},
		// })
		// defer patch2.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			err := ss.RecvMsg(srv)
			if err != nil {
				t.Logf("recv msg error: %v", err)
			}
			return err
		})
		// patch2.Reset()
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
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get metadata from context error")
		// select err
		patch5 := gomonkey.ApplyMethodReturn(lb, "Select", nil, fmt.Errorf("select endpoint error"))
		defer patch5.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "select endpoint error")
		patch5.Reset()
		enp, _ := lb.Select("127.0.0.1")
		patch6 := gomonkey.ApplyMethodReturn(enp, "Get", nil, fmt.Errorf("get connection error"))
		defer patch6.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		patch6.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get connection error")
		// call plugin metadata err
		cli := api.NewPluginServiceClient(nil)
		patch7 := gomonkey.ApplyMethodReturn(cli, "Metadata", nil, fmt.Errorf("call test plugin error"))
		defer patch7.Reset()
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test"))}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		patch7.Reset()
		c.So(err.Error(), c.ShouldContainSubstring, "call test plugin error")
		// call apply err

		patch8 := gomonkey.ApplyMethodReturn(cli, "Apply", nil, fmt.Errorf("call plugin error"))
		defer patch8.Reset()
		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Forwarded-For", "127.0.0.1:9090")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		patch8.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "call plugin error")
	})

}

func TestPluginStreamInterceptorErr(t *testing.T) {
	c.Convey("test plugin unary interceptor", t, func() {
		go example.RunPlugins(":9851")
		time.Sleep(2 * time.Second)
		lb := goloadbalancer.NewGrpcLoadBalance(&goloadbalancer.Server{
			Name: "test",
			Endpoints: []goloadbalancer.EndpointServer{
				{
					Addr: "127.0.0.1:9851",
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
		patch1 := gomonkey.ApplyMethodReturn(mid, "Apply", nil, nil, fmt.Errorf("call test plugin error"))
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

func TestRPCStreamClientInterceptor(t *testing.T) {
	c.Convey("test rpc stream client interceptor", t, func() {
		go example.RunPlugins(":9852")
		time.Sleep(2 * time.Second)
		lb := goloadbalancer.NewGrpcLoadBalance(&goloadbalancer.Server{
			Name: "test",
			Endpoints: []goloadbalancer.EndpointServer{
				{
					Addr: "127.0.0.1:9852",
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
		st, err := mid.StreamClientInterceptor(metadata.NewOutgoingContext(context.Background(), metadata.Pairs("key", "value")), nil, nil, "/INTEGRATION.TESTSERVICE/GET", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &testClientStream{ctx: ctx}, nil
		},
		)
		c.So(err, c.ShouldBeNil)
		c.So(st, c.ShouldNotBeNil)

	},
	)
}
