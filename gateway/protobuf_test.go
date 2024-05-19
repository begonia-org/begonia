package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway/register"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
) // 别名导入

func testInitDescriptorSetErr(t *testing.T) {
	c.Convey("Test initDescriptorSet error", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
		pb, err := os.ReadFile(pbFile)
		c.So(err, c.ShouldBeNil)
		pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(protodesc.NewFiles, nil, fmt.Errorf("Error creating file descriptor"))
		defer patch.Reset()
		err = pd.(*protobufDescription).initDescriptorSet()
		c.So(err, c.ShouldNotBeNil)

	})
}

func testSetHttpResponseErr(t *testing.T) {
	c.Convey("Test SetHttpResponse error", t, func() {
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata")
		os.Remove(filepath.Join(pbFile, "desc.pb"))
		os.Remove(filepath.Join(pbFile, "json"))
		pd, err := NewDescription(pbFile)
		c.So(err, c.ShouldBeNil)
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  os.Open,
				err:    fmt.Errorf("open file error"),
				output: []interface{}{nil, fmt.Errorf("open file error")},
			},
			{
				patch:  (*json.Decoder).Decode,
				output: []interface{}{fmt.Errorf("decode error")},
			},
			{
				patch:  (*protoregistry.Files).FindDescriptorByName,
				output: []interface{}{nil, fmt.Errorf("find descriptor error")},
			},
			{
				patch:  os.Create,
				output: []interface{}{nil, fmt.Errorf("create file error")},
			},
			{
				patch:  (*json.Encoder).Encode,
				output: []interface{}{fmt.Errorf("encode error")},
			},
		}
		for _, v := range cases {
			patch := gomonkey.ApplyFuncReturn(v.patch, v.output...)
			defer patch.Reset()
			err = pd.SetHttpResponse(common.E_HttpResponse)
			c.So(err, c.ShouldNotBeNil)
			patch.Reset()
		}
	})
}
func testNewDescriptionErr(t *testing.T) {
	c.Convey("Test NewDescription error", t, func() {
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  filepath.Glob,
				err:    fmt.Errorf("read file error"),
				output: []interface{}{nil, fmt.Errorf("filepath glob error")},
			},
			{
				patch:  (*exec.Cmd).CombinedOutput,
				err:    fmt.Errorf("create file descriptor error"),
				output: []interface{}{nil, fmt.Errorf("create file descriptor error")},
			},
			{
				patch:  os.ReadFile,
				err:    fmt.Errorf("read file error"),
				output: []interface{}{nil, fmt.Errorf("read file error")},
			},
			{
				patch:  proto.Unmarshal,
				err:    fmt.Errorf("unmarshal error"),
				output: []interface{}{fmt.Errorf("unmarshal error")},
			},
			{
				patch:  protodesc.NewFiles,
				err:    fmt.Errorf("create file descriptor error"),
				output: []interface{}{nil, fmt.Errorf("create file descriptor error")},
			},
		}
		for _, v := range cases {
			patch := gomonkey.ApplyFuncReturn(v.patch, v.output...)
			defer patch.Reset()
			_, filename, _, _ := runtime.Caller(0)
			pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata")
			os.Remove(filepath.Join(pbFile, "desc.pb"))
			os.Remove(filepath.Join(pbFile, "json"))
			pd, err := NewDescription(pbFile)
			c.So(err, c.ShouldNotBeNil)
			c.So(pd, c.ShouldBeNil)
			patch.Reset()
		}
	})
}

func testNewDescriptionFromBinaryErr(t *testing.T) {
	c.Convey("Test NewDescriptionFromBinary error", t, func() {
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  proto.Unmarshal,
				err:    fmt.Errorf("proto unmarshal file error"),
				output: []interface{}{fmt.Errorf("proto unmarshal file error")},
			},
			{
				patch:  protodesc.NewFiles,
				err:    fmt.Errorf("create file descriptor error"),
				output: []interface{}{nil, fmt.Errorf("create file descriptor error")},
			},
			{
				patch:  register.Register,
				err:    fmt.Errorf("register.Register file error"),
				output: []interface{}{nil, fmt.Errorf("read file error")},
			},
			{
				patch:  os.MkdirAll,
				err:    fmt.Errorf("os.MkdirAll error"),
				output: []interface{}{fmt.Errorf("os.MkdirAll error")},
			},
			{
				patch:  os.WriteFile,
				err:    fmt.Errorf("os.WriteFile error"),
				output: []interface{}{fmt.Errorf("create file descriptor error")},
			},
		}
		for _, v := range cases {

			_, filename, _, _ := runtime.Caller(0)
			pbFile := filepath.Join(filepath.Dir(filepath.Dir(filename)), "testdata", "helloworld.pb")
			pb, err := os.ReadFile(pbFile)
			c.So(err, c.ShouldBeNil)
			patch := gomonkey.ApplyFuncReturn(v.patch, v.output...)
			defer patch.Reset()
			pd, err := NewDescriptionFromBinary(pb, filepath.Join("tmp", "test-pd"))
			c.So(err, c.ShouldNotBeNil)
			patch.Reset()
			c.So(pd, c.ShouldBeNil)
		}
	})
}

func TestProtobufDescriptionErr(t *testing.T) {
	t.Run("Test initDescriptorSet error", testInitDescriptorSetErr)
	t.Run("Test SetHttpResponse error", testSetHttpResponseErr)
	t.Run("Test NewDescription error", testNewDescriptionErr)
	t.Run("Test NewDescriptionFromBinary error", testNewDescriptionFromBinaryErr)
}
