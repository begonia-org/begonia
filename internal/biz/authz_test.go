package biz_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"

	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/utils"
	v1 "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/metadata"

	c "github.com/smartystreets/goconvey/convey"
)

var seedAuthToken = ""
var seedTimestampToken = ""
var authzStr = ""

func newAuthzBiz() *biz.AuthzUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewAuthzRepo(config, gateway.Log)
	user := data.NewUserRepo(config, gateway.Log)
	cnf := cfg.NewConfig(config)
	crypto := crypto.NewUsersAuth(cnf)

	return biz.NewAuthzUsecase(repo, user, gateway.Log, crypto, cnf)
}

func testAuthSeed(t *testing.T) {
	authzBiz := newAuthzBiz()
	c.Convey("test auth seed", t, func() {
		token := time.Now().UnixMilli() * 1000
		seed, err := authzBiz.AuthSeed(context.TODO(), &v1.AuthLogAPIRequest{
			Token: fmt.Sprintf("%d", token),
		})
		seedTimestampToken = fmt.Sprintf("%d", token)
		c.So(err, c.ShouldBeNil)
		seedAuthToken = seed

		patch := gomonkey.ApplyFuncReturn((crypto.UsersAuth).GenerateAuthSeed, "", fmt.Errorf("error auth seed"))
		defer patch.Reset()
		_, err = authzBiz.AuthSeed(context.TODO(), &v1.AuthLogAPIRequest{
			Token: fmt.Sprintf("%d", token),
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "error auth seed")

	})

}

// encryptAES 加密函数
func encryptAES(ciphertext string, secretKey string) (string, error) {
	cipherTextBytes, err := hex.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	iv := cipherTextBytes[:aes.BlockSize] // IV 通常与块大小相等
	encrypted := cipherTextBytes[aes.BlockSize:]

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	mode := cipher.NewCFBDecrypter(block, iv)
	mode.XORKeyStream(encrypted, encrypted)

	return string(encrypted), nil
}
func getUserAuth(account, password string, pubKey *rsa.PublicKey, seedToken, timestampToken string) (*v1.LoginAPIRequest, error) {
	auth := &v1.UserAuth{
		Account:  account,
		Password: password,
	}
	payload, err := json.Marshal(auth)
	if err != nil {
		return nil, err
	}
	enc, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, payload)
	if err != nil {
		return nil, err
	}
	encodedData := base64.StdEncoding.EncodeToString(enc)
	msg, err := encryptAES(seedToken, timestampToken)
	if err != nil {
		return nil, err

	}
	authSeed := &v1.AuthSeed{}
	decodedData, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, fmt.Errorf("decode seed failed: %w", err)
	}
	// log.Printf("decodedData: %s", decodedData)
	err = json.Unmarshal(decodedData, authSeed)
	if err != nil {
		return nil, fmt.Errorf("unmarshal auth seed failed: %w", err)
	}
	loginInfo := &v1.LoginAPIRequest{
		Auth:        encodedData,
		Seed:        authSeed.Seed,
		IsKeepLogin: true,
	}
	return loginInfo, nil
}
func testLogin(t *testing.T) {
	authzBiz := newAuthzBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	adminUser := cnf.GetDefaultAdminName()
	adminPasswd := cnf.GetDefaultAdminPasswd()
	_, filename, _, _ := runtime.Caller(0)
	file := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "cert", "auth_public_key.pem")
	pubKey, err := utils.LoadPublicKeyFromFile(file)
	if err != nil {
		t.Error(err)
	}
	c.Convey("test login success", t, func() {
		info, err := getUserAuth(adminUser, adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
		c.So(err, c.ShouldBeNil)
		token, err := authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldBeNil)
		c.So(token, c.ShouldNotBeNil)
		authzStr = token.Token
	})
	c.Convey("test login failed with invalid password", t, func() {
		info, err := getUserAuth(adminUser, "8990efgc765gghyh", pubKey, seedAuthToken, seedTimestampToken)
		c.So(err, c.ShouldBeNil)
		token, err := authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUserPasswordInvalid.Error())
		c.So(token, c.ShouldBeNil)
	})
	c.Convey("test login failed with invalid account", t, func() {
		info, err := getUserAuth("dadasaddecccc", adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
		c.So(err, c.ShouldBeNil)
		token, err := authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUserNotFound.Error())
		c.So(token, c.ShouldBeNil)
	})
	c.Convey("test login failed with invalid pub key", t, func() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("generate rsa error: %v", err)
		}

		publicKey := privateKey.PublicKey
		info, err := getUserAuth(adminUser, adminPasswd, &publicKey, seedAuthToken, seedTimestampToken)
		c.So(err, c.ShouldBeNil)
		token, err := authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldNotBeNil)
		c.So(token, c.ShouldBeNil)
	})

	c.Convey("test login failed with invalid seed token expiration", t, func() {
		// token := (time.Now().UnixMilli() - int64(time.Duration(120)*time.Second * time.Millisecond)) * 1000
		// token := time.Now().UnixMilli() * 1000 - (120 * time.Millisecond * time.Second)
		token := time.Now().UnixMilli() * 1000
		patch := gomonkey.ApplyFunc((crypto.UsersAuth).GenerateRandSeed, func(_ crypto.UsersAuth) int64 {
			timestamp := time.Now().Unix() - (int64(time.Second) * 120)
			seed := time.Now().UnixNano()

			src := mrand.NewSource(seed)
			r := mrand.New(src)
			// 生成一个 0 到 999 之间的随机数
			randomThreeDigits := r.Intn(10000)

			// 将随机数拼接到时间戳后面
			seed = timestamp*10000 + int64(randomThreeDigits)

			return seed
		})
		defer patch.Reset()
		seed, err := authzBiz.AuthSeed(context.TODO(), &v1.AuthLogAPIRequest{
			Token: fmt.Sprintf("%d", token),
		})
		c.So(err, c.ShouldBeNil)

		timestampToken := fmt.Sprintf("%d", token)

		info, err := getUserAuth(adminUser, adminPasswd, pubKey, seed, timestampToken)
		c.So(err, c.ShouldBeNil)
		_, err = authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenExpired.Error())
	})
	c.Convey("test login failed with invalid UserAuth", t, func() {

		info, _ := getUserAuth(adminUser, adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
		patch := gomonkey.ApplyFuncReturn(json.Marshal, nil, fmt.Errorf("error marshal"))
		defer patch.Reset()
		_, err = authzBiz.Login(context.TODO(), info)

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "error marshal")
		patch.Reset()
	})
	c.Convey("test login failed with invalid EncryptAES", t, func() {
		patch := gomonkey.ApplyFuncReturn(tiga.EncryptAES, "", fmt.Errorf("error encryptAES"))
		defer patch.Reset()
		info, _ := getUserAuth(adminUser, adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
		_, err := authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrEncrypt.Error())
		patch.Reset()

	})

	c.Convey("test login failed with invalid user status", t, func() {
		mockUser := &v1.Users{
			Password: adminPasswd,
			Name:     adminUser,
			Status:   v1.USER_STATUS_LOCKED,
		}
		repo := data.NewUserRepo(config, gateway.Log)
		patch := gomonkey.ApplyMethodReturn(repo, "Get", mockUser, nil)
		defer patch.Reset()
		info, err := getUserAuth(adminUser, adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
		c.So(err, c.ShouldBeNil)
		_, err = authzBiz.Login(context.TODO(), info)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUserDisabled.Error())
		patch.Reset()
	})
	c.Convey("test login failed with jwt generate error", t, func() {
		info, _ := getUserAuth(adminUser, adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
		patch := gomonkey.ApplyFuncReturn(tiga.GenerateJWT, nil, fmt.Errorf("error generate jwt"))
		defer patch.Reset()
		_, err = authzBiz.Login(context.TODO(), info)

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "error generate jwt")
		patch.Reset()
	})

}

func testLogout(t *testing.T) {
	authzBiz := newAuthzBiz()
	c.Convey("test logout success", t, func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-token", authzStr))
		err := authzBiz.Logout(ctx, &v1.LogoutAPIRequest{})
		c.So(err, c.ShouldBeNil)

		success, err := authzBiz.CheckInBlackList(context.TODO(), tiga.GetMd5(authzStr))
		c.So(err, c.ShouldBeNil)
		c.So(success, c.ShouldBeTrue)
	})
	c.Convey("test logout fail", t, func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-token-user", authzStr))
		err := authzBiz.Logout(ctx, &v1.LogoutAPIRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenMissing.Error())

		err = authzBiz.Logout(context.TODO(), &v1.LogoutAPIRequest{})

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrNoMetadata.Error())

		patch := gomonkey.ApplyFuncReturn((*biz.AuthzUsecase).PutBlackList, fmt.Errorf("error PutBlackList"))
		defer patch.Reset()
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-token", authzStr))
		err = authzBiz.Logout(ctx, &v1.LogoutAPIRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "error PutBlackList")

	})
}
func testDelToken(t *testing.T) {
	c.Convey("test del token", t, func() {
		authzBiz := newAuthzBiz()

		err := authzBiz.DelToken(context.TODO(), authzStr)
		c.So(err, c.ShouldBeNil)
	})
}
func testPutBlackList(t *testing.T) {
	authzBiz := newAuthzBiz()
	c.Convey("test put black list", t, func() {
		token := tiga.GetMd5("test")
		err := authzBiz.PutBlackList(context.TODO(), token)
		c.So(err, c.ShouldBeNil)
		ok, err := authzBiz.CheckInBlackList(context.TODO(), token)
		c.So(err, c.ShouldBeNil)
		c.So(ok, c.ShouldBeTrue)
	})
}
func TestAuthz(t *testing.T) {
	t.Run("test auth seed", testAuthSeed)
	t.Run("test login", testLogin)
	t.Run("test logout", testLogout)
	t.Run("test del token", testDelToken)
	t.Run("test put black list", testPutBlackList)
}
