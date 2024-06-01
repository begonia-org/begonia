package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	api "github.com/begonia-org/go-sdk/api/example/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

		pb2 := &api.HelloReply{}
		dpb2 := dynamicpb.NewMessage(pb2.ProtoReflect().Descriptor())
		r2 := bytes.NewReader([]byte(`{"message":"John Doe","name":"tester"}`))
		decoder2 := NewMaskDecoder(NewJsonDecoder(r2))
		err = decoder2.Decode(dpb2)
		c.So(err, c.ShouldBeNil)
	})
}
func TestDecodeErr(t *testing.T) {
	c.Convey("TestDecodeErr", t, func() {
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  (*runtime.DecoderWrapper).Decode,
				output: []interface{}{fmt.Errorf("test")},
				err:    fmt.Errorf("test"),
			},
			{
				patch:  json.Marshal,
				output: []interface{}{nil, fmt.Errorf("test DECODE")},
				err:    fmt.Errorf("test DECODE"),
			},
		}
		for _, caseV := range cases {
			pb := &api.ExampleMessage{}
			dpb := dynamicpb.NewMessage(pb.ProtoReflect().Descriptor())
			r := bytes.NewReader([]byte(`{"message":"John Doe","msg":{"msg":"hello world"},"allow":"DENY","repeated_msg":[{"msg":"John Doe"}]}`))
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			decoder := NewMaskDecoder(NewJsonDecoder(r))

			err := decoder.Decode(dpb)
			patch.Reset()
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
		}
		r := bytes.NewReader([]byte(`{"message":"John Doe","msg":{"msg":"hello world"},"allow":"DENY","repeated_msg":[{"msg":"John Doe"}]}`))
		decoder := NewMaskDecoder(NewJsonDecoder(r))
		mapData := make(map[string]interface{})
		err := decoder.Decode(mapData)
		c.So(err, c.ShouldBeNil)
		c.So(len(mapData), c.ShouldEqual, 0)
	})
}
