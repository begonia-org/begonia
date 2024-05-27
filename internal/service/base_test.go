package service_test

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	example "github.com/begonia-org/go-sdk/example"
)

var onceExampleServer sync.Once
var onceServer sync.Once
var shareEndpoint = ""

// "NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe"
var apiAddr = "http://127.0.0.1:12140"
var accessKey = "P6DKyCxyz2ewz4brsFPJtusyMLYh4uex"
var secret = "5Ygsu0UiZJPil4eWf5XcXec3gYKcWiCfJtyGGQbHekUlEMf7KXM7pSWdev8UfhiI"
var sdkAPPID = "449250203195674624"

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
func readInitAPP() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	path := filepath.Join(homeDir, ".begonia")
	path = filepath.Join(path, "admin-app.json")
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf(err.Error())
		return

	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	app := &api.Apps{}
	err = decoder.Decode(app)
	if err != nil {
		log.Fatalf(err.Error())
		return

	}
	accessKey = app.AccessKey
	secret = app.Secret
	sdkAPPID = app.Appid
}
func RunTestServer() {
	log.Printf("run test server")
	onceServer.Do(func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		// log.Printf("env: %s", env)
		config := config.ReadConfig(env)
		go func() {

			worker := internal.New(config, gateway.Log, "0.0.0.0:12140")
			worker.Start()

		}()
		runExampleServer()
		time.Sleep(2 * time.Second)

	})
}

func TestMain(m *testing.M) {
	readInitAPP()
	RunTestServer()
	time.Sleep(5 * time.Second)

	m.Run()

}
