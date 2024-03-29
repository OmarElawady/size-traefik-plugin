package stubs

import (
	"context"
	zbus "github.com/threefoldtech/zbus"
)

type GatewayStub struct {
	client zbus.Client
	module string
	object zbus.ObjectID
}

func NewGatewayStub(client zbus.Client) *GatewayStub {
	return &GatewayStub{
		client: client,
		module: "gateway",
		object: zbus.ObjectID{
			Name:    "manager",
			Version: "0.0.1",
		},
	}
}

func (s *GatewayStub) DeleteNamedProxy(ctx context.Context, arg0 string) (ret0 error) {
	args := []interface{}{arg0}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "DeleteNamedProxy", args...)
	if err != nil {
		panic(err)
	}
	ret0 = new(zbus.RemoteError)
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	return
}

func (s *GatewayStub) ReportConsumption(ctx context.Context, arg0 string, arg1 uint64, arg2 uint64) {
	args := []interface{}{arg0, arg1, arg2}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "ReportConsumption", args...)
	if err != nil {
		panic(err)
	}
	return
}

func (s *GatewayStub) SetNamedProxy(ctx context.Context, arg0 string, arg1 string, arg2 []string) (ret0 string, ret1 error) {
	args := []interface{}{arg0, arg1, arg2}
	result, err := s.client.RequestContext(ctx, s.module, s.object, "SetNamedProxy", args...)
	if err != nil {
		panic(err)
	}
	if err := result.Unmarshal(0, &ret0); err != nil {
		panic(err)
	}
	ret1 = new(zbus.RemoteError)
	if err := result.Unmarshal(1, &ret1); err != nil {
		panic(err)
	}
	return
}
