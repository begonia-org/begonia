# Makefile

# 定义变量
PROTOC = protoc
PYTHON = python3
TS_PROTO_PLUGIN = $(shell which protoc-gen-ts_proto)

# 目标文件夹
PROTO_DIR = ./protos
OUTPUT_DIR = ./$(OUT)

# 文件列表
GO_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
PY_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
TS_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
COMMON_FILES = $(wildcard $(PROTO_DIR)/*/*.proto)
COMMON_PROTO_FILES = $(notdir $(COMMON_FILES))
# 通用参数
GO_ARGS = --go_out=$(OUTPUT_DIR) --go_opt=paths=source_relative \
          --go-grpc_out=$(OUTPUT_DIR) --grpc-gateway_out=$(OUTPUT_DIR) \
          --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --go-grpc_opt=paths=source_relative



common:
	@mkdir -p $(OUTPUT_DIR)
	@sed -i 's|github.com/wetrycode/begonia|$(PKG)|g' go.mod

.PHONY: make_dir generate go python ts common clean
clean:
	@echo $(COMMON_PROTO_FILES) $(COMMON_FILES) “公共文件”
	@rm -rf $(COMMON_PROTO_FILES)
gomain:
	@echo "package main" > main.go
	@echo "func main() {}" >> main.go
	go mod tidy
	go build -buildmode=plugin -o $(PLUGIN_NAME).so .
# 生成所有代码
generate: make_dir go python ts common gomain clean

# 生成 Go 代码
go: $(GO_PROTO_FILES) | common
	$(PROTOC) -I=$(PROTO_DIR) $(GO_ARGS) $(COMMON_FILES) $?
	protoc-go-inject-tag -input="$(OUTPUT_DIR)/*.pb.go"
	@$(MAKE) clean
	@$(MAKE) gomain