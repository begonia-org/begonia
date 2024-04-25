package transport

import (
	"net"
	"strings"
	"testing"
	"time"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	"github.com/begonia-org/go-sdk/example"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

func testServers() {
	go example.Run(":12138")
	go example.Run(":12139")
	go example.Run(":12140")
}
func TestGrpcLoadBalancer(t *testing.T) {
	c.Convey("TestGrpcLoadBalancer", t, func() {
		testServers()
		time.Sleep(time.Second * 3)
		servers := []*EndpointServer{
			{
				Addr: "127.0.0.1:12138",
			},
			{
				Addr: "127.0.0.1:12139",
			},
			{
				Addr: "127.0.0.1:12140",
			},
		}
		endpoints := make([]loadbalance.Endpoint, 0)
		for _, server := range servers {
			pool := NewGrpcConnPool(server.Addr)
			endpoint := NewGrpcEndpoint(server.Addr, pool)
			endpoints = append(endpoints, endpoint)
		}
		rr := loadbalance.NewRoundRobinBalance(endpoints)
		pd, err := NewDescription("./protos")
		c.So(err, c.ShouldBeNil)
		grpcLb := NewGrpcLoadBalancer()
		grpcLb.Register(rr, pd)
		c.So(grpcLb.lb, c.ShouldContainKey, strings.ToUpper("/helloworld.Greeter/SayHello"))
		c.So(grpcLb.lb, c.ShouldContainKey, strings.ToUpper("/helloworld.Greeter/SayHelloGet"))
		c.So(grpcLb.lb, c.ShouldContainKey, strings.ToUpper("/helloworld.Greeter/SayHelloStreamReply"))

		proxy := NewGrpcProxy(grpcLb)
		server := grpc.NewServer(
			grpc.UnknownServiceHandler(proxy.Handler))
		lis, err := net.Listen("tcp", ":39527")
		c.So(err, c.ShouldBeNil)
		go server.Serve(lis)
		time.Sleep(time.Second * 1)
		cli := example.NewClient("127.0.0.1:39527")
		msg, err := cli.SayHello()
		c.So(err, c.ShouldBeNil)
		t.Log(msg)
		c.So(msg, c.ShouldEqual, "Hello begonia")
		ch, err := cli.SayHelloStreamReply()
		c.So(err, c.ShouldBeNil)
		c.So(ch, c.ShouldNotBeNil)
		isRecv := false
		for msg := range ch {
			isRecv = true
			c.So(msg, c.ShouldContainSubstring, "begonia")
		}
		c.So(isRecv, c.ShouldBeTrue)
		msg, err = cli.SayHelloStreamSend()
		c.So(err, c.ShouldBeNil)
		c.So(msg, c.ShouldContainSubstring, "你好:")
		ch, err = cli.SayHelloBidiStream()
		c.So(err, c.ShouldBeNil)
		c.So(ch, c.ShouldNotBeNil)
		words := []string{
			"你好",
			"hello",
			"こんにちは",
			"안녕하세요",
		}
		index := 0
		for msg := range ch {
			t.Log(msg)
			c.So(msg, c.ShouldEqual, words[index])
			index++
		}
		// time.Sleep(time.Second * 3)
		for _, lb := range grpcLb.lb {
			for _, endpoint := range lb.GetEndpoints() {
				c.So(endpoint.Stats().GetActivateConns(), c.ShouldEqual, 0)
				c.So(endpoint.Stats().GetIdleConns(), c.ShouldEqual, 4)
			}
		}
	})
}
