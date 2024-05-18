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
	"github.com/begonia-org/begonia/gateway/serialization"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/cenkalti/backoff.v1"
)

// var endpoint HttpEndpoint
var gwPort = 9527
var gw *GatewayServer
var onceInit sync.Once
var randomNumber int

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

	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("application/json", serialization.NewJSONMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("multipart/form-data", serialization.NewFormDataMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("application/x-www-form-urlencoded", serialization.NewFormUrlEncodedMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption(gwRuntime.MIMEWildcard, serialization.NewRawBinaryUnmarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("application/octet-stream", serialization.NewRawBinaryUnmarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption("text/event-stream", serialization.NewEventSourceMarshaler()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMarshalerOption(serialization.ClientStreamContentType, serialization.NewProtobufWithLengthPrefix()))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithMetadata(IncomingHeadersToMetadata))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithErrorHandler(HandleErrorWithLogger(Log)))
	opts.HttpMiddlewares = append(opts.HttpMiddlewares, gwRuntime.WithForwardResponseOption(HttpResponseBodyModify))

	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithMaxActiveConns(100))
	opts.PoolOptions = append(opts.PoolOptions, loadbalance.WithPoolSize(128))
	loggerMid := NewLoggerMiddleware(Log)
	opts.Options = append(opts.Options, grpc.ChainStreamInterceptor(loggerMid.StreamInterceptor))
	opts.Options = append(opts.Options, grpc.ChainUnaryInterceptor(loggerMid.UnaryInterceptor))
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
		min := 9527
		max := 12138
		randomNumber = rander.Intn(max-min+1) + min
		// randomNumber := 39527
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
		time.Sleep(4 * time.Second)
	})
}
func testRequestGet(t *testing.T) {
	c.Convey("test request GET", t, func() {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/example/world?msg=hello", gwPort)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		r.Header.Set("x-uid", "12345678")
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
		req.Header.Set("Content-Type", serialization.ClientStreamContentType)
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
		min := 9527
		max := 12138
		randomNumber = rander.Intn(max-min+1) + min
		// randomNumber := 39527
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
		pd.SetHttpResponse(common.E_HttpResponse)
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
				patch:  (*GrpcProxy).getClientIP,
				output: []interface{}{"", fmt.Errorf("test getClientIP error")},
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
				patch:  (*serialization.BinaryDecoder).Decode,
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
		}{
			{
				patch:  (*HttpEndpointImpl).inParamsHandle,
				output: []interface{}{fmt.Errorf("test inParamsHandle error")},
			},
			{
				patch:  (*httpForwardGrpcEndpointImpl).ServerSideStream,
				output: []interface{}{nil, fmt.Errorf("test ServerSideStream error")},
			},
			{
				patch:  (*serverSideStreamClient).Header,
				output: []interface{}{nil, fmt.Errorf("test header error")},
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
				patch: (*serialization.HttpProtobufStreamImpl).NewDecoder,
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
		}
		for i, caseV := range cases {
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
			t.Logf("test case:%d", i)
			req, err := http.NewRequest(http.MethodPost, url, reader)
			c.So(err, c.ShouldBeNil)
			req.Header.Set("Content-Type", serialization.ClientStreamContentType)
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
			min := 9527
			max := 12138
			randomNumber = rander.Intn(max-min+1) + min
			// randomNumber := 39527
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
func TestHttp(t *testing.T) {
	t.Run("testRegisterClient", testRegisterClient)

	t.Run("testRequestGet", testRequestGet)
	t.Run("testCors", testCors)
	t.Run("testRequestPost", testRequestPost)
	t.Run("testServerSideEvent", testServerSideEvent)
	t.Run("testWebsocket", testWebsocket)
	t.Run("testClientStreamRequest", testClientStreamRequest)
	t.Run("testLoadGlobalTypes", testLoadGlobalTypes)
	t.Run("testHttpError", testHttpError)
	t.Run("testWebSocketError", testWebSocketError)
	t.Run("testServerSideEventErr", testServerSideEventErr)
	t.Run("testClientStreamErr", testClientStreamErr)
	// t.Run("testInParamsHandle", testInParamsHandle)
	t.Run("testRequestError", testRequestError)
	t.Run("testLoadHttpEndpointItemErr", testLoadHttpEndpointItemErr)
	t.Run("testDeleteEndpoint", testDeleteEndpoint)
	t.Run("testRegisterLocalService", testRegisterLocalService)
	// time.Sleep(30 * time.Second)
}
