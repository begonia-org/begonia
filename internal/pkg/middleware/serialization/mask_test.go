package serialization

import (
	"bytes"
	"testing"

	api "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestMask(t *testing.T) {
	c.Convey("TestMask", t, func() {
		pb := &api.ExampleMessage{}
		dpb := dynamicpb.NewMessage(pb.ProtoReflect().Descriptor())
		r := bytes.NewReader([]byte(`{"message":"John Doe","msg":{"msg":"hello world"},"allow":"DENY","repeated_msg":[{"msg":"John Doe"}]}`))
		decoder := NewMaskDecoder(NewJsonDecoder(r))
		err := decoder.Decode(dpb)
		c.So(err, c.ShouldBeNil)
		bData, err := protojson.Marshal(dpb)
		c.So(err, c.ShouldBeNil)
		err = protojson.Unmarshal(bData, pb)
		c.So(err, c.ShouldBeNil)
		c.So(pb.Message, c.ShouldEqual, "John Doe")
		c.So(pb.Allow, c.ShouldEqual, api.EnumAllow_DENY)
		c.So(pb.RepeatedMsg[0].Msg, c.ShouldEqual, "John Doe")
		c.So(pb.Msg.Msg, c.ShouldEqual, "hello world")
		c.So(pb.Mask.Paths, c.ShouldNotBeEmpty)
	})
}
