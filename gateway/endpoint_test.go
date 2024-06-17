package gateway

import (
	"context"
	"testing"
	"time"

	v1 "github.com/begonia-org/go-sdk/api/example/v1"
	"github.com/begonia-org/go-sdk/example"
	c "github.com/smartystreets/goconvey/convey" // 别名导入
)

func TestRequest(t *testing.T) {
	go example.Run(":12148")
	time.Sleep(time.Second * 3)
	c.Convey("test request", t, func() {

		request := NewGrpcRequest(context.Background(), nil, nil, "helloworld.Greeter/SayHello", WithIn(&v1.HelloRequest{Msg: "begonia"}), WithOut(&v1.HelloReply{}))
		pool := NewGrpcConnPool("127.0.0.1:12148")
		endpoint := NewEndpoint(pool)
		reply, metadata, err := endpoint.Request(request)
		c.So(err, c.ShouldBeNil)
		c.So(reply, c.ShouldHaveSameTypeAs, &v1.HelloReply{})
		reply1 := reply.(*v1.HelloReply)
		c.So(reply1.Message, c.ShouldEqual, "begonia")
		c.So(metadata, c.ShouldNotBeNil)
	})

}
