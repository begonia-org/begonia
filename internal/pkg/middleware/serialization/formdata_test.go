package serialization

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	api "github.com/begonia-org/go-sdk/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestFormDataContentType(t *testing.T) {

	c.Convey("TestFormDataContentType", t, func() {

		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		// 添加文本字段
		_ = writer.WriteField("message", "John Doe")
		_= writer.WriteField("msg", `{"msg":"hello world"}`)
		_ = writer.WriteField("allow", api.EnumAllow_DENY.String())
		_ = writer.WriteField("repeated_msg",`{"msg":"John Doe"}`)
		_, filePath, _, _ := runtime.Caller(0)

		// 添加文件
		file, _ := os.Open(filePath)

		defer file.Close()

		part, _ := writer.CreateFormFile("byte_data", filepath.Base(filePath))

		_, _ = io.Copy(part, file)
		// 发送请求前必须关闭writer
		writer.Close()
		pb := &api.ExampleMessage{}
		dpb := dynamicpb.NewMessage(pb.ProtoReflect().Descriptor())
		decoder := &FormDataDecoder{r: &requestBody, boundary: writer.Boundary()}
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
	})
}

func TestFormUrlEncodedContentType(t *testing.T) {
c.Convey("TestFormUrlEncodedContentType", t, func() {
	formData := url.Values{}
	formData.Add("message", "John Doe")
	formData.Add("allow", api.EnumAllow_DENY.String())
	_, filePath, _, _ := runtime.Caller(0)

	// 添加文件
	data, _ := os.ReadFile(filePath)
	formData.Add("byte_data", string(data))
	pb := &api.ExampleMessage{}
	dpb := dynamicpb.NewMessage(pb.ProtoReflect().Descriptor())
	decoder := &FormUrlEncodedDecoder{r: strings.NewReader(formData.Encode())}
	err := decoder.Decode(dpb)
	c.So(err, c.ShouldBeNil)
	bData, err := protojson.Marshal(dpb)
	c.So(err, c.ShouldBeNil)
	err = protojson.Unmarshal(bData, pb)
	c.So(err, c.ShouldBeNil)
	c.So(pb.Message, c.ShouldEqual, "John Doe")
	c.So(pb.Allow, c.ShouldEqual, api.EnumAllow_DENY)
})
}
