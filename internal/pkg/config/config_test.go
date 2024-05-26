package config_test

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	conf "github.com/begonia-org/begonia/config"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	c.Convey("TestConfig", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		// log.Printf("env: %s", env)
		cnf := conf.ReadConfig(env)
		config := cfg.NewConfig(cnf)
		prefix := config.GetCacheKeyPrefix()
		c.So(config.GetDefaultAdminPasswd(), c.ShouldNotEqual, "")
		c.So(config.GetDefaultAdminName(), c.ShouldNotEqual, "")
		c.So(config.GetDefaultAdminPhone(), c.ShouldNotEqual, "")
		c.So(config.GetDefaultAdminEmail(), c.ShouldNotEqual, "")
		c.So(config.GetAesKey(), c.ShouldNotEqual, "")
		c.So(config.GetAesIv(), c.ShouldNotEqual, "")
		c.So(config.GetJWTLockKey("test"), c.ShouldEndWith, ":jwt_lock:test")
		c.So(config.GetJWTSecret(), c.ShouldNotBeEmpty)
		c.So(config.GetWhiteToken("test"), c.ShouldEndWith, ":white:token:test")
		c.So(config.GetUserBlackListKey("test"), c.ShouldEqual, fmt.Sprintf("%s:%s", prefix, "user:black:test"))
		c.So(config.GetUserBlackListLockKey(), c.ShouldEndWith, ":user:black:lock")
		c.So(config.GetAppsLockKey(), c.ShouldEndWith, ":access_key:lock")
		c.So(config.GetUserBlackListExpiration(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetAPPAccessKeyExpiration(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetUserTokenBlackListBloomKey("test"), c.ShouldEndWith, ":user:black:test")
		c.So(config.GetUserBlackListPrefix(), c.ShouldEqual, fmt.Sprintf("%s:%s", prefix, "user:black"))
		c.So(config.GetUserTokenBlackListBloom(), c.ShouldEndWith, ":user:black")
		c.So(config.GetBlacklistPubSubChannel(), c.ShouldNotBeEmpty)
		c.So(config.GetKeyValuePubsubKey(), c.ShouldEndWith, ":kv")
		c.So(config.GetFilterPubsubKey(), c.ShouldEndWith, ":filter")
		c.So(config.GetMultiCacheReadStrategy(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetKeyValuePrefix(), c.ShouldNotBeEmpty)
		c.So(config.GetFilterPrefix(), c.ShouldNotBeEmpty)
		c.So(config.GetBlacklistPubSubGroup(), c.ShouldNotBeEmpty)
		c.So(config.GetBlacklistFilterEntries(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetBlacklistFilterErrorRate(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetBlacklistBloomErrRate(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetBlacklistBloomM(), c.ShouldBeGreaterThanOrEqualTo, 0)
		c.So(config.GetAPPAccessKey("test"), c.ShouldStartWith, config.GetAPPAccessKeyPrefix())
		c.So(config.GetAPPAccessKeyPrefix(), c.ShouldNotBeEmpty)
		c.So(config.GetAppidKey("test"), c.ShouldStartWith, config.GetAppidPrefix())
		kv, iv := config.GetAesConfig()
		c.So(kv, c.ShouldNotBeEmpty)
		c.So(iv, c.ShouldNotBeEmpty)
		c.So(len(config.GetCorsConfig()), c.ShouldBeGreaterThan, 0)
		c.So(config.GetJWTExpiration(), c.ShouldBeGreaterThan, 0)
		c.So(config.GetPluginDir(), c.ShouldNotBeEmpty)
		c.So(config.GetUploadDir(), c.ShouldNotBeEmpty)
		c.So(config.GetProtosDir(), c.ShouldNotBeEmpty)
		c.So(config.GetLocalAPIDesc(), c.ShouldNotBeEmpty)
		c.So(config.GetPlugins(), c.ShouldNotBeEmpty)
		ss, err := config.GetRPCPlugins()
		c.So(err, c.ShouldBeNil)
		c.So(ss, c.ShouldNotBeEmpty)
		c.So(config.GetEndpointsPrefix(), c.ShouldNotBeEmpty)
		c.So(config.GetGatewayDescriptionOut(), c.ShouldNotBeEmpty)
		c.So(config.GetAdminAPIKey(), c.ShouldNotBeEmpty)
		c.So(config.GetServicePrefix(), c.ShouldEndWith, "/service")
		c.So(config.GetServiceNamePrefix(), c.ShouldEndWith, "/service_name")
		c.So(config.GetServiceTagsPrefix(), c.ShouldEndWith, "/tags")
		c.So(config.GetServiceKey("test"), c.ShouldStartWith, config.GetServicePrefix())
		c.So(config.GetServiceNameKey("test"), c.ShouldStartWith, config.GetServiceNamePrefix())
		c.So(config.GetAppKeyPrefix(), c.ShouldNotBeEmpty)
		c.So(config.GetAPPKey("test"), c.ShouldStartWith, config.GetAppKeyPrefix())
		c.So(config.GetAPPTagsPrefix(), c.ShouldEndWith, "/tags")
		c.So(config.GetAppTagsKey("test", "test"), c.ShouldStartWith, config.GetAPPTagsPrefix())
		c.So(config.GetAppTagsKey("test", "test"), c.ShouldEndWith, "/test")
		c.So(config.GetTagsKey("test", "test"), c.ShouldStartWith, config.GetServiceTagsPrefix())

		c.So(config.GetRSAPriKey(), c.ShouldNotBeEmpty)
		c.So(config.GetRSAPubKey(), c.ShouldNotBeEmpty)
		c.So(config.GetAppPrefix(), c.ShouldNotBeEmpty)
		patch:=gomonkey.ApplyFuncReturn((*viper.Viper).UnmarshalKey,fmt.Errorf("error"))
		defer patch.Reset()
		ss, err = config.GetRPCPlugins()
		c.So(err, c.ShouldNotBeNil)
		c.So(ss, c.ShouldBeNil)
		cnf = conf.ReadConfig("dev")
		config2:=cfg.NewConfig(cnf)
		c.So(config2.GetEndpointsPrefix(), c.ShouldNotBeEmpty)
		c.So(config2.GetAppKeyPrefix(), c.ShouldNotBeEmpty)

	})
}
