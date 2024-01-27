package config

import (
	"path/filepath"
	"runtime"

	"github.com/spark-lence/tiga"
)

func ReadConfig(env string) *tiga.Configuration {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return tiga.InitSettings(env, dir)
}
