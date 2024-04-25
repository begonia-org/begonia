package transport

import (
	"testing"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway/register"
	c "github.com/smartystreets/goconvey/convey"
) // 别名导入

func TestNewDescription(t *testing.T) {
	c.Convey("test new description", t, func() {
		desc, err := NewDescription("./protos")
		c.So(err, c.ShouldBeNil)
		c.So(desc, c.ShouldNotBeNil)
		c.So(desc.GetFileDescriptorSet(), c.ShouldNotBeNil)
		c.So(desc.GetMessageTypeByName("helloworld", "HelloRequest"), c.ShouldNotBeNil)
		contents, err := gw.Register(desc.GetFileDescriptorSet(), false, "")
		c.So(err, c.ShouldBeNil)
		c.So(contents, c.ShouldNotBeNil)
		c.So(contents, c.ShouldNotBeEmpty)
		// for _,content:=range contents{
		// 	t.Log(content)
		// }

	})
}
