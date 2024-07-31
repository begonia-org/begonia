package gateway

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/config"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	"github.com/begonia-org/go-sdk/example"
	c "github.com/smartystreets/goconvey/convey"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type streamMock struct {
	ctx context.Context
}
type clientStreamMock struct {
	ctx context.Context
}

func (*streamMock) SendHeader(md metadata.MD) error {
	return nil
}
func (*streamMock) SetHeader(md metadata.MD) error {
	return nil
}
func (*streamMock) SetTrailer(md metadata.MD) {
}
func (s *streamMock) Context() context.Context {
	return s.ctx
}
func (*streamMock) SendMsg(m interface{}) error {
	time.Sleep(1 * time.Second)
	return nil
}
func (*streamMock) RecvMsg(m interface{}) error {
	return nil
}
func (*clientStreamMock) SendMsg(m interface{}) error {
	return nil
}

func (*clientStreamMock) RecvMsg(m interface{}) error {
	return nil

}
func (*clientStreamMock) CloseSend() error {
	return nil
}
func (*clientStreamMock) Header() (metadata.MD, error) {
	return metadata.MD{}, nil
}
func (*clientStreamMock) Trailer() metadata.MD {
	return metadata.MD{}
}
func (*clientStreamMock) Context() context.Context {
	return context.Background()
}
func (*clientStreamMock) SendHeader(md metadata.MD) error {
	return nil
}

func TestGrpcHandleErr(t *testing.T) {
	c.Convey("test grpc handle err", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		rander := rand.New(rand.NewSource(time.Now().Unix())) // 初始化随机数种子
		min := 1949
		max := 12138
		randomNumber := rander.Intn(max-min+1) + min
		helloAddr := fmt.Sprintf("127.0.0.1:%d", randomNumber+2)
		go example.Run(helloAddr)
		time.Sleep(2 * time.Second)
		endps, err := NewLoadBalanceEndpoint(loadbalance.RRBalanceType, []*api.EndpointMeta{{
			Addr:   helloAddr,
			Weight: 0,
		}})
		c.So(err, c.ShouldBeNil)

		load, _ := loadbalance.New(loadbalance.RRBalanceType, endps)
		lb := NewGrpcLoadBalancer()
		lb.Register(load, pd)
		fullMethod1 := ""
		mid := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			fullMethod1 = method
			return streamer(ctx, desc, cc, method, opts...)
		}
		fullMethod2 := ""
		mid2 := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			fullMethod2 = method
			return streamer(ctx, desc, cc, method, opts...)
		}
		proxy := NewGrpcProxy(lb, Log, mid, mid2)
		proxy.buildServiceDesc(pd)
		proxy.Register(load, pd)

		stream := &streamMock{
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("uri", "/api/v1/example/server/websocket")),
		}
		patch := gomonkey.ApplyFuncReturn(grpc.MethodFromServerStream, strings.ToUpper("/helloworld.Greeter/SayHelloWebsocket"), true)
		patch.ApplyFuncReturn(grpc.NewClientStream, &clientStreamMock{
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("uri", "/api/v1/example/server/websocket")),
		}, nil)
		addrs, _ := net.InterfaceAddrs()
		var localAddr net.Addr
		for _, addr := range addrs {
			// 检查地址类型和如果是IP地址我们就打印它
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					// fmt.Println(ipnet.IP.String())
					localAddr = addr
					break
				}
			}
		}
		patch.ApplyFuncReturn(peer.FromContext, &peer.Peer{Addr: localAddr}, true)
		defer patch.Reset()

		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			// {
			// 	patch:  metadata.FromIncomingContext,
			// 	err:    fmt.Errorf("metadata not exists in context"),
			// 	output: []interface{}{nil, false},
			// },
			{
				patch:  mid,
				err:    fmt.Errorf("mid handle err"),
				output: []interface{}{nil, fmt.Errorf("mid handle err")},
			},
			{
				patch:  (*clientStreamMock).CloseSend,
				err:    fmt.Errorf("close send err"),
				output: []interface{}{fmt.Errorf("close send err")},
			},
			{
				patch:  (*clientStreamMock).Header,
				err:    fmt.Errorf("header err"),
				output: []interface{}{nil, fmt.Errorf("header err")},
			},
			{
				patch:  (*streamMock).SendHeader,
				err:    fmt.Errorf("send header err"),
				output: []interface{}{fmt.Errorf("send header err")},
			},
			{
				patch:  (*streamMock).SendMsg,
				err:    fmt.Errorf("send msg err"),
				output: []interface{}{fmt.Errorf("send msg err")},
			},
		}

		patch3 := gomonkey.ApplyFunc((*clientStreamMock).SendMsg, func(_ *clientStreamMock, m interface{}) error {
			time.Sleep(3 * time.Second)
			return io.EOF
		})
		defer patch3.Reset()
		patch6 := gomonkey.ApplyFuncReturn(grpc.Method, "/helloworld.Greeter/SayHelloWebsocket", true)
		defer patch6.Reset()
		for _, caseV := range cases {
			patch2 := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch2.Reset()

			err = proxy.Do(&hello.HelloRequest{}, stream)
			t.Log(caseV.err.Error())
			// t.Logf("err:%v", err.Error())
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
			patch2.Reset()
		}
		c.So(fullMethod1, c.ShouldEqual, strings.ToUpper("/helloworld.Greeter/SayHelloWebsocket"))
		c.So(fullMethod2, c.ShouldEqual, strings.ToUpper("/helloworld.Greeter/SayHelloWebsocket"))
		patch3.Reset()

		errChan2 := make(chan error, 3)

		errChan2 <- io.EOF
		errChan2 <- io.EOF
		patch4 := gomonkey.ApplyFuncReturn((*GrpcProxy).forwardServerToClient, errChan2)
		err = proxy.Do(&hello.HelloRequest{}, stream)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "proxying should never reach")
		defer patch4.Reset()
		patch.Reset()

		patch4.Reset()
		// patch6.Reset()
		patch7 := gomonkey.ApplyFuncReturn(grpc.MethodFromServerStream, "/helloworld.Greeter/SayHelloWebsocket", false)
		defer patch7.Reset()

		proxy2 := NewGrpcProxy(lb, Log, mid)
		err = proxy2.Do(&hello.HelloRequest{}, stream)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "stream not exists in context")

		proxy2.chainStreamClientInterceptors()
		f := func() {
			proxy.forwardClientToServer(nil, stream)
		}
		c.So(f, c.ShouldNotPanic)

	})
}

func TestProxyDo(t *testing.T) {
	pd, _ := readDesc(config.NewConfig(cfg.ReadConfig("test")))
	p := &GrpcProxy{ioType: make(map[string]*IOType), lb: NewGrpcLoadBalancer(), log: Log}
	p.buildServiceDesc(pd)
}
