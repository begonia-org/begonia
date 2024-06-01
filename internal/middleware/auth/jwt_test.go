package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	hello "github.com/begonia-org/go-sdk/api/example/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/bsm/redislock"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func generateJWT(cnf *cfg.Config, user *api.Users, exp time.Duration, notBefore int64, issuer string, isKeepLogin bool) string {
	secret := cnf.GetString("auth.jwt_secret")
	validateToken := tiga.ComputeHmacSha256(fmt.Sprintf("%s:%d", user.Uid, time.Now().Unix()), secret)
	payload := &api.BasicAuth{
		Uid:         user.Uid,
		Name:        user.Name,
		Role:        user.Role,
		Audience:    user.Name,
		Expiration:  time.Now().Add(exp).Unix(),
		Issuer:      issuer,
		NotBefore:   notBefore,
		IssuedAt:    time.Now().Unix(),
		IsKeepLogin: isKeepLogin,
		Token:       validateToken,
	}
	// err := u.repo.DelToken(ctx, u.config.GetUserBlackListKey(user.Uid))

	token, err := tiga.GenerateJWT(payload, secret)
	if err != nil {
		return ""

	}
	return token
}
func TestJWTUnaryInterceptor(t *testing.T) {
	c.Convey("TestJWTUnaryInterceptor", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)

		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

		pd, _ := gateway.NewDescription(pbFile)
		R.LoadAllRouters(pd)
		user := data.NewUserRepo(config, gateway.Log)
		userAuth := crypto.NewUsersAuth(cnf)
		authzRepo := data.NewAuthzRepo(config, gateway.Log)
		authz := biz.NewAuthzUsecase(authzRepo, user, gateway.Log, userAuth, cnf)
		jwt := auth.NewJWTAuth(cnf, tiga.NewRedisDao(config), authz, gateway.Log)
		jwt.SetPriority(1)
		c.So(jwt.Priority(), c.ShouldEqual, 1)
		c.So(jwt.Name(), c.ShouldEqual, "jwt_auth")
		_, err := jwt.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/inte/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err, c.ShouldBeNil)

		_, err = jwt.UnaryInterceptor(context.Background(), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "metadata not exists in context")

		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("test", "test")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil

		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "token not exists in context")

		token := generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 2*time.Second, time.Now().Unix(), "gateway", false)
		time.Sleep(3 * time.Second)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenExpired.Error())

		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 30*time.Second, time.Now().Add(30*time.Second).Unix(), "gateway", true)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenNotActive.Error())

		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 30*time.Second, time.Now().Unix(), "tester", true)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenIssuer.Error())

		patch := gomonkey.ApplyFuncReturn((*biz.AuthzUsecase).CheckInBlackList, true, fmt.Errorf("check in black list error"))
		defer patch.Reset()

		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 30*time.Second, time.Now().Unix(), "gateway", true)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrTokenBlackList.Error())
		patch.Reset()
		patch1 := gomonkey.ApplyFuncReturn((*biz.AuthzUsecase).CheckInBlackList, false, nil)
		defer patch1.Reset()
		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 3*time.Second, time.Now().Unix(), "gateway", true)
		time.Sleep(2*time.Second + 600*time.Millisecond)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})

		patch1.Reset()
		c.So(err, c.ShouldBeNil)

		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "Bearer ")), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrHeaderTokenFormat.Error())

		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 30*time.Second, time.Now().Unix(), "gateway", true)
		cases := []struct {
			patch  interface{}
			err    error
			output []interface{}
		}{
			{
				patch:  tiga.ComputeHmacSha256,
				err:    pkg.ErrTokenInvalid,
				output: []interface{}{""},
			},
			{
				patch:  tiga.Base64URL2Bytes,
				err:    pkg.ErrAuthDecrypt,
				output: []interface{}{nil, fmt.Errorf("base64 decode error")},
			},
			{
				patch:  json.Unmarshal,
				err:    pkg.ErrDecode,
				output: []interface{}{fmt.Errorf("json unmarshal error")},
			},
		}
		for _, v := range cases {
			patch := gomonkey.ApplyFuncReturn(v.patch, v.output...)
			defer patch.Reset()
			_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
				FullMethod: "/integration.TestService/Get",
			}, func(ctx context.Context, req any) (any, error) {
				return nil, nil
			})
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, v.err.Error())
			patch.Reset()
		}

		patch3 := gomonkey.ApplyFuncReturn((*auth.JWTAuth).JWTLock, nil, redislock.ErrNotObtained)
		patch4 := gomonkey.ApplyFuncReturn((*biz.AuthzUsecase).CheckInBlackList, false, nil)
		defer patch3.Reset()
		defer patch4.Reset()
		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 3*time.Second, time.Now().Unix(), "gateway", true)
		time.Sleep(2*time.Second + 600*time.Millisecond)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})

		patch3.Reset()
		c.So(err, c.ShouldBeNil)
		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 3*time.Second, time.Now().Unix(), "gateway", true)
		patch5 := gomonkey.ApplyFuncReturn(tiga.GenerateJWT, "", fmt.Errorf("generate jwt error"))
		defer patch5.Reset()

		time.Sleep(2*time.Second + 600*time.Millisecond)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{
			FullMethod: "/integration.TestService/Get",
		}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "generate jwt error")
		patch5.Reset()

		token = generateJWT(cnf, &api.Users{
			Uid:      "1",
			Name:     "test",
			Password: "TEST",
		}, 3*time.Second, time.Now().Unix(), "gateway", true)

		patch6 := gomonkey.ApplyFunc(tiga.GenerateJWT, func(v interface{}, secret string) (string, error) {
			panic("test")
		})
		defer patch6.Reset()
		time.Sleep(2*time.Second + 600*time.Millisecond)
		_, err = jwt.UnaryInterceptor(metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token))), &hello.HelloRequest{}, &grpc.UnaryServerInfo{FullMethod: "/integration.TestService/Get"}, func(ctx context.Context, req any) (any, error) {
			return nil, nil
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test")
		patch6.Reset()

	})
}

func TestJWTStreamInterceptor(t *testing.T) {
	c.Convey("TestJWTStreamInterceptor", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)

		R := routers.Get()
		_, filename, _, _ := runtime.Caller(0)
		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename)))), "testdata")

		pd, _ := gateway.NewDescription(pbFile)
		R.LoadAllRouters(pd)
		user := data.NewUserRepo(config, gateway.Log)
		userAuth := crypto.NewUsersAuth(cnf)
		authzRepo := data.NewAuthzRepo(config, gateway.Log)
		authz := biz.NewAuthzUsecase(authzRepo, user, gateway.Log, userAuth, cnf)
		jwt := auth.NewJWTAuth(cnf, tiga.NewRedisDao(config), authz, gateway.Log)
		err := jwt.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-api-key", cnf.GetAdminAPIKey())),
		}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET234dd"}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn((*auth.JWTAuth).StreamRequestBefore, nil, fmt.Errorf("test"))
		defer patch.Reset()
		err = jwt.StreamInterceptor(&hello.HelloRequest{}, &greeterSayHelloWebsocketServer{ServerStream: &testStream{}}, &grpc.StreamServerInfo{FullMethod: "/INTEGRATION.TESTSERVICE/GET"}, func(srv interface{}, ss grpc.ServerStream) error {
			return ss.RecvMsg(srv)
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test")
		patch.Reset()

		err = jwt.StreamResponseAfter(context.TODO(), auth.NewGrpcStream(&testStream{}, "", context.TODO(), nil), nil)
		c.So(err, c.ShouldBeNil)
	})
}
