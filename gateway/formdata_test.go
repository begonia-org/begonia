package gateway

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	api "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestFormDataContentType(t *testing.T) {

	c.Convey("TestFormDataContentType", t, func() {

		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)
		repeatData := []int64{1, 2, 3}
		values := map[string]string{
			"message":      "John Doe",
			"msg":          `{"msg":"hello world"}`,
			"allow":        api.EnumAllow_DENY.String(),
			"repeated_msg": `{"msg":"John Doe"}`,
			"float_num":    `1949.38`,
			"float_data":   `1949.38`,
			"bool_data":    "true",
			"code":         "1949",
			// "repeated_data":"1,2,3",
			"fixed_data":    "1949",
			"sfixed_data":   "11949",
			"sfixed32_data": "21949",
			"fixed32_data":  "31949",
		}
		for key, value := range values {
			_ = writer.WriteField(key, value)
		}
		_ = writer.WriteField("repeated_data", "1")
		_ = writer.WriteField("repeated_data", "2")
		_ = writer.WriteField("repeated_data", "3")
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
		// decoder := &FormDataDecoder{r: &requestBody, boundary: writer.Boundary()}
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
		c.So(pb.FloatNum, c.ShouldEqual, 1949.38)
		c.So(pb.BoolData, c.ShouldEqual, true)
		c.So(pb.Code, c.ShouldEqual, 1949)
		c.So(pb.RepeatedData, c.ShouldResemble, repeatData)
		c.So(pb.FixedData, c.ShouldEqual, 1949)
		c.So(pb.SfixedData, c.ShouldEqual, 11949)
		c.So(pb.Sfixed32Data, c.ShouldEqual, 21949)
		c.So(pb.Fixed32Data, c.ShouldEqual, 31949)
		marshal:=&FormDataMarshaler{}
		c.So(marshal.ContentType(nil),c.ShouldEqual,"multipart/form-data")
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
		marshal:=&FormUrlEncodedMarshaler{}
		c.So(marshal.ContentType(nil),c.ShouldEqual,"application/x-www-form-urlencoded")
	})
}

func TestFormUrlEncodedErr(t *testing.T) {
	c.Convey("TestFormUrlEncodedErr", t, func() {
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  io.ReadAll,
				err:    fmt.Errorf("read error"),
				output: []interface{}{nil, fmt.Errorf("read error")},
			},
			{
				patch:  parseFormToProto,
				err:    fmt.Errorf("parseFormToProto error"),
				output: []interface{}{fmt.Errorf("parseFormToProto error")},
			},
		}
		for _, caseV := range cases {
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
			patch:=gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			err := decoder.Decode(dpb)
			c.So(err, c.ShouldNotBeNil)
			patch.Reset()
		}


	})
}

func TestFormDataDecodeErr(t *testing.T) {
	c.Convey("TestFormDataDecodeErr", t, func() {
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  (*multipart.Reader).ReadForm,
				err:    fmt.Errorf("multipart: NextPart: EOF"),
				output: []interface{}{nil, fmt.Errorf("multipart: NextPart: EOF")},
			},
			{
				patch:  (*multipart.FileHeader).Open,
				err:    fmt.Errorf("open : no such file or directory"),
				output: []interface{}{nil, fmt.Errorf("open : no such file or directory")},
			},
			{
				patch:  io.ReadAll,
				err:    fmt.Errorf("read error"),
				output: []interface{}{nil, fmt.Errorf("read error")},
			},
			{
				patch:  getProtoreflectValue,
				err:    fmt.Errorf("getProtoreflectValue error"),
				output: []interface{}{nil, fmt.Errorf("getProtoreflectValue error")},
			},
		}
		for _, caseV := range cases {

			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)
			values := map[string]string{
				"message": "John Doe",
			}
			for key, value := range values {
				_ = writer.WriteField(key, value)
			}
			_, filePath, _, _ := runtime.Caller(0)

			// 添加文件
			file, _ := os.Open(filePath)

			part, _ := writer.CreateFormFile("byte_data", filepath.Base(filePath))

			_, _ = io.Copy(part, file)

			// 发送请求前必须关闭writer
			writer.Close()
			pb := &api.ExampleMessage{}
			dpb := dynamicpb.NewMessage(pb.ProtoReflect().Descriptor())
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)

			decoder := &FormDataDecoder{r: &requestBody, boundary: writer.Boundary()}
			defer patch.Reset()
			err := decoder.Decode(dpb)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
			patch.Reset()
			if file != nil {
				file.Close()

			}

		}
	})
}

func TestFormDataValueErr(t *testing.T) {
	c.Convey("test formdata value err", t, func() {
		cases := []struct {
			values map[string]string
			err    error
		}{
			{
				values: map[string]string{"bool_data": "ture"},
			},
			{
				values: map[string]string{"float_num": "test"},
			},
			{
				values: map[string]string{"float_data": "test"},
			},
			{
				values: map[string]string{"code": "test"},
			},
			{
				values: map[string]string{"allow": "test"},
			},
			{
				values: map[string]string{"repeated_data": "test"},
			},
			{
				values: map[string]string{"fixed_data": "test"},
			},
			{
				values: map[string]string{"msg": "{test"},
			},
			{
				values: map[string]string{"fixed32_data": "{test"},
			},
		}
		for _, caseV := range cases {
			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)
			for key, value := range caseV.values {
				_ = writer.WriteField(key, value)
				writer.Close()
				pb := &api.ExampleMessage{}
				dpb := dynamicpb.NewMessage(pb.ProtoReflect().Descriptor())
				encode := map[string]interface{}{"result": dpb}
				// decoder := &FormDataDecoder{r: &requestBody, boundary: writer.Boundary()}
				decoder := &FormDataDecoder{r: &requestBody, boundary: writer.Boundary()}
				err := decoder.Decode(encode)
				c.So(err, c.ShouldNotBeNil)
			}
		}
	})
}
