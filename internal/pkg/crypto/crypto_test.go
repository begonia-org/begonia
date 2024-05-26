package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/utils"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

func TestGenerateRandPasswd(t *testing.T) {
	c.Convey("TestGenerateRandPasswd", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		auth := NewUsersAuth(cfg.NewConfig(config))
		passwd, err := auth.GenerateRandPasswd()
		c.So(err, c.ShouldBeNil)
		c.So(len(passwd), c.ShouldEqual, 16)
		patch := gomonkey.ApplyFuncReturn(rand.Int, nil, fmt.Errorf("rand.Int: nil"))
		defer patch.Reset()

		passwd, err = auth.GenerateRandPasswd()
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(passwd, c.ShouldEqual, "")
		c.So(err.Error(), c.ShouldContainSubstring, "generate rand password fail")

	})
}

func TestNewUsersAuth(t *testing.T) {
	c.Convey("TestNewUsersAuth", t, func() {
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  os.ReadFile,
				output: []interface{}{[]byte{}, fmt.Errorf("file not found")},
				err:    fmt.Errorf("file not found"),
			},
			{
				patch:  pem.Decode,
				output: []interface{}{nil, nil},
				err:    fmt.Errorf("invalid private key data"),
			},
			{
				patch:  x509.ParsePKCS8PrivateKey,
				output: []interface{}{nil, fmt.Errorf("x509.ParsePKCS8PrivateKey: nil")},
				err:    fmt.Errorf("invalid private key data: x509.ParsePKCS8PrivateKey: nil"),
			},
			{
				patch:  x509.ParsePKCS8PrivateKey,
				output: []interface{}{&ecdsa.PrivateKey{}, nil},
				err:    fmt.Errorf("not an RSA private key"),
			},
		}
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		// patch := gomonkey.ApplyFuncReturn((*UsersAuth).LoadPrivateKey())
		for _, v := range cases {
			patch := gomonkey.ApplyFuncReturn(v.patch, v.output...)
			defer patch.Reset()
			c.So(func() { NewUsersAuth(cnf) }, c.ShouldPanicWith, v.err.Error())
			patch.Reset()
		}

	})
}

func TestGenerateRandSeed(t *testing.T) {
	c.Convey("TestGenerateRandSeed", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		auth := NewUsersAuth(cnf)
		c.So(auth.GenerateRandSeed(), c.ShouldBeGreaterThan, 0)
	})
}

func TestGenerateKeys(t *testing.T) {
	c.Convey("TestGenerateKeys", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		auth := NewUsersAuth(cnf)
		priv, pub, err := auth.GenerateKeys(1024)
		c.So(err, c.ShouldBeNil)
		c.So(priv, c.ShouldNotBeNil)
		c.So(pub, c.ShouldNotBeNil)

		priv, pub, err = auth.GenerateKeys(0)
		c.So(err, c.ShouldNotBeNil)
		c.So(priv, c.ShouldBeNil)
		c.So(pub, c.ShouldBeNil)
	})
}

func TestEncryptWithPublicKey(t *testing.T) {
	c.Convey("TestEncrypt With PublicKey", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		auth := NewUsersAuth(cnf)
		pri, pub, _ := auth.GenerateKeys(1024)
		plain := "test"
		cipher, err := auth.EncryptWithPublicKey(plain, pub)
		c.So(err, c.ShouldBeNil)
		c.So(cipher, c.ShouldNotBeNil)
		msg, err := auth.DecryptWithPrivateKey(cipher, pri)
		c.So(err, c.ShouldBeNil)
		c.So(msg, c.ShouldEqual, plain)

		pri2, _, _ := auth.GenerateKeys(1024)
		patch := gomonkey.ApplyFuncReturn(rsa.EncryptOAEP, nil, fmt.Errorf("rsa.DecryptOAEP: nil"))
		defer patch.Reset()
		cipher2, err2 := auth.EncryptWithPublicKey(plain, pub)
		patch.Reset()
		c.So(err2, c.ShouldNotBeNil)
		c.So(err2.Error(), c.ShouldContainSubstring, "rsa.DecryptOAEP: nil")
		c.So(cipher2, c.ShouldBeNil)

		msg, err = auth.DecryptWithPrivateKey(cipher, pri2)
		c.So(err, c.ShouldNotBeNil)
		c.So(msg, c.ShouldEqual, "")
	})
}

func TestEncryptAES(t *testing.T) {
	c.Convey("TestEncryptAES", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		auth := NewUsersAuth(cnf)
		key := []byte("1234567890123456")
		// iv := []byte("1234567890123456")
		plain := "test"
		cipher, err := auth.EncryptAES(key, plain)
		c.So(err, c.ShouldBeNil)
		c.So(cipher, c.ShouldNotBeNil)

		cipher2, err2 := auth.EncryptAES([]byte("123456"), plain)
		c.So(err2, c.ShouldNotBeNil)
		c.So(cipher2, c.ShouldEqual, "")

		patch := gomonkey.ApplyFuncReturn(io.ReadFull, 0, fmt.Errorf("io.ReadFull: nil"))
		defer patch.Reset()
		cipher3, err3 := auth.EncryptAES(key, plain)
		c.So(err3, c.ShouldNotBeNil)
		c.So(cipher3, c.ShouldEqual, "")

	})
}

func TestGenerateAuthSeed(t *testing.T) {
	c.Convey("TestGenerateAuthSeed", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		auth := NewUsersAuth(cnf)
		seed, err := auth.GenerateAuthSeed("1234567890123456")
		c.So(err, c.ShouldBeNil)
		c.So(seed, c.ShouldNotBeEmpty)

		patch := gomonkey.ApplyFuncReturn(UsersAuth.GenerateRandPasswd, "", fmt.Errorf("rand.Read: nil"))
		defer patch.Reset()

		seed2, err2 := auth.GenerateAuthSeed("1234567890123456")
		c.So(err2, c.ShouldNotBeNil)
		c.So(seed2, c.ShouldBeEmpty)
		c.So(err2.Error(), c.ShouldContainSubstring, "rand.Read: nil")
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn(tiga.StructToJsonStr, "", fmt.Errorf("StructToJsonStr: error"))
		defer patch2.Reset()
		seed3, err3 := auth.GenerateAuthSeed("1234567890123456")
		c.So(err3, c.ShouldNotBeNil)
		c.So(seed3, c.ShouldBeEmpty)
		c.So(err3.Error(), c.ShouldContainSubstring, "StructToJsonStr: error")

	})
}
func TestRSADecrypt(t *testing.T) {
	c.Convey("TestRSADecrypt", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		auth := NewUsersAuth(cnf)
		plain := "test"
		pubKey, err := utils.LoadPublicKeyFromFile(cnf.GetRSAPubKey())
		c.So(err, c.ShouldBeNil)
		c.So(pubKey, c.ShouldNotBeNil)
		enc, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(plain))
		c.So(err, c.ShouldBeNil)
		encodedData := base64.StdEncoding.EncodeToString(enc)
		data, err := auth.RSADecrypt(encodedData)
		c.So(err, c.ShouldBeNil)
		c.So(string(data), c.ShouldEqual, plain)

		data, err = auth.RSADecrypt("encodedData")
		c.So(err, c.ShouldNotBeNil)
		c.So(data, c.ShouldBeNil)

	})
}
