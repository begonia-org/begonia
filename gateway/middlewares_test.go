package gateway

import (
	"context"
	"net/http"
	"testing"

	hello "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
func TestClientMessageFromCode(t *testing.T){
	c.Convey("TestClientMessageFromCode",t,func(){
		msg:=clientMessageFromCode(codes.NotFound)
		c.So(msg,c.ShouldContainSubstring,"not found")
		msg = clientMessageFromCode(codes.ResourceExhausted)
		c.So(msg,c.ShouldContainSubstring,"resource size exceeds")
		msg = clientMessageFromCode(codes.AlreadyExists)
		c.So(msg,c.ShouldContainSubstring,"already exists")
		msg = clientMessageFromCode(codes.DataLoss)
		c.So(msg,c.ShouldContainSubstring,"Unknown error")
	})
}
func TestLoggerMiddlewares(t *testing.T) {
	mid :=NewLoggerMiddleware(Log)
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
			_ = mid.StreamInterceptor(nil, &streamMock{}, &grpc.StreamServerInfo{FullMethod: "/test"}, func(srv interface{}, ss grpc.ServerStream) error {
				panic("test")
			})
		}
		c.So(f2, c.ShouldNotPanic)
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
		ctx:=metadata.NewIncomingContext(context.Background(),metadata.Pairs(XAccessKey,"123456"))
		resp2 := HttpResponseBodyModify(ctx, resp, &hello.HelloReply{})
		c.So(resp2, c.ShouldBeNil)
	})
}
