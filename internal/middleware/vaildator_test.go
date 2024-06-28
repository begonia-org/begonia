package middleware_test

import (
	"context"
	"testing"

	"github.com/begonia-org/begonia/internal/middleware"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestValidatorUnaryInterceptor(t *testing.T) {
	c.Convey("test validator unary interceptor", t, func() {
		validator := middleware.NewParamsValidator()

		_, err := validator.UnaryInterceptor(context.Background(), &hello.HelloRequestWithValidator{
			Name: "test",
			Msg:  "test",
			Sub: &hello.HelloSubRequest{
				SubMsg:     "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "msg", "sub"}},
		}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "validation failed")

		validator.SetPriority(1)
		c.So(validator.Priority(), c.ShouldEqual, 1)
		c.So(validator.Name(), c.ShouldEqual, "ParamsValidator")

		_, err = validator.UnaryInterceptor(context.Background(), &hello.HelloRequestWithValidator{
			Name: "test",
			Msg:  "test",
			Sub: &hello.HelloSubRequest{
				SubMsg:     "test",
				SubName:    "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
			},
			Subs: []*hello.HelloSubRequest{
				{
					SubMsg:     "test",
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
				},
				{
					SubMsg:     "test",
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "msg", "sub", "subs"}},
		}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		t.Log(err.Error())
		_, err = validator.UnaryInterceptor(context.Background(), &hello.HelloRequestWithValidator{
			Name: "test",
			Msg:  "test",
			Sub: &hello.HelloSubRequest{
				SubMsg:     "test",
				SubName:    "test",
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
			},
			Subs: []*hello.HelloSubRequest{
				{
					SubMsg:     "test",
					SubName:    "test",
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"sub_name", "sub_msg"}},
				},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "msg", "sub", "subs"}},
		}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)
		_, err = validator.UnaryInterceptor(context.Background(), &struct{}{}, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)
	})
}
func TestRequireIf(t *testing.T) {
	c.Convey("test require if", t, func() {
		v := []struct {
			Field  string `validate:"required_if=Field2"`
			Field2 string
			Field3 string `validate:"required_if=Field2"`
		}{{
			Field:  "",
			Field2: "test",
			Field3: "",
		},
		{
			Field: "",
			Field2: "",
			Field3: "",
		},
	}
		validator := middleware.NewParamsValidator()

		err:=validator.ValidateParams(v[0])
		c.So(err, c.ShouldBeNil)
		err=validator.ValidateParams(v[1])
		c.So(err,c.ShouldNotBeNil)
	})

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
