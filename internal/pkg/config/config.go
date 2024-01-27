package config

import (
	"fmt"

	"github.com/spark-lence/tiga"
)

const (
	APIPkg string = "api.v1"
)

type Config struct {
	*tiga.Configuration
}

func NewConfig(config *tiga.Configuration) *Config {
	return &Config{Configuration: config}
}
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
	prefix := c.GetString("common.rdb_key_prefix")
	return fmt.Sprintf("%s:user:black:%s", prefix, uid)
}
func (c *Config) GetUserBlackListPrefix() string {
	prefix := c.GetString("common.rdb_key_prefix")
	return fmt.Sprintf("%s:user:black:*", prefix)
}
func (c *Config) GetWorkerTokenKey(token string) string {
	prefix := c.GetString("common.rdb_key_prefix")
	return fmt.Sprintf("%s:worker:%s", prefix, token)
}
func (c *Config) GetAesConfig() (key string, iv string) {
	key = c.GetString("auth.aes_key")
	iv = c.GetString("auth.aes_iv")
	return key, iv
}
func (c *Config) GetCorsConfig() []string {
	return c.GetStringSlice(fmt.Sprintf("%s.gateway.cors", c.GetEnv()))
}

func (c *Config) GetPluginDir() string {
	return c.GetString("endpoints.plugins.dir")
}
