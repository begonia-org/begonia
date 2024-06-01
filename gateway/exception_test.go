package gateway

import (
	"context"
	"fmt"
	"testing"

	"github.com/begonia-org/go-sdk/logger"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
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
