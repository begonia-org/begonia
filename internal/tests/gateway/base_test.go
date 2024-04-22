package gateway_test

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	example "github.com/begonia-org/go-sdk/example"
)

var onceExampleServer sync.Once
var onceServer sync.Once
var shareEndpoint = ""

func runExampleServer() {
	onceExampleServer.Do(func() {
		// run example server
		go example.Run(":21213")
		go example.Run(":21214")
		go example.Run(":21215")

		go example.RunPlugins(":21216")
		go example.RunPlugins(":21217")
	})

}

func RunTestServer() {
	log.Printf("run test server")
	onceServer.Do(func() {
		config := config.ReadConfig("dev")
		go func() {

			worker := internal.New(config, logger.Log, "0.0.0.0:12140")
			err := worker.Start()
			if err != nil {
				panic(err)
			}
		}()
		runExampleServer()
		time.Sleep(2 * time.Second)

	})
}

func TestMain(m *testing.M) {

	RunTestServer()
	time.Sleep(5 * time.Second)

	m.Run()

}
