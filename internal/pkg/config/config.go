package config

import (
	"fmt"
	"path/filepath"
	"strings"

	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	"github.com/spark-lence/tiga"
)

const (
	APIPkg string = "begonia.org.begonia.common.api.v1"
)

type Config struct {
	*tiga.Configuration
}

type RPCPlugins struct {
	Name     string `mapstructure:"name"`
	Endpoint string `mapstructure:"endpoint"`
	Priority int    `mapstructure:"priority"`
	Timeout  int    `mapstructure:"timeout"`
}

func NewConfig(config *tiga.Configuration) *Config {
	return &Config{Configuration: config}
}

func (c *Config) GetDefaultAdminPasswd() string {
	return c.getWithEnv("admin.password")
}
func (c *Config) GetDefaultAdminName() string {
	return c.getWithEnv("admin.name")
}
func (c *Config) GetDefaultAdminPhone() string {
	return c.getWithEnv("admin.phone")
}
func (c *Config) GetDefaultAdminEmail() string {
	return c.getWithEnv("admin.email")
}

func (c *Config) GetAesKey() string {
	return c.getWithEnv("auth.aes_key")
}
func (c *Config) GetAesIv() string {
	return c.getWithEnv("auth.aes_iv")
}

// func (c *Config) GetCachePrefixKey() string {
// 	return c.getWithEnv("common.cache_key_prefix")
// }

// jwt_secret
func (c *Config) GetJWTLockKey(uid string) string {
	prefix := c.GetCachePrefixKey()
	return fmt.Sprintf("%s:jwt_lock:%s", prefix, uid)
}
func (c *Config) GetJWTSecret() string {
	return c.getWithEnv("auth.jwt_secret")
}
func (c *Config) GetCachePrefixKey() string {
	return fmt.Sprintf("%s:%s", c.GetEnv(), c.getWithEnv("common.cache_prefix_key"))

}
func (c *Config) GetWhiteToken(uid string) string {
	prefix := c.GetCachePrefixKey()
	return fmt.Sprintf("%s:white:token:%s", prefix, uid)
}
func (c *Config) GetUserBlackListKey(uid string) string {
	prefix := c.GetUserBlackListPrefix()
	return fmt.Sprintf("%s:%s", prefix, uid)
}
func (c *Config) GetUserBlackListLockKey() string {
	prefix := c.GetUserBlackListPrefix()
	return fmt.Sprintf("%s:lock", prefix)
}
func (c *Config) GetAppsLockKey() string {
	prefix := c.GetAPPAccessKeyPrefix()
	return fmt.Sprintf("%s:lock", prefix)
}
func (c *Config) GetUserBlackListExpiration() int {
	return c.getIntWithEnv("auth.blacklist.user.cache_expire")
}

func (c *Config) GetAPPAccessKeyExpiration() int {
	return c.getIntWithEnv("auth.app.cache_expire")
}
func (c *Config) GetUserTokenBlackListBloomKey(uid string) string {
	prefix := c.GetUserTokenBlackListBloom()
	return fmt.Sprintf("%s:%s", prefix, uid)
}
func (c *Config) GetUserBlackListPrefix() string {
	prefix := c.GetCachePrefixKey()
	return fmt.Sprintf("%s:user:black", prefix)
}
func (c *Config) GetUserTokenBlackListBloom() string {
	prefix := c.GetFilterPrefix()
	return fmt.Sprintf("%s:user:black", prefix)
}
func (c *Config) GetBlacklistPubSubChannel() string {
	return c.getWithEnv("auth.blacklist.pubsub.channel")
}
func (c *Config) GetKeyValuePubsubKey() string {
	prefix := fmt.Sprintf("%s:%s", c.GetCachePrefixKey(), c.getWithEnv("common.pubsub_key_prefix"))
	return fmt.Sprintf("%s:kv", prefix)
}
func (c *Config) GetFilterPubsubKey() string {
	prefix := fmt.Sprintf("%s:%s", c.GetCachePrefixKey(), c.getWithEnv("common.pubsub_key_channel"))
	return fmt.Sprintf("%s:filter", prefix)
}
func (c *Config) GetMultiCacheReadStrategy() int {
	return c.getIntWithEnv("common.multi_cache_strategy")
}
func (c *Config) GetKeyValuePrefix() string {
	return fmt.Sprintf("%s:%s", c.GetCachePrefixKey(), c.getWithEnv("common.kv_prefix"))
}
func (c *Config) GetFilterPrefix() string {
	prefix := c.GetCachePrefixKey()
	return fmt.Sprintf("%s:%s", prefix, c.getWithEnv("common.filter_key_prefix"))
}
func (c *Config) GetBlacklistPubSubGroup() string {
	return c.getWithEnv("auth.blacklist.pubsub.group")
}
func (c *Config) GetBlacklistFilterEntries() int {
	return c.getIntWithEnv("auth.blacklist.filter.entries")
}
func (c *Config) GetBlacklistFilterErrorRate() int {
	return c.getIntWithEnv("auth.blacklist.filter.error_rate")
}
func (c *Config) GetBlacklistBloomErrRate() float64 {
	return c.GetFloat64(fmt.Sprintf("%s.auth.blacklist.bloom.error_rate", c.GetEnv()))
}

func (c *Config) GetBlacklistBloomM() int {
	return c.getIntWithEnv("auth.blacklist.bloom.m")
}

func (c *Config) GetAPPAccessKey(access string) string {
	prefix := c.GetAPPAccessKeyPrefix()
	return fmt.Sprintf("%s:%s", prefix, access)
}
func (c *Config) GetAppidKey(accessKey string) string {
	prefix := c.GetAppidPrefix()
	return fmt.Sprintf("%s:%s", prefix, accessKey)
}
func (c *Config) GetAppPrefix() string {
	prefix := c.GetCachePrefixKey()
	return fmt.Sprintf("%s:%s", prefix, c.getWithEnv("common.app_key_prefix"))
}
func (c *Config) GetAPPAccessKeyPrefix() string {
	prefix := c.GetAppPrefix()
	return fmt.Sprintf("%s:access_key", prefix)
}
func (c *Config) GetAppidPrefix() string {
	prefix := c.GetAppPrefix()
	return fmt.Sprintf("%s:appid", prefix)
}
func (c *Config) GetAesConfig() (key string, iv string) {
	key = c.getWithEnv("auth.aes_key")
	iv = c.getWithEnv("auth.aes_iv")
	return key, iv
}

func (c *Config) GetCorsConfig() []string {
	return c.GetStringSlice(fmt.Sprintf("%s.gateway.cors", c.GetEnv()))
}
func (c *Config) GetJWTExpiration() int {
	return c.getIntWithEnv("auth.jwt_expiration")
}
func (c *Config) GetPluginDir() string {
	return c.getWithEnv("endpoints.plugins.dir")
}
func (c *Config) GetUploadDir() string {
	return c.getWithEnv("file.upload.dir")
}
func (c *Config) GetProtosDir() string {
	return c.getWithEnv("file.protos.dir")
}
func (c *Config) GetLocalAPIDesc() string {
	return c.getWithEnv("file.protos.desc")
}

func (c *Config) GetPlugins() map[string]interface{} {
	return c.GetStringMap(fmt.Sprintf("%s.gateway.plugins.local", c.GetEnv()))
}
func (c *Config) GetRPCPlugins() ([]*goloadbalancer.Server, error) {
	plugins := make([]*goloadbalancer.Server, 0)

	err := c.UnmarshalKey(fmt.Sprintf("%s.gateway.plugins.rpc", c.GetEnv()), &plugins)
	if err != nil {
		return nil, err
	}
	if len(plugins) == 0 {
		err = c.UnmarshalKey("gateway.plugins.rpc", &plugins)
		if err != nil {
			return nil, err
		}

	}
	return plugins, nil
}
func (c *Config) GetEndpointsPrefix() string {
	return fmt.Sprintf("/%s%s", c.GetEnv(), c.getWithEnv("common.etcd.endpoint.prefix"))
}

func (c *Config) GetGatewayDescriptionOut() string {
	return c.getWithEnv("gateway.descriptor.out_dir")
}

func (c *Config) GetAdminAPIKey() string {
	return c.getWithEnv("auth.admin.apikey")
}

func (c *Config) GetServicePrefix() string {
	prefix := c.GetEndpointsPrefix()
	return fmt.Sprintf("%s/service", prefix)
}
func (c *Config) GetServiceNamePrefix() string {
	prefix := c.GetEndpointsPrefix()
	return fmt.Sprintf("%s/service_name", prefix)
}
func (c *Config) GetServiceTagsPrefix() string {
	prefix := c.GetEndpointsPrefix()
	return fmt.Sprintf("%s/tags", prefix)
}
func (c *Config) GetServiceKey(key string) string {
	if tiga.IsSnowflakeID(key){
		prefix := c.GetServicePrefix()
		return filepath.Join(prefix, key)
	}
	return c.GetServiceNameKey(key)

}
func (c *Config) GetServiceNameKey(name string) string {
	prefix := c.GetServiceNamePrefix()
	return filepath.Join(prefix, name)
}
func (c *Config) GetAppKeyPrefix() string {
	prefix := c.getWithEnv("common.etcd.app.prefix")
	return fmt.Sprintf("%s%s", c.GetEnv(), prefix)
}
func (c *Config) GetAPPKey(id string) string {
	prefix := c.GetAppKeyPrefix()
	return filepath.Join(prefix, id)
}
func (c *Config) GetAPPTagsPrefix() string {
	prefix := c.GetAppKeyPrefix()
	return fmt.Sprintf("%s/tags", prefix)
}
func (c *Config) GetAppTagsKey(tag, key string) string {
	prefix := c.GetAPPTagsPrefix()
	return filepath.Join(prefix, tag, key)
}
func (c *Config) GetTagsKey(tag, id string) string {

	prefix := c.GetServiceTagsPrefix()
	return filepath.Join(prefix, tag, id)
}
func (c *Config) getWithEnv(key string) string {
	originKey := strings.TrimPrefix(key, "common.")

	envKey := fmt.Sprintf("%s.%s", c.GetEnv(), originKey)
	if val := c.GetString(envKey); val != "" {
		return val
	}
	return c.GetString(key)
}
func (c *Config) getIntWithEnv(key string) int {
	envKey := fmt.Sprintf("%s.%s", c.GetEnv(), key)
	if val := c.GetInt(envKey); val != 0 {
		return val
	}
	return c.GetInt(key)
}
func (c *Config) GetRSAPriKey() string {

	return c.getWithEnv("auth.rsa.private_key")

}
func (c *Config) GetRSAPubKey() string {

	return c.getWithEnv("auth.rsa.public_key")
}
