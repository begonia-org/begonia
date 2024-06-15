package config

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/spark-lence/tiga"
)

func ReadConfig(env string) *tiga.Configuration {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	log.Printf("config dir: %s", dir)
	return tiga.InitSettings(env, dir)
}
func ReadConfigWithDir(env, filename string) *tiga.Configuration {
	dir := filepath.Dir(filename)
	log.Printf("config dir: %s", dir)
	return tiga.InitSettings(env, dir)
}
