// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: helloword.proto

// It translates gRPC into RESTful JSON APIs.
// package v1
package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/example/api/v1"
	"github.com/wetrycode/begonia/runtime/endpoint"
	"google.golang.org/grpc"
)

type endpointRegisters struct {
	serviceRegisters map[string]func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
}

func NewEndpointRegisters() endpoint.EndpointRegister {
	services := map[string]func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error){

		"Greeter": v1.RegisterGreeterHandlerFromEndpoint,
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