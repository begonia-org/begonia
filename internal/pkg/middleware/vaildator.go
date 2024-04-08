package middleware

import (
	"context"
	"fmt"
	"sync"

	"github.com/begonia-org/begonia/internal/pkg/errors"
	gosdk "github.com/begonia-org/go-sdk"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type validatePluginStream struct {
	grpc.ServerStream
	// fullName string
	// plugin   gosdk.RemotePlugin
	ctx       context.Context
	validator ParamsValidator
}

var validatePluginStreamPool = &sync.Pool{
	New: func() interface{} {
		return &validatePluginStream{
			// validate: validator,
		}
	},
}

type ParamsValidator interface {
	gosdk.LocalPlugin
	ValidateParams(v interface{}) error
}

type ParamsValidatorImpl struct {
	priority int
}

func (p *validatePluginStream) Context() context.Context {
	return p.ctx
}
func (p *validatePluginStream) RecvMsg(m interface{}) error {
	err := p.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}
	err = p.validator.ValidateParams(m)
	return err

}

//	type Plugin interface {
//		SetPriority(priority int)
//		Priority() int
//		Name () string
//	}
//
//	type LocalPlugin interface {
//		Plugin
//		UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error)
//		StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
//	}
func (p *ParamsValidatorImpl) ValidateParams(v interface{}) error {
	validate := validator.New()
	err := validate.Struct(v)
	if errs, ok := err.(validator.ValidationErrors); ok {
		clientMsg := fmt.Sprintf("params %s validation failed with %v,except %s", errs[0].Field(), errs[0].Value(), errs[0].ActualTag())
		return errors.New(fmt.Errorf("params %s validation failed: %v", errs[0].Field(), errs[0].Value()), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "params_validation", errors.WithClientMessage(clientMsg))
	}
	return nil
}

func (p *ParamsValidatorImpl) SetPriority(priority int) {
	p.priority = priority
}

func (p *ParamsValidatorImpl) Priority() int {
	return p.priority
}
func (p *ParamsValidatorImpl) Name() string {
	return "ParamsValidator"
}

func (p *ParamsValidatorImpl) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	err = p.ValidateParams(req)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (p *ParamsValidatorImpl) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	validateStream := validatePluginStreamPool.Get().(*validatePluginStream)
	validateStream.ServerStream = ss
	validateStream.validator = p
	validateStream.ctx = ss.Context()
	err := handler(srv, validateStream)
	validatePluginStreamPool.Put(validateStream)
	return err
}

func NewParamsValidator() ParamsValidator {
	return &ParamsValidatorImpl{}
}
