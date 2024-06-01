package data

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/spark-lence/tiga"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// func TestCreateInBatches(t *testing.T) {
func TestMain(m *testing.M) {
	log.Printf("Start testing")
	setup()
	code := m.Run()
	log.Printf("All tests passed with code %d", code)
}

func setup() {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	conf := cfg.ReadConfig(env)

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
}
