package grpc

import "testing"

func TestEndpoint(t *testing.T) {
	e := NewEndpointImpl("api.v1")
	e.RegisterService("ManagerService")
}
