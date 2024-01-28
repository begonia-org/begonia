// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: file.proto

// It translates gRPC into RESTful JSON APIs.
// package v1
package main

import (
	"context"
	"github.com/begonia-org/begonia/common/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/endpoint"
	"google.golang.org/grpc"
)

type endpointRegisters struct {
	serviceRegisters map[string]endpoint.EndpointRegisterFunc
}

func NewEndpointRegisters() endpoint.EndpointRegister {
	services := map[string]endpoint.EndpointRegisterFunc{

		"FileService": v1.RegisterFileServiceHandlerFromEndpoint,
	}
	return &endpointRegisters{
		serviceRegisters: services,
	}
}

func (e *endpointRegisters) RegisterService(serviceName string, ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	if registerFunc, ok := e.serviceRegisters[serviceName]; ok {
		return registerFunc(ctx, mux, endpoint, opts)
	}
	return nil
}

func (e *endpointRegisters) GetAllEndpoints() map[string]endpoint.EndpointRegisterFunc {
	return e.serviceRegisters
}
func (e *endpointRegisters) RegisterAll(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	for _, registerFunc := range e.serviceRegisters {
		if err = registerFunc(ctx, mux, endpoint, opts); err != nil {
			return err
		}
	}
	return nil
}
