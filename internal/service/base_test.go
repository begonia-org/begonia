package service_test

import (
	"context"
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
	user "github.com/begonia-org/go-sdk/api/user/v1"
	example "github.com/begonia-org/go-sdk/example"
	"github.com/spark-lence/tiga"
	clientv3 "go.etcd.io/etcd/client/v3"
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

		log.Fatalf("use home idr error: %s", err.Error())
		return
	}
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	op := internal.InitOperatorApp(config.ReadConfig(env))
	_ = op.Init()
	path := filepath.Join(homeDir, ".begonia")
	path = filepath.Join(path, "admin-app.json")
	file, err := os.Open(path)
	if err != nil {

		log.Fatalf("open file %s error: %s", path, err.Error())
		return

	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	app := &api.Apps{}
	err = decoder.Decode(app)
	if err != nil {
		log.Fatalf("decode file %s error: %s", path, err.Error())
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
	// setup()
	RunTestServer()
	time.Sleep(5 * time.Second)

	m.Run()
	log.Printf("All tests passed")
	time.Sleep(20 * time.Second)
	clean()

}

func clean() {
	log.Printf("Cleaned up test data start")

	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := config.ReadConfig(env)

	// cnf:=config.NewConfig(conf)
	rdb := tiga.NewRedisDao(conf)
	luaScript := `
		local prefix = KEYS[1]
		local cursor = "0"
		local count = 100
		repeat
			local result = redis.call("SCAN", cursor, "MATCH", prefix, "COUNT", count)
			cursor = result[1]
			local keys = result[2]
			if #keys > 0 then
				redis.call("DEL", unpack(keys))
			end
		until cursor == "0"
		return "OK"
		`

	_, err := rdb.GetClient().Eval(context.Background(), luaScript, []string{"test:*"}).Result()
	if err != nil {
		log.Fatalf("Could not execute Lua script: %v", err)
	}
	etcd := tiga.NewEtcdDao(conf)
	// 设置前缀
	prefix := "/test"

	// 使用前缀删除所有键
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = etcd.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("Failed to delete keys with prefix %s: %v", prefix, err)
	}

	mysql := tiga.NewMySQLDao(conf)
	mysql.RegisterTimeSerializer()
	err=mysql.GetModel(&user.Users{}).Where("`group` = ?", "test-user-01").Delete(&user.Users{}).Error
	if err != nil {
		log.Fatalf("Failed to delete keys with prefix %s: %v", prefix, err)
	}
	log.Printf("Cleaned up test data")
}
