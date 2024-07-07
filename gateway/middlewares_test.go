package gateway

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type responseWriter struct {
	header http.Header
}

func (r *responseWriter) Header() http.Header {
	return r.header
}
func (r *responseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (r *responseWriter) WriteHeader(int) {

}
func TestClientMessageFromCode(t *testing.T) {
	c.Convey("TestClientMessageFromCode", t, func() {
		msg := clientMessageFromCode(codes.NotFound)
		c.So(msg, c.ShouldContainSubstring, "not found")
		msg = clientMessageFromCode(codes.ResourceExhausted)
		c.So(msg, c.ShouldContainSubstring, "resource size exceeds")
		msg = clientMessageFromCode(codes.AlreadyExists)
		c.So(msg, c.ShouldContainSubstring, "already exists")
		msg = clientMessageFromCode(codes.DataLoss)
		c.So(msg, c.ShouldContainSubstring, "Unknown error")
	})
}
func TestLoggerMiddlewares(t *testing.T) {
	mid := NewLoggerMiddleware(Log)
	c.Convey("TestLoggerMiddlewares panic", t, func() {
		f := func() {
			_, _ = mid.UnaryInterceptor(context.TODO(), nil, &grpc.UnaryServerInfo{
				FullMethod: "/test",
			}, func(ctx context.Context, req any) (any, error) {
				panic("test")
			})
		}
		c.So(f, c.ShouldNotPanic)
		f2 := func() {
			_ = mid.StreamInterceptor(nil, &streamMock{ctx: context.Background()}, &grpc.StreamServerInfo{FullMethod: "/test"}, func(srv interface{}, ss grpc.ServerStream) error {
				panic("test")
			})
		}
		// f2()
		c.So(f2, c.ShouldNotPanic)

		desc := &grpc.StreamDesc{
			StreamName:    "/INTEGRATION.TESTSERVICE/GET",
			ClientStreams: true,
			ServerStreams: true,
			Handler: func(srv interface{}, ss grpc.ServerStream) error {
				panic("test painc")
			},
		}
		patch := gomonkey.ApplyFuncReturn(grpc.Method, "/INTEGRATION.TESTSERVICE/GET", true)
		defer patch.Reset()
		st, err := mid.StreamClientInterceptor(context.Background(), desc, nil, "/INTEGRATION.TESTSERVICE/GET", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &testClientStream{ctx: metadata.NewOutgoingContext(context.Background(), metadata.New(make(map[string]string)))}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(st, c.ShouldNotBeNil)

		// has request id
		st, err = mid.StreamClientInterceptor(metadata.NewOutgoingContext(context.Background(), metadata.Pairs(XRequestID, uuid.New().String())), desc, nil, "/INTEGRATION.TESTSERVICE/GET", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &testClientStream{ctx: metadata.NewOutgoingContext(context.Background(), metadata.New(make(map[string]string)))}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(st, c.ShouldNotBeNil)
		err = desc.Handler(nil, &streamMock{ctx: context.Background()})
		c.So(err, c.ShouldNotBeNil)
		// no request id
		st, err = mid.StreamClientInterceptor(metadata.NewOutgoingContext(context.Background(), metadata.Pairs(XRequestID, "test")), desc, nil, "/INTEGRATION.TESTSERVICE/GET", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &testClientStream{ctx: metadata.NewOutgoingContext(context.Background(), metadata.Pairs(XRequestID, "test"))}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(st, c.ShouldNotBeNil)
		err = desc.Handler(nil, &streamMock{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(XRequestID, "test"))})
		c.So(err, c.ShouldNotBeNil)

	})
}

func TestIncomingHeadersToMetadata(t *testing.T) {
	c.Convey("TestIncomingHeadersToMetadata", t, func() {
		req, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)
		req.Header.Add("test", "test")
		req.Header.Add("pragma", ":")
		req.Header.Add(XAccessKey, "123456")
		md := IncomingHeadersToMetadata(context.TODO(), req)
		c.So(md, c.ShouldNotBeNil)
	})
}
func TestHttpResponseBodyModify(t *testing.T) {
	c.Convey("TestHttpResponseBodyModify", t, func() {

		resp := &responseWriter{header: make(http.Header)}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(XAccessKey, "123456"))
		header := metadata.New(map[string]string{"Content-Type": "application/grpc", "Grpc-Metadata-key": "value", "Grpc-Key": "grpc-val"})
		patch := gomonkey.ApplyFuncReturn(runtime.ServerMetadataFromContext, runtime.ServerMetadata{HeaderMD: header, TrailerMD: metadata.New(make(map[string]string))}, true)
		defer patch.Reset()
		resp2 := HttpResponseBodyModify(ctx, resp, &hello.HelloReply{})
		c.So(resp2, c.ShouldBeNil)
		for k, v := range header {
			writeHttpHeaders(resp, k, v)
		}
	})
}

func TestHandleErrorWithLogger(t *testing.T) {
	c.Convey("TestHandleErrorWithLogger", t, func() {
		f := HandleErrorWithLogger(Log)
		resp := &responseWriter{header: make(http.Header)}
		req, _ := http.NewRequest("Get", "http://www.example.com", nil)
		st := status.New(codes.NotFound, "not found")
		srvErr := &common.Errors{
			Code:    int32(common.Code_NOT_FOUND),
			Message: "not found",
			Action:  "action",
			File:    "file",
			Line:    int32(0),
			Fn:      "funcName",
		}
		st, _ = st.WithDetails(srvErr)
		f(metadata.NewIncomingContext(context.Background(), metadata.Pairs(XRequestID, "123456")), &runtime.ServeMux{}, nil, resp, req, st.Err())
		st1 := status.New(codes.NotFound, "not found")
		st1, _ = st1.WithDetails(&common.APIResponse{})
		f(metadata.NewIncomingContext(context.Background(), metadata.Pairs(XRequestID, "123456")), &runtime.ServeMux{}, nil, resp, req, st1.Err())

	})
}

func TestHandleServerStreamError(t *testing.T) {
	c.Convey("TestHandleServerStreamError", t, func() {
		f := HandleServerStreamError(Log)
		st := status.New(codes.NotFound, "not found")
		srvErr := &common.Errors{
			Code:            int32(common.Code_NOT_FOUND),
			Message:         "not found",
			Action:          "action",
			File:            "file",
			Line:            int32(0),
			Fn:              "funcName",
			ToClientMessage: "not found resource",
		}
		st, _ = st.WithDetails(srvErr)
		err := f(context.Background(), st.Err())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Err().Error(), c.ShouldContainSubstring, "not found resource")

		err = f(context.Background(), status.Error(codes.Internal, "internal error"))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Err().Error(), c.ShouldContainSubstring, "Unknown error")
		err = f(context.Background(), fmt.Errorf("test error"))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Err().Error(), c.ShouldContainSubstring, "test error")

		patch := gomonkey.ApplyFuncReturn((*anypb.Any).UnmarshalTo, fmt.Errorf("test error"))
		defer patch.Reset()
		ay, _ := anypb.New(srvErr)
		st = status.New(codes.NotFound, "not found")

		st, _ = st.WithDetails(ay)
		err = f(context.Background(), st.Err())
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Err().Error(), c.ShouldContainSubstring, "The requested resource is not found.")
	})
}
