package middleware_test

import (
	"context"
	"testing"

	"github.com/begonia-org/begonia/internal/middleware"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/go-playground/validator/v10"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type HelloSubRequest struct {
	SubMsg string `protobuf:"bytes,1,opt,name=sub_msg,proto3" json:"sub_msg,omitempty"`
	// @gotags: validate:"required"
	SubName string `protobuf:"bytes,2,opt,name=sub_name,proto3" json:"sub_name,omitempty" validate:"required"`
	// @gotags: validate:"required,gte=18,lte=35"
	SubAge     int32                  `protobuf:"varint,4,opt,name=sub_age,proto3" json:"sub_age,omitempty" validate:"required,gte=18,lte=35"`
	UpdateMask *fieldmaskpb.FieldMask `protobuf:"bytes,3,opt,name=update_mask,proto3" json:"update_mask,omitempty"`
}

type HelloRequestWithValidator struct {

	// @gotags: validate:"required"
	Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty" validate:"required"`
	// @gotags: validate:"required"
	Name string           `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty" validate:"required"`
	Age  int32            `protobuf:"varint,3,opt,name=age,proto3" json:"age,omitempty" validate:"required,gte=18,lte=35"`
	Sub  *HelloSubRequest `protobuf:"bytes,4,opt,name=sub,proto3" json:"sub,omitempty"`
	// @gotags: validate:"required,dive"
	Subs       []*HelloSubRequest     `protobuf:"bytes,5,rep,name=subs,proto3" json:"subs,omitempty" validate:"required,dive"`
	UpdateMask *fieldmaskpb.FieldMask `protobuf:"bytes,6,opt,name=update_mask,proto3" json:"update_mask,omitempty"`
	// @gotags: validate:"required,dive"
	SubMap map[string]*HelloSubRequest `protobuf:"bytes,7,rep,name=sub_map,proto3" json:"sub_map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3" validate:"required,dive"`
	// @gotags: validate:"required"
	SubMap2 map[string]string `protobuf:"bytes,8,rep,name=sub_map2,proto3" json:"sub_map2,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3" validate:"required"`
}

func TestValidateDynamicProtoMessage(t *testing.T) {
	c.Convey("test dynamic proto message", t, func() {
		req := &hello.HelloRequestWithValidator{
			Name:     "test",
			Msg:      "test",
			Age:      16,
			FloatNum: 0.0,
			BoolData: true,
			ExEnum:   hello.ExampleEnum_EX_RUNNING,
			ExEnums: []hello.ExampleEnum{
				hello.ExampleEnum_EX_RUNNING,
			},
			EnumMap: map[string]hello.ExampleEnum{
				"test": hello.ExampleEnum_EX_RUNNING,
			},
			EnumMap2: map[string]hello.ExampleEnum{
				"test": hello.ExampleEnum_EX_UNKNOWN,
			},
			Strs: []string{"test"},
			Sub: &hello.HelloSubRequest{
				SubMsg:     "test",
				SubAge:     19,
				SubName:    "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
			},
			Subs: []*hello.HelloSubRequest{
				{
					SubAge:     19,
					SubName:    "test",
					SubMsg:     "test",
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
				},
				{
					SubName:    "test",
					SubAge:     19,
					SubMsg:     "test",
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_msg"}},
				},
			},
			SubMap: map[string]*hello.HelloSubRequest{
				"TEST1": {
					SubName:    "test",
					SubAge:     19,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_age"}},
				},
			},
			SubMap2: map[string]string{
				"TEST1": "test",
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "msg", "sub", "subs", "sub_map", "sub_map2"}},
		}

		dpb := dynamicpb.NewMessage(req.ProtoReflect().Descriptor())
		b, _ := protojson.Marshal(req)
		_ = protojson.Unmarshal(b, dpb)

		pv := middleware.NewProtobufValidate(validator.New())
		err := pv.Protobuf(dpb, common.E_Validate)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Age")
	})

}
func TestValidateProtoMessage(t *testing.T) {
	req := &hello.HelloRequestWithValidator{
		Name:      "test",
		Msg:       "test",
		Age:       19,
		Age2:      19,
		FloatNum:  1.1,
		BoolData:  true,
		BytesData: []byte("test"),
		ExEnum:    hello.ExampleEnum_EX_RUNNING,
		ExEnums: []hello.ExampleEnum{
			hello.ExampleEnum_EX_RUNNING,
		},
		EnumMap: map[string]hello.ExampleEnum{
			"test": hello.ExampleEnum_EX_RUNNING,
		},
		EnumMap2: map[string]hello.ExampleEnum{
			"test": hello.ExampleEnum_EX_UNKNOWN,
		},
		Strs: []string{"test"},
		Sub: &hello.HelloSubRequest{
			SubMsg:     "test",
			SubAge:     19,
			SubName:    "test",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
		},
		Subs: []*hello.HelloSubRequest{
			{
				SubAge:     19,
				SubName:    "test",
				SubMsg:     "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
			},
			{
				SubName:    "test",
				SubAge:     19,
				SubMsg:     "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_msg"}},
			},
		},
		SubMap: map[string]*hello.HelloSubRequest{
			"TEST1": {
				SubName:    "test",
				SubAge:     19,
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_age"}},
			},
		},
		SubMap2: map[string]string{
			"TEST1": "test",
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "msg", "sub", "subs", "sub_map", "sub_map2"}},
	}
	c.Convey("test validator unary interceptor", t, func() {
		validator := middleware.NewParamsValidator()

		validator.SetPriority(1)
		c.So(validator.Priority(), c.ShouldEqual, 1)
		c.So(validator.Name(), c.ShouldEqual, "ParamsValidator")

		_, err := validator.UnaryInterceptor(context.Background(), req, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)
		req2 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req2.Age = 16
		_, err = validator.UnaryInterceptor(context.Background(), req2, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Age")

		req3 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req3.Subs[1].SubAge = 16
		_, err = validator.UnaryInterceptor(context.Background(), req3, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Subs[1].SubAge")
		req4 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req4.SubMap["TEST1"].SubAge = 16
		_, err = validator.UnaryInterceptor(context.Background(), req4, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "SubMap[TEST1].SubAge")
		req5 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req5.Subs[0] = &hello.HelloSubRequest{
			SubName: "test2",
		}
		_, err = validator.UnaryInterceptor(context.Background(), req5, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Subs[0].SubAge")

		req6 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req6.Sub.SubAge = 16
		_, err = validator.UnaryInterceptor(context.Background(), req6, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Sub.SubAge")

		req7 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req7.Subs[1] = &hello.HelloSubRequest{
			SubAge:     19,
			SubMsg:     "test",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_age"}},
		}
		_, err = validator.UnaryInterceptor(context.Background(), req7, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Subs[1].SubName")

		req8 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req8.SubMap2 = nil
		_, err = validator.UnaryInterceptor(context.Background(), req8, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "SubMap2")

		req9 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req9.ExEnum = hello.ExampleEnum_EX_UNKNOWN
		_, err = validator.UnaryInterceptor(context.Background(), req9, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "ExEnum")
		req10 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req10.Sub = nil
		_, err = validator.UnaryInterceptor(context.Background(), req10, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Sub")

		req11 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req11.EnumMap = nil
		_, err = validator.UnaryInterceptor(context.Background(), req11, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "EnumMap")

		req12 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req12.EnumMap2 = nil
		req12.UpdateMask.Paths = append(req12.UpdateMask.Paths, "enum_map2")
		_, err = validator.UnaryInterceptor(context.Background(), req12, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)

		req14 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req14.Subs = nil
		_, err = validator.UnaryInterceptor(context.Background(), req14, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Subs")
		req15 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req15.Name = "hello"
		_, err = validator.UnaryInterceptor(context.Background(), req15, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Sub2")
		req16 := tiga.DeepCopy(req).(*hello.HelloRequestWithValidator)
		req16.Sub = nil
		_, err = validator.UnaryInterceptor(context.Background(), req16, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "Sub")

		st := struct {
			Name string `validate:"required"`
		}{}
		_, err = validator.UnaryInterceptor(context.Background(), st, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "params is not a proto.Message")

	})
	c.Convey("test NewStructFromProtobuf", t, func() {
		validate := validator.New()
		vd := middleware.NewProtobufValidate(validate)
		v := vd.NewStructFromProtobuf(tiga.DeepCopy(req).(*hello.HelloRequestWithValidator), common.E_Validate)
		err := validate.Struct(v)
		c.So(err, c.ShouldBeNil)
		err = vd.ProtobufPartialCtx(context.Background(), tiga.DeepCopy(req).(*hello.HelloRequestWithValidator), common.E_Validate)
		c.So(err, c.ShouldBeNil)

	})
}
func TestValidator(t *testing.T) {
	st := struct {
		IntNum    int  `validate:"required"`
		BoolField bool `validate:"required"`
	}{
		IntNum:    1,
		BoolField: false,
	}
	v := validator.New()
	err := v.Struct(st)
	t.Log(err)
}
func TestValidatorStreamInterceptor(t *testing.T) {
	c.Convey("test stream interceptor", t, func() {
		validator := middleware.NewParamsValidator()

		err := validator.StreamInterceptor(&hello.HelloRequestWithValidator{
			Name: "test",
			Msg:  "test",
			Sub: &hello.HelloSubRequest{
				SubMsg:     "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "msg", "sub"}},
		}, &testStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, func(srv interface{}, ss grpc.ServerStream) error {
			ss.Context()

			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
	})
}
func TestValidatorStreamClientInterceptor(t *testing.T) {
	c.Convey("test stream client interceptor", t, func() {
		validator := middleware.NewParamsValidator()

		_, err := validator.StreamClientInterceptor(context.Background(), nil, nil, "/INTEGRATION.TESTSERVICE/NOT_FOUND", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			// return middleware.NewGrpcPluginClientStream(ctx, desc, cc, method, opts...),nil
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)

	})
}
