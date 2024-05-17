# Makefile

# 定义变量
PROTOC = protoc
PYTHON = python3
TS_PROTO_PLUGIN = $(shell which protoc-gen-ts_proto)

# 目标文件夹
PROTO_DIR = ./
OUTPUT_DIR = $(OUT)

# 文件列表
GO_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
# 通用参数
GO_ARGS = --go_out=$(OUTPUT_DIR) --go_opt=paths=source_relative \
          --go-grpc_out=$(OUTPUT_DIR) --grpc-gateway_out=$(OUTPUT_DIR) \
          --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --go-grpc_opt=paths=source_relative 



common:
	@mkdir -p $(OUTPUT_DIR)

.PHONY: generate go common

# 生成所有代码
generate: go common

# 生成 Go 代码
go: $(GO_PROTO_FILES) | common
	$(PROTOC) -I=$(PROTO_DIR) $(GO_ARGS)$?
	protoc-go-inject-tag -input="$(OUTPUT_DIR)/*.pb.go"