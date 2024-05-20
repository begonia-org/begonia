package gateway_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/begonia-org/begonia/gateway"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

func TestBuildGrpcRequest(t *testing.T) {
	c.Convey("TestBuildGrpcRequest", t, func() {
		in := &hello.HelloRequest{}
		out := &hello.HelloReply{}
		httpReq, _ := http.NewRequest("GET", "http://127.0.0.1:8080", nil)

		req := gateway.NewGrpcRequest(context.Background(),
			in.ProtoReflect().Descriptor(),
			out.ProtoReflect().Descriptor(),
			"helloworld.Greeter/SayHello",
			gateway.WithGatewayCallOptions(grpc.CompressorCallOption{}),
			gateway.WithGatewayMarshaler(gateway.NewJSONMarshaler()),
			gateway.WithGatewayPathParams(map[string]string{"key": "value"}),
			gateway.WithGatewayReq(httpReq),
			gateway.WithIn(in),
			gateway.WithOut(out),
		)
		c.So(req.GetFullMethodName(), c.ShouldEqual, "helloworld.Greeter/SayHello")
		c.So(len(req.GetCallOptions()), c.ShouldEqual, 1)
		c.So(req.GetMarshaler(), c.ShouldHaveSameTypeAs, gateway.NewJSONMarshaler())
		c.So(req.GetPathParams(), c.ShouldResemble, map[string]string{"key": "value"})
		c.So(req.GetReq().URL.String(), c.ShouldEqual, httpReq.URL.String())
		c.So(req.GetIn(), c.ShouldHaveSameTypeAs, in)
		c.So(req.GetOut(), c.ShouldHaveSameTypeAs, out)
		c.So(req.GetInType(), c.ShouldEqual, in.ProtoReflect().Descriptor())
		c.So(req.GetOutType(), c.ShouldEqual, out.ProtoReflect().Descriptor())
	})
}
