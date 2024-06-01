package utils_test

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/utils"
	c "github.com/smartystreets/goconvey/convey"
)

func TestGenerateRandomString(t *testing.T) {
	c.Convey("TestGenerateRandomString", t, func() {
		s, err := utils.GenerateRandomString(10)
		c.So(err, c.ShouldBeNil)
		c.So(len(s), c.ShouldEqual, 10)

		patch := gomonkey.ApplyFuncReturn(rand.Read, 0, fmt.Errorf("rand.Read: nil"))
		defer patch.Reset()
		s, err = utils.GenerateRandomString(10)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(s, c.ShouldEqual, "")

	})
}

func TestLoadPublicKeyFromFile(t *testing.T) {
	c.Convey("TestLoadPublicKeyFromFile", t, func() {
		pub, err := utils.LoadPublicKeyFromFile("test")
		c.So(err, c.ShouldNotBeNil)
		c.So(pub, c.ShouldBeNil)

		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		pub, err = utils.LoadPublicKeyFromFile(cnf.GetRSAPubKey())
		c.So(err, c.ShouldBeNil)
		c.So(pub, c.ShouldNotBeNil)
		patch := gomonkey.ApplyFuncReturn(pem.Decode, nil, nil)
		defer patch.Reset()
		pub, err = utils.LoadPublicKeyFromFile(cnf.GetRSAPubKey())
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(pub, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "failed to decode PEM block")

		patch1 := gomonkey.ApplyFuncReturn(x509.ParsePKIXPublicKey, nil, fmt.Errorf("failed to parse PKIX public key"))
		defer patch1.Reset()
		pub, err = utils.LoadPublicKeyFromFile(cnf.GetRSAPubKey())
		patch1.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(pub, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "failed to parse PKIX public key")
	})
}
