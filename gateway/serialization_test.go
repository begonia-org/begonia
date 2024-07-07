package gateway

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/genproto/googleapis/api/httpbody"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
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

		c.So(decoder3.Decode(nil), c.ShouldBeNil)

		patch1 := gomonkey.ApplyFuncReturn(proto.Marshal, nil, fmt.Errorf("proto.Marshal: nil"))
		defer patch1.Reset()
		c.So(marshal.ContentType(msg2), c.ShouldEqual, "application/octet-stream")
		patch1.Reset()

		patch2 := gomonkey.ApplyFuncReturn(proto.Unmarshal, fmt.Errorf("io.ReadAll: unexpected EOF"))
		defer patch2.Reset()
		c.So(marshal.ContentType(msg2), c.ShouldEqual, "application/octet-stream")
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
func TestJSONMarshaler(t *testing.T) {
	c.Convey("TestJSONMarshaler", t, func() {
		marshaler := NewJSONMarshaler()
		data := map[string]interface{}{
			"test": "test",
		}
		buf, err := marshaler.Marshal(data)
		c.So(err, c.ShouldBeNil)
		c.So(string(buf), c.ShouldEqual, `{"test":"test"}`)

		httpBody := &httpbody.HttpBody{
			ContentType: "application/octet-stream-test",
			Data:        []byte("test"),
		}
		msg2 := dynamicpb.NewMessage(httpBody.ProtoReflect().Descriptor()).New()
		patch := gomonkey.ApplyFuncReturn((*runtime.JSONPb).Marshal, nil, fmt.Errorf("runtime.JSONPb{}.Marshal: nil"))
		defer patch.Reset()
		_, err = marshaler.Marshal(msg2)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "runtime.JSONPb{}.Marshal: nil")
	})
}

func TestEventSourceMarshaler(t *testing.T) {
	c.Convey("TestEventSourceMarshaler", t, func() {
		marshaler := NewEventSourceMarshaler()
		cases := []struct {
			data      interface{}
			err       error
			exception string
		}{
			{
				data: map[string]interface{}{
					"result": &common.EventStream{
						Event: "test",
						Id:    1,
						Data:  "test",
						Retry: 0,
					},
				},
				err:       nil,
				exception: fmt.Sprintf("id: %d\nevent: %s\nretry: %d\ndata: %s\n", 1, "test", 0, "test"),
			},
			{
				data: &common.EventStream{
					Event: "test-data",
					Id:    1,
					Data:  "test-data",
					Retry: 0,
				},
				err:       nil,
				exception: fmt.Sprintf("id: %d\nevent: %s\nretry: %d\ndata: %s\n", 1, "test-data", 0, "test-data"),
			},
			{
				data: map[string]proto.Message{
					"error": &spb.Status{
						Message: "test error",
						Code:    int32(codes.Internal),
						Details: []*anypb.Any{},
					},
				},
				err:       nil,
				exception: fmt.Sprintf("id: %d\nevent: %s\nretry: %d\ndata: %s\n", 0, "error", 0, "test error"),
			},
		}
		for _, caseV := range cases {
			buf, err := marshaler.Marshal(caseV.data)
			c.So(err, c.ShouldBeNil)
			c.So(string(buf), c.ShouldEqual, caseV.exception)
		}
	})
}
