package main

import (
	"context"
	"plugin"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/wetrycode/begonia/endpoint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	p, err := plugin.Open("/data/work/wetrycode/begonia/protos/20240127200712/protos/example.so")
	if err != nil {
		panic(err)
	}
	newEndpointSymbol, err := p.Lookup("NewEndpointRegisters")
	if err != nil {
		panic(err)
	}
	newRegisterFunc := newEndpointSymbol.(func() endpoint.EndpointRegister)
	newEndpointRegisters := newRegisterFunc()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = newEndpointRegisters.RegisterService("Greeter", context.Background(), mux, "localhost:50051", opts)
	if err != nil {
		panic(err)
	}

}
