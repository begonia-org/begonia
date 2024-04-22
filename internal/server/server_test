package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	dp "github.com/begonia-org/dynamic-proto"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/v1"
	"github.com/begonia-org/go-sdk/example"

	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga/loadbalance"
)

var serverForTest *dp.GatewayServer
var onceServer sync.Once

func RunTestServer() {
	onceServer.Do(func() {
		config := config.ReadConfig("dev")
		serverForTest = New(config, logger.Logger, "0.0.0.0:12140")
		go func() {
			err := serverForTest.Start()
			log.Printf("启动服务:%v", err)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
		}()
		go func() {
			example.Run("127.0.0.1:29527")
		}()
		go func() {
			example.RunPlugins("127.0.0.1:9000")
		}()
		go func() {
			example.RunPlugins("127.0.0.1:9001")
		}()

	})
}
func TestMain(m *testing.M) {

	RunTestServer()
	time.Sleep(5 * time.Second)

	m.Run()

}
func TestServer(t *testing.T) {
	c.Convey("test server init", t, func() {

		time.Sleep(3 * time.Second)

		url := "http://127.0.0.1:12140/api/v1/auth/log"
		method := "POST"

		payload := strings.NewReader(`{"timestamp":"1710331340850000"}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)
		req.Header.Add("accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		c.So(err, c.ShouldBeNil)

		res, err := client.Do(req)

		body, err := io.ReadAll(res.Body)
		t.Log(string(body))
		defer res.Body.Close()

		c.So(err, c.ShouldBeNil)
		c.So(res.StatusCode, c.ShouldEqual, 200)

		c.So(err, c.ShouldBeNil)
		c.So(body, c.ShouldNotBeNil)
	})
}

func TestCreateEndpointAPI(t *testing.T) {
	c.Convey("test create endpoint api", t, func() {

		// url := "http://127.0.0.1:12140/api/v1/endpoint/create"
		cli := gosdk.NewBegoniaClient("http://127.0.0.1:12140", "NWkbCslfh9ea2LjVIUsKehJuopPb65fn",
			"oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")
		// basePath, _ := os.Getwd()
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("Failed to retrieve current file path")
		}
		targetFilePath := filepath.Join(filepath.Dir(currentFile), "../../example/protos.tar.gz")

		uri, err := cli.UploadFile(context.TODO(), targetFilePath, "endpoints/protos.tar.gz")
		c.So(err, c.ShouldBeNil)

		rsp, err := cli.CreateEndpoint(context.TODO(), &api.AddEndpointRequest{
			Name:        "test",
			ServiceName: "test",
			Description: "test endpoint",
			ProtoPath:   uri,
			Endpoints: []*api.EndpointMeta{
				{
					Addr: "127.0.0.1:29527",
				},
			},
			Balance: string(loadbalance.RRBalanceType),
			Tags:    []string{"test"},
		})
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:12140/api/v1/example/helloworld", nil)

		c.So(err, c.ShouldBeNil)
		req.Header.Add("accept", "application/json")

		httpRsp, err := http.DefaultClient.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(httpRsp.StatusCode, c.ShouldEqual, 200)
		defer httpRsp.Body.Close()
		body, err := io.ReadAll(httpRsp.Body)
		c.So(err, c.ShouldBeNil)
		response := make(map[string]interface{})
		t.Log(string(body))
		err = json.Unmarshal(body, &response)
		c.So(err, c.ShouldBeNil)
		c.So(response["message"], c.ShouldEqual, "Hello, world!")

	})
}
