package gateway

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	loadbalance "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/endpoint/v1"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/example"
	"github.com/gorilla/websocket"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/r3labs/sse/v2"
	c "github.com/smartystreets/goconvey/convey" // 别名导入
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/cenkalti/backoff.v1"
)

// var endpoint HttpEndpoint
var gwPort = 1949
var gw *GatewayServer
var onceInit sync.Once
var randomNumber int
var eps []loadbalance.Endpoint

func newTestServer(gwPort, randomNumber int) (*GrpcServerOptions, *GatewayConfig) {
	opts := &GrpcServerOptions{
		Middlewares:     make([]GrpcProxyMiddleware, 0),
		Options:         make([]grpc.ServerOption, 0),
		PoolOptions:     make([]loadbalance.PoolOptionsBuildOption, 0),
		HttpMiddlewares: make([]gwRuntime.ServeMuxOption, 0),
		HttpHandlers:    make([]func(http.Handler) http.Handler, 0),
	}
	gwCnf := &GatewayConfig{
		GatewayAddr:   fmt.Sprintf("127.0.0.1:%d", gwPort),
		GrpcProxyAddr: fmt.Sprintf("127.0.0.1:%d", randomNumber+1),
	}

	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("application/json", NewJSONMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("multipart/form-data", NewFormDataMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("application/x-www-form-urlencoded", NewFormUrlEncodedMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption(gwRuntime.MIMEWildcard, NewRawBinaryUnmarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("application/octet-stream", NewRawBinaryUnmarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("text/event-stream", NewEventSourceMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption(ClientStreamContentType, NewProtobufWithLengthPrefix()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMetadata(IncomingHeadersToMetadata))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithErrorHandler(HandleErrorWithLogger(Log)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithForwardResponseOption(HttpResponseBodyModify))

	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithPoolSize(128))
	loggerMid := NewLoggerMiddleware(Log)
	except := NewException(Log)
	mid := NewMiddlewaresTest(Log)
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(loggerMid.StreamInterceptor, except.StreamInterceptor, mid.StreamInterceptor))
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(loggerMid.UnaryInterceptor, except.UnaryInterceptor, mid.UnaryInterceptor))
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	// log.Printf("env: %s", env)
	cnf := config.ReadConfig(env)
	conf := cfg.NewConfig(cnf)
	cors := &CorsHandler{
		Cors: conf.GetCorsConfig(),
	}
	opts.HttpHandlers = append(opts.HttpHandlers, cors.Handle)
	return opts, gwCnf
}
func init() {
	onceInit.Do(func() {

		rander := rand.New(rand.NewSource(time.Now().Unix())) // 初始化随机数种子
		min := 1949
		max := 12138
		randomNumber = rander.Intn(max-min+1) + min
		// randomNumber := 31949
		gwPort = randomNumber
		// gw = newTestServer(gwPort, randomNumber)
		opts, cnf := newTestServer(gwPort, randomNumber)
		gw = New(cnf, opts)

	})

}
func testRegisterClient(t *testing.T) {
	c.Convey("test register client", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(pd.GetMessageTypeByName("helloworld.HelloRequest", "hello"), c.ShouldBeNil)
		c.So(pd.GetMessageTypeByFullName("test.helloworld.HelloRequest"), c.ShouldBeNil)
		// t.Logf("pd:%+v", pd.GetGatewayJsonSchema())
		c.So(err, c.ShouldBeNil)
		c.So(pd.GetMessageTypeByFullName("helloworld.HelloRequest"), c.ShouldNotBeNil)
		c.So(pd.GetDescription(), c.ShouldNotBeEmpty)
		helloAddr := fmt.Sprintf("127.0.0.1:%d", randomNumber+2)
		endps, err := NewLoadBalanceEndpoint(loadbalance.RRBalanceType, []*api.EndpointMeta{{
			Addr:   helloAddr,
			Weight: 0,
		}})
		c.So(err, c.ShouldBeNil)

		load, err := loadbalance.New(loadbalance.RRBalanceType, endps)
		c.So(err, c.ShouldBeNil)
		err = gw.RegisterService(context.Background(), pd, load)
		c.So(err, c.ShouldBeNil)
		c.So(gw.GetLoadbalanceName(), c.ShouldEqual, loadbalance.RRBalanceType)
		go example.Run(helloAddr)
		time.Sleep(2 * time.Second)
		go gw.Start()
		time.Sleep(2 * time.Second)
		f := func() {
			gw.Start()
		}
		c.So(f, c.ShouldPanic)
		time.Sleep(4 * time.Second)
		_, err = gw.proxyLB.Select("test/.test")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, loadbalance.ErrNoEndpoint.Error())
	})
}
func testRequestGet(t *testing.T) {
	c.Convey("test request GET", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		r.Header.Set("x-uid", "12345678")
		r.Header.Set(XAccessKey, "12345678")
		// r.Header.Set("Origin", "http://www.example.com")
		c.So(err, c.ShouldBeNil)

		resp, err := http.DefaultClient.Do(r)

		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		c.So(resp.Header.Get("test"), c.ShouldEqual, "test")
		c.So(resp.Header.Get("content-type"), c.ShouldEqual, "application/json")
		c.So(resp.Header.Get("trace_id"), c.ShouldEqual, "123456")
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)
		rsp := &hello.HelloReply{}

		err = json.Unmarshal(body, rsp)
		c.So(err, c.ShouldBeNil)

		c.So(rsp.Message, c.ShouldEqual, "hello")
		c.So(rsp.Name, c.ShouldEqual, "world")

		url = fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/http-code?msg=hello", gwPort)
		r, err = http.NewRequest(http.MethodGet, url, nil)
		// r.Header.Set("Origin", "http://www.example.com")
		c.So(err, c.ShouldBeNil)

		resp, err = http.DefaultClient.Do(r)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusIMUsed)

		url = fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		r2, err := http.NewRequest(http.MethodPost, url, nil)

		c.So(err, c.ShouldBeNil)

		resp2, err := http.DefaultClient.Do(r2)

		c.So(err, c.ShouldBeNil)
		c.So(resp2.StatusCode, c.ShouldEqual, http.StatusNotImplemented)

	})
}
func testCors(t *testing.T) {
	c.Convey("test cors", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		r, err := http.NewRequest(http.MethodOptions, url, nil)
		r.Header.Set("Origin", "http://www.example.com")
		c.So(err, c.ShouldBeNil)

		resp, err := http.DefaultClient.Do(r)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusNoContent)
		c.So(resp.Header.Get("Access-Control-Allow-Origin"), c.ShouldEqual, "http://www.example.com")

		r, err = http.NewRequest(http.MethodOptions, url, nil)
		r.Header.Set("Origin", "http://www.begonia-org.com")
		c.So(err, c.ShouldBeNil)
		resp, err = http.DefaultClient.Do(r)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusForbidden)
	})
}
func testRequestPost(t *testing.T) {
	c.Convey("test request POST json", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/post", gwPort)

		r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(`{"name":"world","msg":"hello"}`))
		c.So(err, c.ShouldBeNil)
		r.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(r)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)
		rsp := &hello.HelloReply{}
		err = json.Unmarshal(body, rsp)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.Message, c.ShouldEqual, "hello")
		c.So(rsp.Name, c.ShouldEqual, "world")

	})

	c.Convey("test request POST form-data", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/post", gwPort)

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("name", "world")
		_ = writer.WriteField("msg", "hello")
		err := writer.Close()
		c.So(err, c.ShouldBeNil)

		req, err := http.NewRequest(http.MethodPost, url, payload)

		c.So(err, c.ShouldBeNil)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)
		rsp := &hello.HelloReply{}
		err = json.Unmarshal(body, rsp)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.Message, c.ShouldEqual, "hello")
		c.So(rsp.Name, c.ShouldEqual, "world")
	})

	c.Convey("test request POST form-urlencoded", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/post", gwPort)
		payload := strings.NewReader("name=world&msg=hello")
		req, err := http.NewRequest(http.MethodPost, url, payload)
		c.So(err, c.ShouldBeNil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)
		rsp := &hello.HelloReply{}
		err = json.Unmarshal(body, rsp)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.Message, c.ShouldEqual, "hello")
		c.So(rsp.Name, c.ShouldEqual, "world")

	})
	c.Convey("test request POST binary", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/body", gwPort)
		payload, _ := json.Marshal(&hello.HelloRequest{Name: "world", Msg: "hello"})
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		c.So(err, c.ShouldBeNil)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("accept", "application/octet-stream")
		resp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		c.So(resp.Header.Get("content-type"), c.ShouldEqual, "application/octet-stream")
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)
		rsp := &hello.HelloReply{}
		err = json.Unmarshal(body, rsp)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.Message, c.ShouldEqual, "hello")
		c.So(rsp.Name, c.ShouldEqual, "world")

	})
}
func testServerSideEvent(t *testing.T) {
	c.Convey("test server side event", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/server/sse/world?msg=hello", gwPort)
		// t.Logf("url:%s", url)
		// time.Sleep(30 * time.Second)
		client := sse.NewClient(url, func(c *sse.Client) {
			c.ReconnectStrategy = &backoff.StopBackOff{}
		})

		err := client.Subscribe("message", func(msg *sse.Event) {

			reply := &hello.HelloReply{}
			err := protojson.Unmarshal(msg.Data, reply)
			c.So(err, c.ShouldBeNil)
			if reply.Name == "world" && reply.Message == fmt.Sprintf("hello-%s", msg.ID) {
				c.So(true, c.ShouldBeTrue)
			} else {
				c.So(false, c.ShouldBeTrue)
			}

		})
		c.So(err, c.ShouldBeNil)

	})
}
func testWebsocket(t *testing.T) {
	c.Convey("test websocket", t, func() {
		url := fmt.Sprintf("ws://127.0.0.1:%d/api/v1/example/server/websocket", gwPort)
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer conn.Close()
		exceptRecv := make(map[string]bool)
		for i := 0; i < 10; i++ {
			in := &hello.HelloRequest{
				Name: "world",
				Msg:  fmt.Sprintf("hello-%d", i),
			}
			exceptRecv[fmt.Sprintf("hello-%d-%d", i, i)] = true
			body, _ := json.Marshal(in)

			err := conn.WriteMessage(websocket.BinaryMessage, body)

			c.So(err, c.ShouldBeNil)
			_, message, err := conn.ReadMessage()
			c.So(err, c.ShouldBeNil)
			reply := &hello.HelloReply{}
			err = json.Unmarshal(message, reply)
			c.So(err, c.ShouldBeNil)
			c.So(reply.Message, c.ShouldEqual, fmt.Sprintf("hello-%d-%d", i, i))

		}
		err = conn.Close()
		c.So(err, c.ShouldBeNil)
	})
}
func testClientStreamRequest(t *testing.T) {
	c.Convey("test client stream request", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/client/stream", gwPort)
		wg := &sync.WaitGroup{}
		reader, writer := io.Pipe()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer writer.Close()

			for i := 0; i < 10; i++ {
				in := &hello.HelloRequest{
					Name: "world",
					Msg:  fmt.Sprintf("hello-%d", i),
				}
				data, _ := json.Marshal(in)
				err := binary.Write(writer, binary.BigEndian, uint32(len(data)))
				if err != nil {
					t.Error(err)
					return
				}
				_, err = writer.Write(data)
				if err != nil {
					t.Error(err)
					return
				}
			}
		}()
		req, err := http.NewRequest(http.MethodPost, url, reader)
		c.So(err, c.ShouldBeNil)
		req.Header.Set("Content-Type", ClientStreamContentType)
		req.Header.Set("accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		wg.Wait()
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)

		reply := &hello.RepeatedReply{}
		err = json.Unmarshal(data, reply)

		c.So(err, c.ShouldBeNil)
		c.So(len(reply.Replies), c.ShouldEqual, 10)
		for i := 0; i < 10; i++ {
			c.So(reply.Replies[i].Message, c.ShouldEqual, fmt.Sprintf("hello-%d-%d", i, i))
		}
	})
}
func testClientStreamRequestByMarshal(t *testing.T) {
	c.Convey("test client stream request", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/client/stream", gwPort)
		wg := &sync.WaitGroup{}
		reader, writer := io.Pipe()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer writer.Close()

			for i := 0; i < 10; i++ {
				in := &hello.HelloRequest{
					Name: "world",
					Msg:  fmt.Sprintf("hello-%d", i),
				}
				marshaler := NewProtobufWithLengthPrefix()
				encoder := marshaler.NewEncoder(writer)
				err := encoder.Encode(in)
				if err != nil {
					t.Error(err)
					return
				}

			}
		}()
		req, err := http.NewRequest(http.MethodPost, url, reader)
		c.So(err, c.ShouldBeNil)
		req.Header.Set("Content-Type", NewProtobufWithLengthPrefix().ContentType(nil))
		req.Header.Set("accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		wg.Wait()
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)

		reply := &hello.RepeatedReply{}
		err = json.Unmarshal(data, reply)

		c.So(err, c.ShouldBeNil)
		c.So(len(reply.Replies), c.ShouldEqual, 10)
		for i := 0; i < 10; i++ {
			c.So(reply.Replies[i].Message, c.ShouldEqual, fmt.Sprintf("hello-%d-%d", i, i))
		}

		marshal := NewProtobufWithLengthPrefix()
		helloReq := &hello.HelloRequest{Msg: "hello", Name: "world"}
		patch := gomonkey.ApplyFuncReturn(protojson.Marshal, nil, fmt.Errorf("test json marshal error"))
		defer patch.Reset()
		_, err = marshal.Marshal(helloReq)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test json marshal error")
		writer2 := &bytes.Buffer{}
		encoder := marshal.NewEncoder(writer2)
		in := &hello.HelloRequest{
			Name: "world",
			Msg:  "hello",
		}
		err = encoder.Encode(in)
		c.So(err, c.ShouldNotBeNil)
		patch.Reset()

		err = encoder.Encode(in)
		c.So(err, c.ShouldBeNil)
		decoder := marshal.NewDecoder(writer2)
		patch2 := gomonkey.ApplyFuncReturn(binary.Read, fmt.Errorf("test json read error"))
		defer patch2.Reset()
		out := &hello.HelloRequest{}
		err = decoder.Decode(out)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test json read error")
		patch2.Reset()

		patch3 := gomonkey.ApplyFuncReturn((*bytes.Buffer).Read, 0, fmt.Errorf("test io read error"))
		defer patch3.Reset()
		err = decoder.Decode(out)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test io read error")
		patch3.Reset()

	})
}
func testDeleteEndpoint(t *testing.T) {
	c.Convey("test delete endpoint", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		err = gw.DeleteHandlerClient(context.TODO(), pd)
		c.So(err, c.ShouldBeNil)
		gw.DeleteLoadBalance(pd)
		example.Stop()
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		c.So(err, c.ShouldBeNil)

		resp, err := http.DefaultClient.Do(r)

		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusNotFound)
		for _, ep := range eps {
			c.So(ep.Close(), c.ShouldBeNil)
		}
	})
}
func testRegisterLocalService(t *testing.T) {
	var pd ProtobufDescription
	var gwPort int
	var localGW *GatewayServer
	c.Convey("test register local service", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err = NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd-2"))
		c.So(err, c.ShouldBeNil)
		exampleServer := example.NewExampleServer()

		rander := rand.New(rand.NewSource(time.Now().Unix())) // 初始化随机数种子
		min := 1949
		max := 12138
		randomNumber = rander.Intn(max-min+1) + min
		// randomNumber := 31949
		gwPort = randomNumber
		// gw = newTestServer(gwPort, randomNumber)
		opts, cnf := newTestServer(gwPort, randomNumber)
		localGW = NewGateway(cnf, opts)

		err = localGW.RegisterLocalService(context.Background(), pd, exampleServer.Desc(), exampleServer)
		c.So(err, c.ShouldBeNil)
		go localGW.Start()
		time.Sleep(3 * time.Second)
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		c.So(err, c.ShouldBeNil)

		resp, err := http.DefaultClient.Do(r)

		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusOK)

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		c.So(err, c.ShouldBeNil)
		rsp := &hello.HelloReply{}

		err = json.Unmarshal(body, rsp)
		c.So(err, c.ShouldBeNil)

		c.So(rsp.Message, c.ShouldEqual, "hello")
		c.So(rsp.Name, c.ShouldEqual, "world")

		err = localGW.RegisterLocalService(context.Background(), pd, exampleServer.Desc(), exampleServer)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "already exists")

	})
	c.Convey("test del local service", t, func() {
		localGW.DeleteLocalService(pd)
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		// t.Logf("local server url:%s", url)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		c.So(err, c.ShouldBeNil)

		resp, err := http.DefaultClient.Do(r)

		c.So(err, c.ShouldBeNil)
		// c.So(resp.ContentLength, c.ShouldBeGreaterThan, 0)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusNotFound)
	})
}
func testLoadGlobalTypes(t *testing.T) {
	c.Convey("test load global types", t, func() {

		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata")
		os.Remove(filepath.Join(pbFile, "desc.pb"))
		os.Remove(filepath.Join(pbFile, "json"))
		pd, err := NewDescription(pbFile)
		c.So(err, c.ShouldBeNil)
		_ = pd.SetHttpResponse(common.E_HttpResponse)
		err = gw.RegisterHandlerClient(context.Background(), pd)
		c.So(err, c.ShouldBeNil)

		mt, err := protoregistry.GlobalTypes.FindMessageByName("integration.TestRequest")
		c.So(err, c.ShouldBeNil)
		c.So(mt, c.ShouldNotBeNil)

		et, err := protoregistry.GlobalTypes.FindEnumByName("integration.TestStaus")
		c.So(err, c.ShouldBeNil)
		c.So(et, c.ShouldNotBeNil)
		err = pd.SetHttpResponse(common.E_HttpResponse)
		c.So(err, c.ShouldBeNil)
	})
}
func testRequestError(t *testing.T) {
	c.Convey("test request error", t, func() {
		errChan := make(chan error, 1)

		cases := []struct {
			patch  interface{}
			output []interface{}
		}{
			{
				patch:  (*gwRuntime.DecoderWrapper).Decode,
				output: []interface{}{fmt.Errorf("test json decode error")},
			},
			{
				patch:  (*HttpEndpointImpl).inParamsHandle,
				output: []interface{}{fmt.Errorf("test inParamsHandle error")},
			},
			{
				patch:  (*HttpEndpointImpl).addHexEncodeSHA256HashV2,
				output: []interface{}{fmt.Errorf("test io copy error")},
			},
			{
				patch:  gwRuntime.AnnotateContext,
				output: []interface{}{context.Background(), fmt.Errorf("test AnnotateContext error")},
			},
			{
				patch:  grpc.MethodFromServerStream,
				output: []interface{}{"", false},
			},
			{
				patch:  peer.FromContext,
				output: []interface{}{nil, false},
			},
			{
				patch:  (*GrpcProxy).getXForward,
				output: []interface{}{fmt.Errorf("test getXForward error")},
			},
			{
				patch:  (*GrpcLoadBalancer).Select,
				output: []interface{}{nil, fmt.Errorf("test select error")},
			},
			{
				patch:  (*GrpcProxy).forwardServerToClient,
				output: []interface{}{errChan},
			},
		}
		errChan <- fmt.Errorf("test forwardServerToClient error")
		defer close(errChan)
		for i, caseV := range cases {
			t.Logf("test error test case:%d", i)
			url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			r.Header.Set("x-uid", "12345678")
			c.So(err, c.ShouldBeNil)
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()

			resp, err := http.DefaultClient.Do(r)

			patch.Reset()
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldBeGreaterThanOrEqualTo, http.StatusBadRequest)

			defer resp.Body.Close()
		}
	})
}
func testHttpError(t *testing.T) {
	c.Convey("test http error", t, func() {
		// url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/error?code=%d&msg=ok", codes.OK, gwPort)
		cases := []struct {
			code               int32
			msg                string
			expectHttpCode     int
			expectInternalCode int32
		}{
			{
				code:               int32(codes.OK),
				msg:                "ok",
				expectHttpCode:     http.StatusOK,
				expectInternalCode: int32(common.Code_OK),
			},
			{
				code:               int32(codes.Internal),
				msg:                codes.Internal.String(),
				expectHttpCode:     http.StatusInternalServerError,
				expectInternalCode: int32(common.Code_INTERNAL_ERROR),
			},
			{
				code:               int32(codes.InvalidArgument),
				msg:                codes.InvalidArgument.String(),
				expectHttpCode:     http.StatusBadRequest,
				expectInternalCode: int32(common.Code_PARAMS_ERROR),
			},
			{
				code:               int32(codes.NotFound),
				msg:                codes.NotFound.String(),
				expectHttpCode:     http.StatusNotFound,
				expectInternalCode: int32(common.Code_NOT_FOUND),
			},
			{
				code:               int32(codes.PermissionDenied),
				msg:                codes.PermissionDenied.String(),
				expectHttpCode:     http.StatusForbidden,
				expectInternalCode: int32(common.Code_AUTH_ERROR),
			},
			{
				code:               int32(codes.Unauthenticated),
				msg:                codes.Unauthenticated.String(),
				expectHttpCode:     http.StatusUnauthorized,
				expectInternalCode: int32(common.Code_AUTH_ERROR),
			},
			{
				code:               int32(codes.ResourceExhausted),
				msg:                codes.ResourceExhausted.String(),
				expectHttpCode:     http.StatusTooManyRequests,
				expectInternalCode: int32(common.Code_RESOURCE_EXHAUSTED),
			},
			{
				code:               int32(codes.DeadlineExceeded),
				msg:                codes.DeadlineExceeded.String(),
				expectHttpCode:     http.StatusGatewayTimeout,
				expectInternalCode: int32(common.Code_TIMEOUT_ERROR),
			},
			{
				code:               int32(99999),
				msg:                codes.Internal.String(),
				expectHttpCode:     http.StatusInternalServerError,
				expectInternalCode: int32(common.Code_INTERNAL_ERROR),
			},
		}
		for _, v := range cases {
			url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/error/test?msg=ok&code=%d", gwPort, v.code)
			r, err := http.NewRequest(http.MethodGet, url, nil)
			c.So(err, c.ShouldBeNil)
			c.So(r, c.ShouldNotBeNil)
			resp, err := http.DefaultClient.Do(r)
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldEqual, v.expectHttpCode)

		}
	})
}
func testWebSocketError(t *testing.T) {
	c.Convey("test websocket error", t, func() {
		url := fmt.Sprintf("ws://127.0.0.1:%d/api/v1/example/server/websocket", gwPort)
		i := 0
		in := &hello.HelloRequest{
			Name: "world",
			Msg:  fmt.Sprintf("hello-%d", i),
		}
		body, _ := json.Marshal(in)
		cases := []struct {
			patch     interface{}
			output    []interface{}
			exceptErr error
		}{
			{
				patch:  (*httpForwardGrpcEndpointImpl).Stream,
				output: []interface{}{nil, fmt.Errorf("test stream error")},
			},
			{
				patch:  (*streamClient).Header,
				output: []interface{}{nil, fmt.Errorf("test header error")},
			},
			{
				patch:  (*websocketForwarder).NextReader,
				output: []interface{}{nil, fmt.Errorf("test send error")},
			},
			{
				patch:  (*BinaryDecoder).Decode,
				output: []interface{}{fmt.Errorf("test BIN DECODE error")},
			},
			{
				patch:  (*streamClient).Send,
				output: []interface{}{fmt.Errorf("test send error")},
			},
			{
				patch:     (*websocket.Upgrader).Upgrade,
				output:    []interface{}{nil, fmt.Errorf("test upgrade error")},
				exceptErr: websocket.ErrBadHandshake,
			},
			{
				patch:  (*goloadbalancer.ConnPool).Get,
				output: []interface{}{nil, fmt.Errorf("test get conn error")},
			},
			{
				patch:  (*grpc.ClientConn).NewStream,
				output: []interface{}{nil, fmt.Errorf("test new stream error")},
			},
		}
		for _, caseV := range cases {
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			conn, _, err := websocket.DefaultDialer.Dial(url, nil)
			c.So(err, c.ShouldEqual, caseV.exceptErr)
			if err != nil {
				patch.Reset()
				continue
			}
			defer conn.Close()
			err = conn.WriteMessage(websocket.BinaryMessage, body)
			c.So(err, c.ShouldBeNil)
			_, _, err = conn.ReadMessage()

			c.So(err, c.ShouldNotBeNil)
			err = conn.Close()
			c.So(err, c.ShouldBeNil)
			patch.Reset()

		}

	})
}
func testServerSideEventErr(t *testing.T) {
	c.Convey("test server side event error", t, func() {
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  (*HttpEndpointImpl).inParamsHandle,
				output: []interface{}{fmt.Errorf("test inParamsHandle error")},
				err:    fmt.Errorf("test inParamsHandle error"),
			},
			{
				patch:  (*httpForwardGrpcEndpointImpl).ServerSideStream,
				output: []interface{}{nil, fmt.Errorf("test ServerSideStream error")},
				err:    fmt.Errorf("test ServerSideStream error"),
			},
			{
				patch:  (*serverSideStreamClient).Header,
				output: []interface{}{nil, fmt.Errorf("test header error")},
				err:    fmt.Errorf("test header error"),
			},
			{
				patch:  (*goloadbalancer.ConnPool).Get,
				output: []interface{}{nil, fmt.Errorf("test get conn error")},
				err:    fmt.Errorf("test get conn error"),
			},
			{
				patch:  (*grpc.ClientConn).NewStream,
				output: []interface{}{nil, fmt.Errorf("test new stream error")},
				err:    fmt.Errorf("test new stream error"),
			},
			{
				patch:  (*serverSideStreamClient).SendMsg,
				output: []interface{}{fmt.Errorf("test send error")},
				err:    fmt.Errorf("test send error"),
			},
			{
				patch:  protojson.Marshal,
				output: []interface{}{nil, fmt.Errorf("test marshal error")},
				err:    fmt.Errorf("test marshal error"),
			},
			{
				patch:  (*gwRuntime.DecoderWrapper).Decode,
				output: []interface{}{fmt.Errorf("test decode error")},
				err:    fmt.Errorf("test decode error"),
			},
		}
		for _, caseV := range cases {
			url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/server/sse/world?msg=hello", gwPort)

			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			client := sse.NewClient(url, func(c *sse.Client) {
				c.ReconnectStrategy = &backoff.StopBackOff{}
				// c.ReconnectStrategy = nil
			})

			err := client.Subscribe("message", func(msg *sse.Event) {

			})
			patch.Reset()
			c.So(err, c.ShouldNotBeNil)
			// c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
		}
	})
}
func testClientStreamErr(t *testing.T) {
	c.Convey("test client stream error", t, func() {
		cases := []struct {
			patch  interface{}
			output []interface{}
		}{
			{
				patch:  (*httpForwardGrpcEndpointImpl).ClientSideStream,
				output: []interface{}{nil, fmt.Errorf("test ClientSideStream error")},
			},
			{
				patch: (*HttpProtobufStreamImpl).NewDecoder,
				output: []interface{}{gwRuntime.DecoderFunc(func(value interface{}) error {
					return fmt.Errorf("test NewDecoder error")
				})},
			},
			{
				patch:  (*HttpEndpointImpl).inParamsHandle,
				output: []interface{}{fmt.Errorf("test inParamsHandle error")},
			},
			{
				patch:  (*clientSideStreamClient).Send,
				output: []interface{}{fmt.Errorf("test client stream send error")},
			},
			{
				patch:  (*clientSideStreamClient).CloseSend,
				output: []interface{}{fmt.Errorf("test client stream close error")},
			},
			{
				patch:  (*clientSideStreamClient).Header,
				output: []interface{}{nil, fmt.Errorf("test header error")},
			},
			{
				patch:  (*goloadbalancer.ConnPool).Get,
				output: []interface{}{nil, fmt.Errorf("test get conn error")},
			},
			{
				patch:  (*grpc.ClientConn).NewStream,
				output: []interface{}{nil, fmt.Errorf("test new stream error")},
			},
			{
				patch:  protojson.Unmarshal,
				output: []interface{}{fmt.Errorf("test unmarshal error")},
			},
		}
		for _, caseV := range cases {
			url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/client/stream", gwPort)
			wg := &sync.WaitGroup{}
			reader, writer := io.Pipe()
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer writer.Close()

				for i := 0; i < 1; i++ {
					in := &hello.HelloRequest{
						Name: "world",
						Msg:  fmt.Sprintf("hello-%d", i),
					}
					data, _ := json.Marshal(in)
					err := binary.Write(writer, binary.BigEndian, uint32(len(data)))
					if err != nil {
						t.Error(err)
						return
					}
					_, err = writer.Write(data)
					if err != nil {
						t.Error(err)
						return
					}
				}
			}()
			req, err := http.NewRequest(http.MethodPost, url, reader)
			c.So(err, c.ShouldBeNil)
			req.Header.Set("Content-Type", ClientStreamContentType)
			req.Header.Set("accept", "application/json")
			resp, err := http.DefaultClient.Do(req)
			patch.Reset()
			c.So(err, c.ShouldBeNil)
			c.So(resp.StatusCode, c.ShouldBeGreaterThanOrEqualTo, http.StatusBadRequest)
			data, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			t.Logf("data:%s", data)
			wg.Wait()
		}
	})
}

func testLoadHttpEndpointItemErr(t *testing.T) {
	c.Convey("test load http endpoint item", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		cases := []struct {
			patch  interface{}
			output []interface{}
		}{
			{
				patch:  os.Open,
				output: []interface{}{nil, fmt.Errorf("test open error")},
			},
			{
				patch:  io.ReadAll,
				output: []interface{}{nil, fmt.Errorf("test read error")},
			},
			{
				patch:  json.Unmarshal,
				output: []interface{}{fmt.Errorf("test unmarshal error")},
			},
		}
		for index, caseV := range cases {
			rander := rand.New(rand.NewSource(time.Now().Unix())) // 初始化随机数种子
			min := 1949
			max := 12138
			randomNumber = rander.Intn(max-min+1) + min
			// randomNumber := 31949
			gwPort := randomNumber
			// gw = newTestServer(gwPort, randomNumber)
			opts, cnf := newTestServer(gwPort, randomNumber)
			localGW := NewGateway(cnf, opts)
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			err := localGW.RegisterHandlerClient(context.Background(), pd)
			t.Logf("err test case:%d", index)
			c.So(err, c.ShouldNotBeNil)
			patch.Reset()
		}

	})
}

func testUpdateLoadbalance(t *testing.T) {
	c.Convey("test update loadbalance", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		helloAddr1 := fmt.Sprintf("127.0.0.1:%d", randomNumber+4)
		helloAddr2 := fmt.Sprintf("127.0.0.1:%d", randomNumber+5)
		helloAddr3 := fmt.Sprintf("127.0.0.1:%d", randomNumber+6)
		go example.Run(helloAddr1)
		time.Sleep(1 * time.Second)
		go example.Run(helloAddr2)
		time.Sleep(1 * time.Second)

		go example.Run(helloAddr3)
		time.Sleep(1 * time.Second)
		endps, err := NewLoadBalanceEndpoint(loadbalance.WRRBalanceType, []*api.EndpointMeta{{
			Addr:   helloAddr1,
			Weight: 1,
		},
			{
				Addr:   helloAddr2,
				Weight: 3,
			},
			{
				Addr:   helloAddr3,
				Weight: 2,
			},
		})
		for _, v := range endps {
			c.So(v.Addr(), c.ShouldNotEqual, fmt.Sprintf("127.0.0.1:%d", randomNumber+2))
			c.So(v.Stats().GetActivateConns(), c.ShouldEqual, 0)
		}
		eps = endps
		c.So(err, c.ShouldBeNil)
		load, err := loadbalance.New(loadbalance.WRRBalanceType, endps)
		c.So(err, c.ShouldBeNil)
		gw.UpdateLoadbalance(pd, load)
		wg := &sync.WaitGroup{}
		output := make(chan int, 10)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(output chan int) {
				defer wg.Done()
				url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
				r, _ := http.NewRequest(http.MethodGet, url, nil)
				r.Header.Set("x-uid", "12345678")

				resp, err := http.DefaultClient.Do(r)
				if err != nil {
					t.Errorf("test update loadbalance error:%v", err)
					output <- 500
					return
				}
				output <- resp.StatusCode
			}(output)
		}
		wg.Wait()
		close(output)
		for v := range output {
			c.So(v, c.ShouldEqual, http.StatusOK)
		}
		for _, v := range endps {
			c.So(v.Stats().GetIdleConns(), c.ShouldBeGreaterThan, 0)
		}
	})
}
func testStartErr(t *testing.T) {
	c.Convey("test start error", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		opts, cnf := newTestServer(0, 0)
		localGW := NewGateway(cnf, opts)
		err = localGW.RegisterHandlerClient(context.Background(), pd)
		c.So(err, c.ShouldBeNil)

		c.So(localGW.Start, c.ShouldPanic)

		min := 1949
		max := 12138
		rander := rand.New(rand.NewSource(time.Now().Unix())) // 初始化随机数种子

		randomNumber := rander.Intn(max-min+1) + min
		opts1, cnf1 := newTestServer(randomNumber+3, randomNumber)
		localGW1 := NewGateway(cnf1, opts1)
		err = localGW1.RegisterHandlerClient(context.Background(), pd)
		c.So(err, c.ShouldBeNil)
		go localGW1.Start()
		time.Sleep(3 * time.Second)
		opts2, cnf2 := newTestServer(gwPort, randomNumber+4)
		localGW2 := NewGateway(cnf2, opts2)
		err = localGW2.RegisterHandlerClient(context.Background(), pd)
		c.So(err, c.ShouldBeNil)
		c.So(localGW2.Start, c.ShouldPanic)

	})
}
func testAddHexEncodeSHA256HashV2Err(t *testing.T) {
	c.Convey("test add hex encode sha256 hash v2 err", t, func() {
		httpEp := &HttpEndpointImpl{}
		err := httpEp.addHexEncodeSHA256HashV2(nil)
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(io.Copy, int64(0), fmt.Errorf("test io copy error"))
		defer patch.Reset()
		req, err := http.NewRequest(http.MethodPost, "http://www.example.com", strings.NewReader("hello"))
		c.So(err, c.ShouldBeNil)
		err = httpEp.addHexEncodeSHA256HashV2(req)
		c.So(err.Error(), c.ShouldContainSubstring, "test io copy error")
		patch.Reset()
	})
}

func testRegisterHandlerClientErr(t *testing.T) {
	c.Convey("test register handler client err", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(protodesc.NewFiles, nil, fmt.Errorf("test NewFiles error"))
		defer patch.Reset()
		err = gw.RegisterHandlerClient(context.Background(), pd)
		patch.Reset()

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test NewFiles error")

	})
}
func testPanicRecover(t *testing.T) {
	c.Convey("test panic recover", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello&test=panic", gwPort)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		r.Header.Set("x-uid", "12345678")
		r.Header.Set(XAccessKey, "12345678")
		// r.Header.Set("Origin", "http://www.example.com")
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFunc((*MiddlewaresTest).StreamInterceptor, func(_ *MiddlewaresTest, srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			panic("test panic")
		})
		defer patch.Reset()
		resp, err := http.DefaultClient.Do(r)
		c.So(err, c.ShouldBeNil)
		c.So(resp.StatusCode, c.ShouldEqual, http.StatusInternalServerError)
		patch.Reset()
	})
}
func TestHttp(t *testing.T) {
	t.Run("testRegisterClient", testRegisterClient)
	t.Run("testRequestGet", testRequestGet)
	t.Run("testPanicRecover", testPanicRecover)
	t.Run("testCors", testCors)
	t.Run("testRequestPost", testRequestPost)
	t.Run("testServerSideEvent", testServerSideEvent)
	t.Run("testWebsocket", testWebsocket)
	t.Run("testClientStreamRequest", testClientStreamRequest)
	t.Run("testClientStreamRequestByMarshal", testClientStreamRequestByMarshal)
	t.Run("testLoadGlobalTypes", testLoadGlobalTypes)
	t.Run("testUpdateLoadbalance", testUpdateLoadbalance)
	t.Run("testHttpError", testHttpError)
	t.Run("testWebSocketError", testWebSocketError)
	t.Run("testServerSideEventErr", testServerSideEventErr)
	t.Run("testClientStreamErr", testClientStreamErr)
	t.Run("testAddHexEncodeSHA256HashV2Err", testAddHexEncodeSHA256HashV2Err)
	t.Run("testRegisterHandlerClientErr", testRegisterHandlerClientErr)
	t.Run("testRequestError", testRequestError)
	t.Run("testLoadHttpEndpointItemErr", testLoadHttpEndpointItemErr)
	t.Run("testDeleteEndpoint", testDeleteEndpoint)
	t.Run("testRegisterLocalService", testRegisterLocalService)
	t.Run("testStartErr", testStartErr)

	// time.Sleep(30 * time.Second)
}
