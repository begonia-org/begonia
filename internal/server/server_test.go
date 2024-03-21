package server

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	dp "github.com/begonia-org/dynamic-proto"
	c "github.com/smartystreets/goconvey/convey"
)

var serverForTest *dp.GatewayServer
var onceServer sync.Once

func RunTestServer() {
	onceServer.Do(func() {
		config := config.ReadConfig("dev")
		serverForTest = New(config, logger.Logger, "0.0.0.0:12140")
		go func() {
			err := serverForTest.Start()
			if err != nil {
				panic(err)
			}
		}()
	})
}
func TestServer(t *testing.T) {
	c.Convey("test server init", t, func() {

		config := config.ReadConfig("dev")
		server := New(config, logger.Logger, "0.0.0.0:12141")
		go func() {
			err := server.Start()
			t.Error(err)
		}()
		// err := server.Start()
		// t.Error(err)
		time.Sleep(4 * time.Second)

		url := "http://127.0.0.1:12140/api/v1/auth/log"
		method := "POST"

		payload := strings.NewReader(`{"timestamp":"1710331340850000"}`)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)
		c.So(err, c.ShouldBeNil)

		res, err := client.Do(req)
		c.So(err, c.ShouldBeNil)
		c.So(res.StatusCode, c.ShouldEqual, 200)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		c.So(err, c.ShouldBeNil)
		c.So(body, c.ShouldNotBeNil)
	})
}
