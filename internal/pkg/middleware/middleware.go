package middleware

import (
	"sort"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/middleware/validator"
	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	gosdk "github.com/begonia-org/go-sdk"
	"github.com/sirupsen/logrus"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
)

// var Plugins = map[string]gosdk.GrpcPlugin{
// 	"jwt": &validator.JWTAuth{},
// }

func New(config *config.Config,
	rdb *tiga.RedisDao,
	user *biz.UsersUsecase,
	log *logrus.Logger,
	app biz.AppRepo,
	local *data.LayeredCache) *PluginsApply {
	jwt := validator.NewJWTAuth(config, rdb, user, log)
	ak := validator.NewAccessKeyAuth(app, config, local, log)
	plugins := map[string]gosdk.LocalPlugin{
		"onlyJWT":   jwt,
		"onlyAK":    ak,
		"logger":    NewLoggerMiddleware(log),
		"exception": NewException(log),
		"http":      NewHttp(),
		"auth":      NewAuth(ak, jwt),
		// "logger":NewLoggerMiddleware(log),
	}
	pluginsApply := NewPluginsApply()
	pluginsNeed := config.GetPlugins()
	for pluginName, priority := range pluginsNeed {
		if plugin, ok := plugins[pluginName]; ok {
			pluginsApply.Register(plugin, priority.(int))
		} else {
			log.Warnf("plugin %s not found", pluginName)

		}
	}

	rpcPlugins, err := config.GetRPCPlugins()
	if err != nil {
		log.Errorf("get rpc plugins error:%v", err)
		return pluginsApply
	}
	for _, rpc := range rpcPlugins {
		lb:=goloadbalancer.NewGrpcLoadBalance(rpc)
		pluginsApply.Register(&pluginImpl{
			lb: lb,
			name:     rpc.Name,
			timeout:  time.Duration(rpc.Timeout) * time.Second,
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
