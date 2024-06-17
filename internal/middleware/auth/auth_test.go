package auth_test

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	"github.com/begonia-org/begonia/internal/middleware/auth"
	"github.com/begonia-org/begonia/internal/pkg"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/internal/pkg/utils"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	v1 "github.com/begonia-org/go-sdk/api/user/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testStream struct {
	ctx context.Context
}

func (t *testStream) SetHeader(metadata.MD) error {
	return nil
}
func (t *testStream) SendHeader(metadata.MD) error {
	return nil
}
func (t *testStream) SetTrailer(metadata.MD) {

}
func (t *testStream) Context() context.Context {
	return t.ctx

}
func (t *testStream) SendMsg(m interface{}) error {
	return nil
}
func (t *testStream) RecvMsg(m interface{}) error {
	return nil
}

type greeterSayHelloWebsocketServer struct {
	grpc.ServerStream
}

func (x *greeterSayHelloWebsocketServer) Send(m *hello.HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greeterSayHelloWebsocketServer) Recv() (*hello.HelloRequest, error) {
	m := new(hello.HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
func (x *greeterSayHelloWebsocketServer) Context() context.Context {
	return x.ServerStream.Context()

}
func readInitAPP() (string, string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
		return "", "", ""
	}
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	path := filepath.Join(homeDir, ".begonia")
	path = filepath.Join(path, fmt.Sprintf("admin-app.%s.json", env))
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf(err.Error())
		return "", "", ""

	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	app := &api.Apps{}
	err = decoder.Decode(app)
	if err != nil {
		log.Fatalf(err.Error())
		return "", "", ""

	}
	return app.AccessKey, app.Secret, app.Appid
}
func getJWT() string {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	user := data.NewUserRepo(config, gateway.Log)
	userAuth := crypto.NewUsersAuth(cnf)
	authzRepo := data.NewAuthzRepo(config, gateway.Log)
	appRepo:=data.NewAppRepo(config,gateway.Log)
	authz := biz.NewAuthzUsecase(authzRepo, user,appRepo, gateway.Log, userAuth, cnf)
	adminUser := cnf.GetDefaultAdminName()
	adminPasswd := cnf.GetDefaultAdminPasswd()
	_, filename, _, _ := runtime.Caller(0)
	file := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "cert", "auth_public_key.pem")
	pubKey, err := utils.LoadPublicKeyFromFile(file)
	if err != nil {
		log.Printf("load public key failed: %v", err)
		return ""

	}
	seedToken := time.Now().UnixMilli() * 1000
	seed, _ := authz.AuthSeed(context.TODO(), &v1.AuthLogAPIRequest{
		Token: fmt.Sprintf("%d", seedToken),
	})
	seedTimestampToken := fmt.Sprintf("%d", seedToken)
	seedAuthToken := seed
	info, _ := getUserAuth(adminUser, adminPasswd, pubKey, seedAuthToken, seedTimestampToken)
	token, _ := authz.Login(context.TODO(), info)
	return token.Token

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
func getMid() gosdk.LocalPlugin {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	repo := data.NewAppRepo(config, gateway.Log)
	cnf := cfg.NewConfig(config)

	akBiz := biz.NewAccessKeyAuth(repo, cnf, gateway.Log)
	user := data.NewUserRepo(config, gateway.Log)
	userAuth := crypto.NewUsersAuth(cnf)
	authzRepo := data.NewAuthzRepo(config, gateway.Log)
	appRepo:=data.NewAppRepo(config,gateway.Log)
	authz := biz.NewAuthzUsecase(authzRepo, user,appRepo, gateway.Log, userAuth, cnf)
	jwt := auth.NewJWTAuth(cnf, tiga.NewRedisDao(config), authz, gateway.Log)
	ak := auth.NewAccessKeyAuth(akBiz, cnf, gateway.Log)

	apiKey := auth.NewApiKeyAuth(cnf,authz)
	mid := auth.NewAuth(ak, jwt, apiKey)
	return mid

}
func TestUnaryInterceptor(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)

	mid := getMid()
	ctx := context.Background()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		md,ok:=metadata.FromIncomingContext(ctx)
		if !ok{
			return nil,fmt.Errorf("no metadata")
		}
		if identify:=md.Get(gosdk.HeaderXIdentity);len(identify)==0||identify[0]==""{
			return nil,fmt.Errorf("no app identity")
		}
		XAccessKey:=md.Get(gosdk.HeaderXAccessKey)
		XApiKey:=md.Get(gosdk.HeaderXApiKey)
		XAuthz:=md.Get("authorization")
		if len(XAccessKey)==0 && len(XApiKey)==0 && len(XAuthz)==0{
			return nil,fmt.Errorf("no app auth key")
		}
		return nil, nil
	}
	R := routers.Get()
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

	pd, _ := gateway.NewDescription(pbFile)
	R.LoadAllRouters(pd)
	c.Convey("TestUnaryInterceptor api-key", t, func() {

		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(gosdk.HeaderXApiKey, cnf.GetAdminAPIKey()))
		_, err := mid.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, handler)
		c.So(err, c.ShouldBeNil)

		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(gosdk.HeaderXApiKey, "123"))
		_, err = mid.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, handler)
		c.So(err, c.ShouldNotBeNil)
		mid.SetPriority(1)
		c.So(mid.Priority(), c.ShouldEqual, 1)
		c.So(mid.Name(), c.ShouldEqual, "auth")

	})
	c.Convey("TestUnaryInterceptor jwt", t, func() {
		ctx = context.Background()
		token := getJWT()
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
		_, err := mid.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, handler)
		c.So(err, c.ShouldBeNil)
		ctx = context.Background()
		token = "123"
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
		_, err = mid.UnaryInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, handler)
		c.So(err, c.ShouldNotBeNil)
	})
	c.Convey("TestUnaryInterceptor aksk", t, func() {
		u := &v1.Users{}
		bData, _ := json.Marshal(u)
		req, _ := http.NewRequest(http.MethodPost, "/test/post", bytes.NewReader(bData))
		req.Header.Set("Content-Type", "application/json")
		access, secret, appid := readInitAPP()
		sgin := gosdk.NewAppAuthSigner(access, secret)
		gwReq, err := gosdk.NewGatewayRequestFromHttp(req)
		c.So(err, c.ShouldBeNil)
		err = sgin.SignRequest(gwReq)
		c.So(err, c.ShouldBeNil)
		md := metadata.New(nil)

		headers := gwReq.Headers
		for _, k := range headers.Keys() {
			md.Append(k, headers.Get(k))
		}
		md.Append("uri", "/test/post")
		md.Append("x-http-method", http.MethodPost)
		patch := gomonkey.ApplyFuncReturn((*biz.AccessKeyAuth).GetSecret, secret, nil)
		patch = patch.ApplyFuncReturn((*biz.AccessKeyAuth).GetAppOwner, appid, nil)
		defer patch.Reset()
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err = mid.UnaryInterceptor(ctx, u, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, handler)
		c.So(err, c.ShouldBeNil)

		sign1 := gosdk.NewAppAuthSigner("ASDASDCASDFQ", "ASDASDCASDFQ")
		gwReq1, err := gosdk.NewGatewayRequestFromHttp(req)
		c.So(err, c.ShouldBeNil)
		err = sign1.SignRequest(gwReq1)
		c.So(err, c.ShouldBeNil)
		md1 := metadata.New(nil)
		headers1 := gwReq1.Headers
		for _, k := range headers1.Keys() {
			md1.Append(k, headers1.Get(k))

		}
		md1.Append("uri", "/test/post")
		md1.Append("x-http-method", http.MethodPost)
		ctx1 := metadata.NewIncomingContext(context.Background(), md1)
		_, err = mid.UnaryInterceptor(ctx1, u, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, handler)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrAppSignatureInvalid.Error())

	})
}

func TestStreamInterceptor(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	mid := getMid()
	ctx := context.Background()

	R := routers.Get()
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

	pd, err := gateway.NewDescription(pbFile)
	R.LoadAllRouters(pd)
	c.Convey("TestStreamInterceptor apikey", t, func() {

		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(gosdk.HeaderXApiKey, cnf.GetAdminAPIKey()))
		err = mid.StreamInterceptor(&v1.Users{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			err:= ss.RecvMsg(srv)
			md,_:=metadata.FromIncomingContext(ss.Context())
			if identify:=md.Get(gosdk.HeaderXIdentity);len(identify)==0||identify[0]==""{
				return fmt.Errorf("no app identity")
			}
			if xAppKey:=md.Get(gosdk.HeaderXApiKey);len(xAppKey)==0||xAppKey[0]==""{
				return fmt.Errorf("no app key")
			}
			return err
		})
		c.So(err, c.ShouldBeNil)
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(gosdk.HeaderXApiKey, "cnf.GetAdminAPIKey()"))
		err = mid.StreamInterceptor(&v1.Users{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)

	})
	c.Convey("TestStreamInterceptor jwt", t, func() {
		jwt := getJWT()
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+jwt))
		err = mid.StreamInterceptor(&v1.Users{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			err:= ss.RecvMsg(srv)
			md,_:=metadata.FromIncomingContext(ss.Context())
			if identify:=md.Get(gosdk.HeaderXIdentity);len(identify)==0||identify[0]==""{
				return fmt.Errorf("no app identity")
			}
			if xAuthorization:=md.Get("authorization");len(xAuthorization)==0||xAuthorization[0]==""{
				return fmt.Errorf("no jwt key")
			}
			return err
		})
		c.So(err, c.ShouldBeNil)
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer 123"))
		err = mid.StreamInterceptor(&v1.Users{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
	})

	c.Convey("TestUnaryInterceptor aksk", t, func() {
		u := &v1.Users{}
		bData, _ := json.Marshal(u)
		req, _ := http.NewRequest(http.MethodPost, "/test/post", bytes.NewReader(bData))
		req.Header.Set("Content-Type", "application/json")
		access, secret, appid := readInitAPP()
		sgin := gosdk.NewAppAuthSigner(access, secret)
		gwReq, err := gosdk.NewGatewayRequestFromHttp(req)
		c.So(err, c.ShouldBeNil)
		err = sgin.SignRequest(gwReq)
		c.So(err, c.ShouldBeNil)
		md := metadata.New(nil)

		headers := gwReq.Headers
		for _, k := range headers.Keys() {
			// t.Logf("header:%s,value:%s", k, headers.Get(k))
			md.Append(k, headers.Get(k))
		}
		md.Append("uri", "/test/post")
		md.Append("x-http-method", http.MethodPost)
		patch := gomonkey.ApplyFuncReturn((*biz.AccessKeyAuth).GetSecret, secret, nil)
		patch = patch.ApplyFuncReturn((*biz.AccessKeyAuth).GetAppOwner, appid, nil)
		defer patch.Reset()
		ctx := metadata.NewIncomingContext(context.Background(), md)
		err = mid.StreamInterceptor(u, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/POST"}, func(srv interface{}, ss grpc.ServerStream) error {
			err:= ss.RecvMsg(srv)
			md,ok:=metadata.FromIncomingContext(ss.Context())
			if !ok{
				return fmt.Errorf("no metadata")
			}
			if identify:=md.Get(gosdk.HeaderXIdentity);len(identify)==0||identify[0]==""{
				return fmt.Errorf("no app identity")
			}
			if xAccessKey:=md.Get(gosdk.HeaderXAccessKey);len(xAccessKey)==0||xAccessKey[0]==""{
				t.Logf("error metadata:%v",md)
				return fmt.Errorf("no app access key")
			}
			return err
		})
		c.So(err, c.ShouldBeNil)

		sign1 := gosdk.NewAppAuthSigner("ASDASDCASDFQ", "ASDASDCASDFQ")
		gwReq1, err := gosdk.NewGatewayRequestFromHttp(req)
		c.So(err, c.ShouldBeNil)
		err = sign1.SignRequest(gwReq1)
		c.So(err, c.ShouldBeNil)
		md1 := metadata.New(nil)
		headers1 := gwReq1.Headers
		for _, k := range headers1.Keys() {
			md1.Append(k, headers1.Get(k))

		}
		md1.Append("uri", "/test/post")
		md1.Append("x-http-method", http.MethodPost)
		ctx1 := metadata.NewIncomingContext(context.Background(), md1)
		err = mid.StreamInterceptor(&v1.Users{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: ctx1}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrAppSignatureInvalid.Error())
	})
}
func TestTestUnaryInterceptorErr(t *testing.T) {
	mid := getMid()
	ctx := context.Background()

	R := routers.Get()
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

	pd, err := gateway.NewDescription(pbFile)
	R.LoadAllRouters(pd)
	c.Convey("TestUnaryInterceptor err", t, func() {
		_, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/NO"}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldBeNil)

		_, err = mid.UnaryInterceptor(ctx, &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err.Error(), c.ShouldContainSubstring, "metadata not exists in context")

		_, err = mid.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenMissing.Error())
	})
}

func TestStreamInterceptorErr(t *testing.T) {
	mid := getMid()

	R := routers.Get()
	_, filename, _, _ := runtime.Caller(0)
	pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

	pd, err := gateway.NewDescription(pbFile)
	R.LoadAllRouters(pd)
	c.Convey("TestStreamInterceptor err", t, func() {
		err = mid.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: context.Background()}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/NO"}, func(srv any, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldBeNil)

		err = mid.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: context.Background()}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv any, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, fmt.Errorf("metadata not exists in context").Error())

		err = mid.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", ""))}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv any, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenMissing.Error())

	})

}
