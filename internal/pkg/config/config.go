package config

import (
	"fmt"

	"github.com/spark-lence/tiga"
)

const (
	APIPkg string = "begonia.org.begonia.common.api.v1"
)

type Config struct {
	*tiga.Configuration
}

func NewConfig(config *tiga.Configuration) *Config {
	return &Config{Configuration: config}
}

func (c *Config) GetDefaultAdminPasswd() string {
	return c.GetString("admin.password")
}
func (c *Config) GetDefaultAdminName() string {
	return c.GetString("admin.name")
}
func (c *Config) GetDefaultAdminPhone() string {
	return c.GetString("admin.phone")
}
func (c *Config) GetDefaultAdminEmail() string {
	return c.GetString("admin.email")
}

func (c *Config) GetAesKey() string {
	return c.GetString("auth.aes_key")
}
func (c *Config) GetAesIv() string {
	return c.GetString("auth.aes_iv")
}

// jwt_secret
func (c *Config) GetJWTLockKey(uid string) string {
	prefix := c.GetString("common.rdb_key_prefix")
	return fmt.Sprintf("%s:jwt_lock:%s", prefix, uid)
}
func (c *Config) GetJWTSecret() string {
	return c.GetString("auth.jwt_secret")
}
func (c *Config) GetWhiteToken(uid string) string {
	prefix := c.GetString("common.rdb_key_prefix")
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
	return c.GetInt("auth.user.blacklist.cache_expire")
}

func (c *Config) GetAPPAccessKeyExpiration() int {
	return c.GetInt("auth.app.cache_expire")
}
func (c *Config) GetUserTokenBlackListBloomKey(uid string) string {
	prefix := c.GetUserTokenBlackListBloom()
	return fmt.Sprintf("%s:%s", prefix, uid)
}
func (c *Config) GetUserBlackListPrefix() string {
	prefix := c.GetString("common.rdb_key_prefix")
	return fmt.Sprintf("%s:user:black:", prefix)
}
func (c *Config) GetUserTokenBlackListBloom() string {
	prefix := c.GetFilterPrefix()
	return fmt.Sprintf("%s:user:black", prefix)
}
func (c *Config) GetBlacklistPubSubChannel() string {
	return c.GetString("auth.blacklist.pubsub.channel")
}
func (c *Config) GetKeyValuePubsubKey() string {
	prefix := c.GetString("common.pubsub_key_prefix")
	return fmt.Sprintf("%s:kv", prefix)
}
func (c *Config) GetFilterPubsubKey() string {
	prefix := c.GetString("common.pubsub_key_channel")
	return fmt.Sprintf("%s:filter", prefix)
}
func (c *Config) GetMultiCacheReadStrategy() int {
	return c.GetInt("common.multi_cache_strategy")
}
func (c *Config) GetKeyValuePrefix() string {
	return c.GetString("common.kv_prefix")
}
func (c *Config) GetFilterPrefix() string {
	return c.GetString("common.filter_key_prefix")
}
func (c *Config) GetBlacklistPubSubGroup() string {
	return c.GetString("auth.blacklist.pubsub.group")
}
func (c *Config) GetBlacklistFilterEntries() int {
	return c.GetInt("auth.blacklist.filter.entries")
}
func (c *Config) GetBlacklistFilterErrorRate() int {
	return c.GetInt("auth.blacklist.filter.error_rate")
}
func (c *Config) GetBlacklistBloomErrRate() float64 {
	return c.GetFloat64(fmt.Sprintf("%s.auth.blacklist.bloom.error_rate", c.GetEnv()))
}

func (c *Config) GetBlacklistBloomM() int {
	return c.GetInt("auth.blacklist.bloom.m")
}

//	func (c *Config) GetBlacklistBloomErrRate() float64 {
//		return c.GetFloat64(fmt.Sprintf("%s.auth.blacklist.bloom.error_rate", c.GetEnv()))
//	}
func (c *Config) GetAPPAccessKey(access string) string {
	prefix := c.GetAPPAccessKeyPrefix()
	return fmt.Sprintf("%s:%s", prefix, access)
}
func (c *Config) GetAPPAccessKeyPrefix() string {
	prefix := c.GetString("common.app_key_prefix")
	return fmt.Sprintf("%s:access_key", prefix)
}

func (c *Config) GetAesConfig() (key string, iv string) {
	key = c.GetString("auth.aes_key")
	iv = c.GetString("auth.aes_iv")
	return key, iv
}

func (c *Config) GetCorsConfig() []string {
	return c.GetStringSlice(fmt.Sprintf("%s.gateway.cors", c.GetEnv()))
}
func (c *Config) GetJWTExpiration() int {
	return c.GetInt("auth.jwt_expiration")
}
func (c *Config) GetPluginDir() string {
	return c.GetString("endpoints.plugins.dir")
}
func (c *Config) GetUploadDir() string {
	return c.GetString("file.upload.dir")
}
func (c *Config) GetProtosDir() string {
	return c.GetString("file.protos.dir")
}

// func (c *Config) GetAppKeyPrefix() string {
// 	prefix := c.GetString("common.app_key_prefix")
// 	return prefix
// }

// func GetBlacklistPubSubChannel(c *Config) golayeredbloom.ChannelName {
// 	return golayeredbloom.ChannelName(c.GetBlacklistPubSubChannel())
// }

// func GetBlacklistPubSubGroup(c *Config) golayeredbloom.GroupName {
// 	return golayeredbloom.GroupName(c.GetBlacklistPubSubGroup())
// }
