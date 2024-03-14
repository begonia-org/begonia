package validator

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/signature"
	"github.com/begonia-org/go-layered-cache/source"
	"github.com/bsm/redislock"

	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	dp "github.com/begonia-org/dynamic-proto"
	"github.com/sirupsen/logrus"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockServerStream struct {
	grpc.ServerStream
	In  metadata.MD
	Out metadata.MD
	ctx context.Context
}



func NewCommonMock() *gomonkey.Patches {
	patches := gomonkey.ApplyFunc((*data.LayeredCache).LoadRemoteCache, func(_ *data.LayeredCache, _ context.Context, _ string) {
	})
	patches.ApplyFunc((*data.Data).DelCache, func(_ *data.Data, _ context.Context, _ string) error {
		return nil
	})
	patches.ApplyFunc((*source.CacheSourceImpl).Del, func(_ *source.CacheSourceImpl, _ context.Context, _ interface{}, _ ...interface{}) error {
		return nil
	})
	// patches.ApplyFunc((*lbf.BloomPubSubImpl).Publish, func(_ *lbf.BloomPubSubImpl, _ context.Context, _ string, _ *lbf.BloomBroadcastMessage) error {
	// 	return nil
	// })
	return patches
}

// RecvMsg mock实现
func (m *MockServerStream) RecvMsg(msg interface{}) error {
	if pb, ok := msg.(*api.EndpointRequest); ok {
		pb.Endpoints = []*api.Endpoints{
			{
				Name:        "test",
				Description: "test",
				Endpoint: []*api.EndpointMeta{
					{
						Weight: 1,
						Addr:   "fffffff",
					},
				},
			},
		}
	}
	return nil
}
func (m *MockServerStream) Context() context.Context {
	return m.ctx
}
func (m *MockServerStream) SetHeader(md metadata.MD) error {
	m.In = metadata.Join(m.In, md)
	return nil
}

func (m *MockServerStream) SendHeader(md metadata.MD) error {
	m.Out = metadata.Join(m.Out, md)
	return nil
}

// SendMsg mock实现
func (m *MockServerStream) SendMsg(msg interface{}) error {
	return nil
}
func MockHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

type mockValidate struct {
	ctx      context.Context
	jwt      string
	validate *APIValidator
}

var onceMockValidate sync.Once
var mockValidator *mockValidate

func newMockValidate() *mockValidate {
	onceMockValidate.Do(func() {
		conf := config.NewConfig(cfg.ReadConfig("dev"))
		log := logrus.New()
		rdb := &tiga.RedisDao{}
		mysql := &tiga.MySQLDao{}
		d := data.NewData(mysql, rdb)
		// pubsub := lbf.NewBloomPubSub(rdb.GetClient(), "gateway-blacklist", "gateway-01", log)
		// bf := lbf.NewLayeredBloomFilter(pubsub, "gateway-blacklist", "gateway-01")
		local := data.NewLayeredCache(context.Background(), d, conf, log)
		repo := data.NewUserRepo(d, logrus.New(), local)
		bizUseCase := biz.NewUsersUsecase(repo, log, crypto.NewUsersAuth(), conf)
		bgCtx := context.Background()
		protos := conf.GetProtosDir()

		pd, err := dp.NewDescription(protos)
		if err != nil {
			log.Fatal(err)
		}
		router := routers.Get()
		router.LoadAllRouters(pd)
		validator := NewAPIValidator(
			rdb,
			log,
			bizUseCase,
			conf,
			mysql,
			local)
		userUseCase := biz.NewUsersUsecase(repo, log, crypto.NewUsersAuth(), conf)
		user := &api.Users{
			Name:     "test",
			Phone:    "1234567890",
			Email:    "test@example.com",
			Password: "123456",
			Uid:      "123456",
			Role:     api.Role_ADMIN,
		}
		jwt, _ := userUseCase.GenerateJWT(bgCtx, user, true)
		mockValidator = &mockValidate{
			ctx:      bgCtx,
			jwt:      jwt,
			validate: validator,
		}
	})
	return mockValidator

}
func TestMain(m *testing.M) {
	patches := NewCommonMock()
	defer patches.Reset()
	code := m.Run()

	os.Exit(code)

}
func TestJWTGrpcValidate(t *testing.T) {
	c.Convey("check grpc jwt token if validate", t, func() {

		ep := []*api.Endpoints{
			{
				Name:        "test",
				Description: "test",
				Endpoint: []*api.EndpointMeta{
					{
						Weight: 1,
						Addr:   "fffffff",
					},
				},
			},
		}
		req := &api.EndpointRequest{
			Endpoints: ep,
		}
		validator := newMockValidate()
		auth := "Bearer " + validator.jwt
		md := metadata.New(map[string]string{"authorization": auth})
		ctx := metadata.NewIncomingContext(validator.ctx, md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/begonia.org.begonia.EndpointService/Create",
		}
		// patch := gomonkey.ApplyFunc((*queue.EsQueue).Put, func(_ *queue.EsQueue, _ interface{}) (bool, uint32) { return false, 0 })
		newCtx := context.TODO()
		_, err := validator.validate.ValidateUnaryInterceptor(ctx, req, info, grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
			// 返回你预期的响应或错误
			newCtx = ctx
			return nil, nil
		}))

		c.So(err, c.ShouldBeNil)
		md, ok := metadata.FromIncomingContext(newCtx)
		c.So(ok, c.ShouldBeTrue)
		c.So(md.Get("x-uid"), c.ShouldNotBeEmpty)
		md = metadata.New(map[string]string{})
		ctx1 := metadata.NewIncomingContext(validator.ctx, md)
		_, err = validator.validate.ValidateUnaryInterceptor(ctx1, req, info, grpc.UnaryHandler(MockHandler))
		c.So(err, c.ShouldNotBeNil)

		md = metadata.New(map[string]string{"authorization": "Bearer " + "ttttfvgsfvd"})
		ctx2 := metadata.NewIncomingContext(validator.ctx, md)

		_, err = validator.validate.ValidateUnaryInterceptor(ctx2, req, info, grpc.UnaryHandler(MockHandler))
		c.So(err, c.ShouldNotBeNil)

	})
}

func TestGrpcJWTExpired(t *testing.T) {
	c.Convey("check jwt token if expired", t, func() {

		ep := []*api.Endpoints{
			{
				Name:        "test",
				Description: "test",
				Endpoint: []*api.EndpointMeta{
					{
						Weight: 1,
						Addr:   "fffffff",
					},
				},
			},
		}
		req := &api.EndpointRequest{
			Endpoints: ep,
		}
		validator := newMockValidate()
		auth := "Bearer " + validator.jwt
		md := metadata.New(map[string]string{"authorization": auth})
		ctx := metadata.NewIncomingContext(validator.ctx, md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/begonia.org.begonia.EndpointService/Create",
		}
		patches := gomonkey.ApplyFunc((*config.Config).GetJWTExpiration, func(_ *config.Config) int64 {
			return 86400 * 100
		})
		patches.ApplyFunc((*APIValidator).JWTLock, func(_ *APIValidator, _ string) (*redislock.Lock, error) {
			return &redislock.Lock{}, nil
		})
		patches.ApplyFunc((*redislock.Lock).Release, func(_ *redislock.Lock, _ context.Context) error {
			return nil
		})
		patches.ApplyFunc((*data.Data).Cache, func(_ *data.Data, _ context.Context, _ string, _ string, _ time.Duration) error {
			return nil

		})
		defer patches.Reset()

		// 等待一段时间，避免刷新token计算出来的时间戳一样
		time.Sleep(3 * time.Second)
		_, err := validator.validate.ValidateUnaryInterceptor(ctx, req, info, grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
			out, ok := metadata.FromOutgoingContext(ctx)
			c.So(ok, c.ShouldBeTrue)
			c.So(out.Get("Authorization"), c.ShouldNotBeEmpty)
			c.So(out.Get("Authorization")[0], c.ShouldNotEqual, auth)
			return nil, nil
		}))
		c.So(err, c.ShouldBeNil)
	})
}

func TestStreamGrpcJWTExpired(t *testing.T) {
	c.Convey("check grpc stream jwt token if expired", t, func() {
		patches := gomonkey.ApplyFunc((*data.LayeredCache).LoadRemoteCache, func(_ *data.LayeredCache, _ context.Context, _ string) {
		})
		patches.ApplyFunc((*data.Data).DelCache, func(_ *data.Data, _ context.Context, _ string) error {
			return nil
		})
		defer patches.Reset()

		validator := newMockValidate()
		auth := "Bearer " + validator.jwt
		md := metadata.New(map[string]string{"authorization": auth})
		ctx := metadata.NewIncomingContext(validator.ctx, md)

		patches.ApplyFunc((*config.Config).GetJWTExpiration, func(_ *config.Config) int64 {
			return 86400 * 100
		})
		patches.ApplyFunc((*APIValidator).JWTLock, func(_ *APIValidator, _ string) (*redislock.Lock, error) {
			return &redislock.Lock{}, nil
		})
		patches.ApplyFunc((*redislock.Lock).Release, func(_ *redislock.Lock, _ context.Context) error {
			return nil
		})
		patches.ApplyFunc((*data.Data).Cache, func(_ *data.Data, _ context.Context, _ string, _ string, _ time.Duration) error {
			return nil

		})
		patches.ApplyFunc((*source.DataSourceFromRedis).TxWriteHandle, func(_ *source.DataSourceFromRedis, _ context.Context, _ *source.TxHandleKeysOptions) error {
			return nil
		})
		defer patches.Reset()
	
		// 等待一段时间，避免刷新token计算出来的时间戳一样
		time.Sleep(3 * time.Second)

		ss := &MockServerStream{
			In:  metadata.MD{},
			Out: metadata.MD{},
			ctx: ctx,
		}
		info := &grpc.StreamServerInfo{
			FullMethod:     "/begonia.org.begonia.EndpointService/Create",
			IsServerStream: true,
			IsClientStream: true,
		}
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			// 在这里触发RecvMsg
			err := stream.RecvMsg(&api.EndpointRequest{})
			out, ok := metadata.FromOutgoingContext(stream.Context())
			c.So(ok, c.ShouldBeTrue)
			c.So(out.Get("Authorization"), c.ShouldNotBeEmpty)
			c.So(out.Get("Authorization")[0], c.ShouldNotEqual, auth)
			return err
		}
		err := validator.validate.ValidateStreamInterceptor(nil, ss, info, handler)
		c.So(err, c.ShouldBeNil)

	})
}

func TestJWTStreamGrpcValidate(t *testing.T) {
	c.Convey("check  stream grpc jwt token if validate", t, func() {

		validator := newMockValidate()
		md := metadata.New(map[string]string{"authorization": "Bearer " + validator.jwt})
		ctx := metadata.NewIncomingContext(validator.ctx, md)
		ss := &MockServerStream{
			In:  metadata.MD{},
			Out: metadata.MD{},
			ctx: ctx,
		}
		info := &grpc.StreamServerInfo{
			FullMethod:     "/begonia.org.begonia.EndpointService/Create",
			IsServerStream: true,
			IsClientStream: true,
		}
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			// 在这里触发RecvMsg
			return stream.RecvMsg(&api.EndpointRequest{})
		}
		err := validator.validate.ValidateStreamInterceptor(nil, ss, info, handler)
		c.So(err, c.ShouldBeNil)
		c.So(ss.In.Get("x-uid"), c.ShouldNotBeEmpty)
		md = metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(validator.ctx, md)
		ss = &MockServerStream{
			In:  metadata.MD{},
			Out: metadata.MD{},
			ctx: ctx,
		}
		err = validator.validate.ValidateStreamInterceptor(nil, ss, info, handler)
		c.So(err, c.ShouldNotBeNil)
	})
}

func TestGrpcAppValidate(t *testing.T) {
	c.Convey("check grpc app key if validate", t, func() {

		validator := newMockValidate()
		app := &api.Apps{
			AccessKey: "tesFFFFFFFFSSt",
			Secret:    "132e423dfwfwefrwefw",
			Appid:     "123456",
			Name:      "test",
		}
		patch2 := gomonkey.ApplyFunc((*APIValidator).getSecret, func(_ *APIValidator, _ context.Context, _ string) (string, error) {
			return app.Secret, nil
		})
		defer patch2.Reset()
		signer := signature.NewAppAuthSigner(app.AccessKey, app.Secret)

		info := &grpc.UnaryServerInfo{
			FullMethod: "/begonia.org.begonia.EndpointService/Create",
		}

		md := metadata.New(map[string]string{
			strings.ToLower(":authority"): "127.0.0.1:9090",
		})

		ep := []*api.Endpoints{
			{
				Name:        "test",
				Description: "test",
				Endpoint: []*api.EndpointMeta{
					{
						Weight: 1,
						Addr:   "fffffff",
					},
				},
			},
		}
		req := &api.EndpointRequest{
			Endpoints: ep,
		}
		ctx := metadata.NewIncomingContext(validator.ctx, md)
		gwRequest, err := signature.NewGatewayRequestFromGrpc(ctx, req, info.FullMethod)
		c.So(err, c.ShouldBeNil)
		err = signer.SignRequest(gwRequest)
		c.So(err, c.ShouldBeNil)
		md = gwRequest.Headers.ToMetadata()
		ctx = metadata.NewIncomingContext(validator.ctx, md)
		_, err = validator.validate.ValidateUnaryInterceptor(ctx, req, info, grpc.UnaryHandler(MockHandler))
		c.So(err, c.ShouldBeNil)
		md = metadata.New(map[string]string{})
		ctx = metadata.NewIncomingContext(validator.ctx, md)
		_, err = validator.validate.ValidateUnaryInterceptor(ctx, req, info, grpc.UnaryHandler(MockHandler))
		c.So(err, c.ShouldNotBeNil)

	})
}

func TestStreamGrpcAppValidate(t *testing.T) {
	c.Convey("check stream grpc app key if validate", t, func() {

		validator := newMockValidate()
		app := &api.Apps{
			AccessKey: "tesFFFFFFFFSSt",
			Secret:    "132e423dfwfwefrwefw",
			Appid:     "123456",
			Name:      "test",
		}
		patch2 := gomonkey.ApplyFunc((*APIValidator).getSecret, func(_ *APIValidator, _ context.Context, _ string) (string, error) {
			return app.Secret, nil
		})
		defer patch2.Reset()
		signer := signature.NewAppAuthSigner(app.AccessKey, app.Secret)
		md := metadata.New(map[string]string{
			strings.ToLower(":authority"): "127.0.0.1:9090",
		})

		ctx := metadata.NewIncomingContext(validator.ctx, md)

		info := &grpc.StreamServerInfo{
			FullMethod:     "/begonia.org.begonia.EndpointService/Create",
			IsServerStream: true,
			IsClientStream: true,
		}

		req := &api.EndpointRequest{
			Endpoints: []*api.Endpoints{
				{
					Name:        "test",
					Description: "test",
					Endpoint: []*api.EndpointMeta{
						{
							Weight: 1,
							Addr:   "fffffff",
						},
					},
				},
			},
		}

		handler := func(srv interface{}, stream grpc.ServerStream) error {
			// 在这里触发RecvMsg
			return stream.RecvMsg(&api.EndpointRequest{})
		}
		gwRequest, err := signature.NewGatewayRequestFromGrpc(ctx, req, info.FullMethod)
		c.So(err, c.ShouldBeNil)
		err = signer.SignRequest(gwRequest)
		c.So(err, c.ShouldBeNil)

		md = gwRequest.Headers.ToMetadata()
		ctx = metadata.NewIncomingContext(validator.ctx, md)
		ss := &MockServerStream{
			In:  metadata.MD{},
			Out: metadata.MD{},
			ctx: ctx,
		}

		err = validator.validate.ValidateStreamInterceptor(nil, ss, info, handler)
		c.So(err, c.ShouldBeNil)
		patch3 := gomonkey.ApplyFunc((*APIValidator).getSignature, func(_ *APIValidator, _ string) string {
			return "testcc98cccccccc"
		})
		defer patch3.Reset()
		err = validator.validate.ValidateStreamInterceptor(nil, ss, info, handler)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrAppSignatureInvalid.Error())
	})
}
