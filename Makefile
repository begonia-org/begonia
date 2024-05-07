OUTPUT_DIR=./bin
version=$(shell git describe --tags --abbrev=0)
commit=$(shell git rev-parse --short HEAD)
build_time=$(shell date '+%Y%m%d%H%M%S')
build:
	go build -ldflags -X=github.com/begonia-org/begonia.Version=$(version)\ -X=github.com/begonia-org/begonia.BuildTime=$(build_time)\ -X=github.com/begonia-org/begonia.Commit=$(commit) -o $(OUTPUT_DIR)/begonia cmd/gateway/main.go

.DEFAULT_GOAL := build
