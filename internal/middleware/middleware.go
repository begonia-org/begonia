package middleware

import (
	"context"
	"sort"
	"time"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/middleware/auth"
	"github.com/begonia-org/begonia/internal/pkg/config"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	gosdk "github.com/begonia-org/go-sdk"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/google/wire"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
)

//	var Plugins = map[string]gosdk.GrpcPlugin{
//		"jwt": &auth.JWTAuth{},
//	}
var ProviderSet = wire.NewSet(New,auth.NewAccessKeyAuth)

func New(config *config.Config,
	rdb *tiga.RedisDao,
	user *biz.AuthzUsecase,
	log logger.Logger,
	authz *biz.AccessKeyAuth,
	local *data.LayeredCache) *PluginsApply {
	jwt := auth.NewJWTAuth(config, rdb, user, log)
	ak := auth.NewAccessKeyAuth(authz, config, log)
	apiKey := auth.NewApiKeyAuth(config)
	plugins := map[string]gosdk.LocalPlugin{
		"onlyJWT":           jwt,
		"onlyAK":            ak,
		"logger":            gateway.NewLoggerMiddleware(log),
		"exception":         gateway.NewException(log),
		"http":              NewHttp(),
		"auth":              auth.NewAuth(ak, jwt, apiKey),
		"params_validator":  NewParamsValidator(),
		"only_api_key_auth": apiKey,
		// "logger":NewLoggerMiddleware(log),
	}
	pluginsApply := NewPluginsApply()
	pluginsNeed := config.GetPlugins()
	for pluginName, priority := range pluginsNeed {
		log.Infof(context.TODO(), "plugin %s priority %d", pluginName, priority)
		if plugin, ok := plugins[pluginName]; ok {
			pluginsApply.Register(plugin, priority.(int))
		} else {
			log.Warnf(context.TODO(), "plugin %s not found", pluginName)

		}
	}

	rpcPlugins, err := config.GetRPCPlugins()
	if err != nil {
		log.Errorf(context.TODO(), "get rpc plugins error:%v", err)
		return pluginsApply
	}
	for _, rpc := range rpcPlugins {
		lb := goloadbalancer.NewGrpcLoadBalance(rpc)
		pluginsApply.Register(&pluginImpl{
			lb:      lb,
			name:    rpc.Name,
			timeout: time.Duration(rpc.Timeout) * time.Second,
		}, rpc.Priority)
	}
	return pluginsApply
}

type PluginsApply struct {
	Plugins gosdk.Plugins
}

func NewPluginsApply() *PluginsApply {
	apply := &PluginsApply{
		Plugins: make(gosdk.Plugins, 0),
	}

	return apply
}
func (p *PluginsApply) Register(plugin gosdk.LocalPlugin, priority int) {
	plugin.SetPriority(priority)
	p.Plugins = append(p.Plugins, plugin)
	sort.Sort(p.Plugins)
}
func (p *PluginsApply) UnaryInterceptorChains() []grpc.UnaryServerInterceptor {
	chains := make([]grpc.UnaryServerInterceptor, 0)
	for _, plugin := range p.Plugins {
		chains = append(chains, plugin.(gosdk.LocalPlugin).UnaryInterceptor)
	}
	return chains
}

func (p *PluginsApply) StreamInterceptorChains() []grpc.StreamServerInterceptor {
	chains := make([]grpc.StreamServerInterceptor, 0)
	for _, plugin := range p.Plugins {
		chains = append(chains, plugin.(gosdk.LocalPlugin).StreamInterceptor)
	}
	return chains
}
