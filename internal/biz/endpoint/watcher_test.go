package endpoint_test

import (
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz/endpoint"
	"github.com/begonia-org/begonia/internal/data"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
)

func newWatcher() *endpoint.EndpointWatcher {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := config.ReadConfig(env)
	cnf := cfg.NewConfig(conf)
	repo := data.NewEndpointRepo(conf, gateway.Log)
	return endpoint.NewWatcher(cnf, repo)
}
