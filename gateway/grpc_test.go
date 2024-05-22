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
}
type clientStreamMock struct{}

func (*streamMock) SendHeader(md metadata.MD) error {
	return nil
}
func (*streamMock) SetHeader(md metadata.MD) error {
	return nil
}
func (*streamMock) SetTrailer(md metadata.MD) {
}
func (*streamMock) Context() context.Context {
	return context.Background()
}
func (*streamMock) SendMsg(m interface{}) error {
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
		mid := func(srv interface{}, serverStream grpc.ServerStream) error {
			return nil
		}
		proxy := NewGrpcProxy(lb, mid)
		stream := &streamMock{}
		patch := gomonkey.ApplyFuncReturn(grpc.MethodFromServerStream, strings.ToUpper("/helloworld.Greeter/SayHelloWebsocket"), true)
		patch.ApplyFuncReturn(grpc.NewClientStream, &clientStreamMock{}, nil)
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
			{
				patch:  mid,
				err:    fmt.Errorf("mid handle err"),
				output: []interface{}{fmt.Errorf("mid handle err")},
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
		for _, caseV := range cases {
			patch2 := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch2.Reset()
			err = proxy.Handler(&hello.HelloRequest{}, stream)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
			patch2.Reset()
		}
		patch3.Reset()

		errChan2 := make(chan error, 3)

		errChan2 <- io.EOF
		errChan2 <- io.EOF
		patch4 := gomonkey.ApplyFuncReturn((*GrpcProxy).forwardServerToClient, errChan2)
		err = proxy.Handler(&hello.HelloRequest{}, stream)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "proxying should never reach")
		defer patch4.Reset()
		patch.Reset()
		patch4.Reset()

	})
}
