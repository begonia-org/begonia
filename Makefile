OUTPUT_DIR=./bin
GO_BIN_DIR := $(shell go env GOPATH)/bin
version=$(shell git describe --tags --abbrev=0)
commit=$(shell git rev-parse --short HEAD)
build_time=$(shell date '+%Y%m%d%H%M%S')
build:
	go build -ldflags -X=github.com/begonia-org/begonia.Version=$(version)\ -X=github.com/begonia-org/begonia.BuildTime=$(build_time)\ -X=github.com/begonia-org/begonia.Commit=$(commit) -o $(OUTPUT_DIR)/begonia cmd/begonia/*.go
install:
	go install -ldflags -X=github.com/begonia-org/begonia.Version=$(version)\ -X=github.com/begonia-org/begonia.BuildTime=$(build_time)\ -X=github.com/begonia-org/begonia.Commit=$(commit) cmd/begonia/*.go
all: build install
.DEFAULT_GOAL := all
