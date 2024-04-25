package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	loadbalance "github.com/begonia-org/go-loadbalancer"
	v1 "github.com/begonia-org/go-sdk/api/v1"
	"github.com/begonia-org/go-sdk/example"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

func TestGateway(t *testing.T) {
	c.Convey("test gateway", t, func() {
		if err := os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info"); err != nil {
			log.Fatalf("Failed to set GRPC_GO_LOG_SEVERITY_LEVEL: %v", err)
		}
		testServers()
		time.Sleep(time.Second * 3)
		cfg := &GatewayConfig{
			GrpcProxyAddr: ":19527",
			GatewayAddr:   ":9527",
		}
		opts := &GrpcServerOptions{
			Middlewares: make([]GrpcProxyMiddleware, 0),
			Options:     make([]grpc.ServerOption, 0),
			PoolOptions: make([]loadbalance.PoolOptionsBuildOption, 0),
		}
		gateway := NewGateway(cfg, opts)

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
		err = gateway.RegisterService(context.Background(), pd, rr)
		c.So(err, c.ShouldBeNil)
		go gateway.Start()
		time.Sleep(time.Second * 3)

		cli := example.NewClient("127.0.0.1:9527")
		msg, err := cli.SayHello()
		c.So(err, c.ShouldBeNil)
		t.Log(msg)
		c.So(msg, c.ShouldEqual, "Hello begonia")
		ch, err := cli.SayHelloStreamReply()
		c.So(err, c.ShouldBeNil)
		c.So(ch, c.ShouldNotBeNil)

		req := &v1.HelloRequest{
			Msg: "begonia",
		}
		jsonData, err := json.Marshal(req)
		c.So(err, c.ShouldBeNil)
		// 创建请求体
		reqBody := bytes.NewBuffer(jsonData)
		// 配置HTTP客户端使用代理
		proxyURL, _ := url.Parse("http://127.0.0.1:8888")

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client := &http.Client{
			Transport: transport,
		}
		// 发送 POST 请求
		request, _ := http.NewRequest("POST", "http://127.0.0.1:9527/api/v1/example/hello/helloworld", reqBody)
		request.Header.Set("content-type", "application/json")
		resp, err := client.Do(request)
		c.So(err, c.ShouldBeNil)
		c.So(resp, c.ShouldNotBeNil)
		var respData v1.HelloReply
		err = json.NewDecoder(resp.Body).Decode(&respData)
		c.So(err, c.ShouldBeNil)
		c.So(respData.Message, c.ShouldEqual, "Hello begonia")
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		defer resp.Body.Close()
	})
}
