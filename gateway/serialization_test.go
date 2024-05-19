package gateway

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestRawBinaryUnmarshaler(t *testing.T) {
	c.Convey("TestRawBinaryUnmarshaler", t, func() {
		r := bytes.NewReader([]byte(`{"test":"test"}`))
		marshal := NewRawBinaryUnmarshaler()
		decoder := marshal.NewDecoder(r)
		val := make(map[string]string)
		err := decoder.Decode(&val)
		c.So(err, c.ShouldBeNil)

		r2 := bytes.NewReader([]byte(`"test":"test"}`))
		decoder2 := marshal.NewDecoder(r2)
		err = decoder2.Decode(&val)
		c.So(err, c.ShouldNotBeNil)
		// t.Logf("err: %v", err)

		body := &httpbody.HttpBody{
			Data: []byte("test"),
		}
		msg := dynamicpb.NewMessage(body.ProtoReflect().Descriptor()).New()

		r3 := bytes.NewReader([]byte(`test`))
		decoder3 := marshal.NewDecoder(r3)
		err = decoder3.Decode(msg)
		c.So(err, c.ShouldBeNil)
		c.So(msg.Get(msg.Descriptor().Fields().ByName("content_type")).String(), c.ShouldEqual, "application/octet-stream")
		c.So(msg.Get(msg.Descriptor().Fields().ByName("data")).Bytes(), c.ShouldEqual, []byte(`test`))

		httpBody := &httpbody.HttpBody{
			ContentType: "application/octet-stream-test",
			Data:        []byte("test"),
		}
		msg2 := dynamicpb.NewMessage(httpBody.ProtoReflect().Descriptor()).New()
		httpBodyBytes, err := proto.Marshal(httpBody)
		c.So(err, c.ShouldBeNil)
		err2 := proto.Unmarshal(httpBodyBytes, msg2.Interface())
		c.So(err2, c.ShouldBeNil)
		c.So(marshal.ContentType(msg2), c.ShouldContainSubstring, "application/octet-stream-test")

		c.So(decoder3.Decode(nil),c.ShouldBeNil)
	})
}

func TestRawBinaryDecodeErr(t *testing.T) {
	c.Convey("TestRawBinaryDecodeErr", t, func() {
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  io.ReadAll,
				output: []interface{}{[]byte{}, fmt.Errorf("io.ReadAll: unexpected EOF")},
				err:    fmt.Errorf("io.ReadAll: unexpected EOF"),
			},
			{
				patch:  io.ReadAll,
				output: []interface{}{[]byte{}, nil},
				err:    io.EOF,
			},
			{
				patch:  proto.Marshal,
				output: []interface{}{nil, fmt.Errorf("proto.Marshal: nil")},
				err:    fmt.Errorf("proto.Marshal: nil"),
			},
		}
		marshal := NewRawBinaryUnmarshaler()

		for _, caseV := range cases {
			body := &httpbody.HttpBody{
				Data: []byte("test"),
			}
			msg := dynamicpb.NewMessage(body.ProtoReflect().Descriptor()).New()

			r3 := bytes.NewReader([]byte(`test`))
			decoder3 := marshal.NewDecoder(r3)
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			err := decoder3.Decode(msg)
			patch.Reset()
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())

		}

	})
}
