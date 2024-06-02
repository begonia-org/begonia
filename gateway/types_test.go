package gateway

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
)

func TestClientSideStreamClient(t *testing.T) {
	c.Convey("test clientSideStreamClient", t, func() {
		out := &hello.HelloReply{}
		client := &clientSideStreamClient{ClientStream: &clientStreamMock{}, out: out.ProtoReflect().Descriptor()}
		patch := gomonkey.ApplyFuncReturn((*clientStreamMock).CloseSend, fmt.Errorf("test"))
		defer patch.Reset()
		_, err := client.CloseAndRecv()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test")
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn((*clientStreamMock).RecvMsg, fmt.Errorf("test recv error"))
		defer patch2.Reset()
		_, err = client.CloseAndRecv()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test recv error")
		patch2.Reset()

	})
}
