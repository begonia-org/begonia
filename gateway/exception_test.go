package gateway

import (
	"context"
	"fmt"
	"testing"

	"github.com/begonia-org/go-sdk/logger"
	"github.com/google/uuid"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MiddlewaresTest struct {
	log      logger.Logger
	priority int
	name     string
}

func (e *MiddlewaresTest) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	return handler(ctx, req)

}
func (e *MiddlewaresTest) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, ss)
}
func NewMiddlewaresTest(log logger.Logger) *MiddlewaresTest {
	return &MiddlewaresTest{log: log, name: "MiddlewaresTest"}
}

func (e *MiddlewaresTest) Priority() int {
	return e.priority
}
func (e *MiddlewaresTest) SetPriority(priority int) {
	e.priority = priority
}

func (e *MiddlewaresTest) Name() string {
	return e.name
}

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
func TestUnaryInterceptor(t *testing.T) {
	c.Convey("TestUnaryInterceptor", t, func() {
		mid := NewException(Log)

		ctx := context.Background()
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("test panic error")
		}
		_, err := mid.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test panic error")

		handler2 := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("test error")
		}
		_, err2 := mid.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler2)
		c.So(err2, c.ShouldNotBeNil)
		c.So(err2.Error(), c.ShouldContainSubstring, "test error")
		mid.SetPriority(2)
		c.So(mid.Priority(), c.ShouldEqual, 2)
		c.So(mid.Name(), c.ShouldEqual, "exception")

	})

}
func TestExceptionStreamClientInterceptor(t *testing.T) {
	c.Convey("TestExceptionStreamClientInterceptor", t, func() {
		mid := NewException(Log)
		ctx := context.Background()

		desc := &grpc.StreamDesc{
			StreamName:    "/INTEGRATION.TESTSERVICE/GET",
			ClientStreams: true,
			ServerStreams: true,
			Handler: func(srv interface{}, ss grpc.ServerStream) error {
				panic("test painc")
			},
		}
		st, err := mid.StreamClientInterceptor(ctx, desc, nil, "/INTEGRATION.TESTSERVICE/GET", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &testClientStream{ctx: context.Background()}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(st, c.ShouldNotBeNil)

		// has request id
		st, err = mid.StreamClientInterceptor(metadata.NewOutgoingContext(context.Background(), metadata.Pairs(XRequestID, uuid.New().String())), desc, nil, "/INTEGRATION.TESTSERVICE/GET", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &testClientStream{ctx: context.Background()}, nil
		})
		c.So(err, c.ShouldBeNil)
		c.So(st, c.ShouldNotBeNil)
		err = desc.Handler(nil, nil)
		c.So(err, c.ShouldNotBeNil)

	})
}
