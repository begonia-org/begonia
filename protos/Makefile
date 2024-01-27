# Makefile

# 定义变量
PROTOC = protoc
PYTHON = python3
TS_PROTO_PLUGIN = $(shell which protoc-gen-ts_proto)

# 目标文件夹
PROTO_DIR = .
OUTPUT_DIR = ../api/v1

# 文件列表
GO_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
PY_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
TS_PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)
COMMON_FILES = $(wildcard $(PROTO_DIR)/*/*.proto)
COMMON_PROTO_FILES = $(notdir $(COMMON_FILES))

# 通用参数
GO_ARGS = --go_out=$(OUTPUT_DIR) --go_opt=paths=source_relative \
          --go-grpc_out=$(OUTPUT_DIR) --grpc-gateway_out=$(OUTPUT_DIR) \
          --grpc-gateway_opt=paths=source_relative --go-grpc_opt=paths=source_relative
PY_ARGS = -m grpc_tools.protoc -I=$(PROTO_DIR) --python_out=$(OUTPUT_DIR) \
          --pyi_out=$(OUTPUT_DIR) --grpc_python_out=$(OUTPUT_DIR) 
TS_ARGS = --plugin="protoc-gen-ts=$(TS_PROTO_PLUGIN)" \
          --ts_proto_opt=esModuleInterop=true --ts_proto_opt=paths=source_relative \
          --ts_proto_opt=snakeToCamel=false --ts_proto_opt=oneof=unions \
          --ts_proto_out=$(OUTPUT_DIR)



common:
	@mkdir -p $(OUTPUT_DIR)
.PHONY: make_dir generate go python ts common
clean:
	@echo $(COMMON_PROTO_FILES) $(COMMON_FILES) “公共文件”
	@rm -rf $(COMMON_PROTO_FILES)
# 生成所有代码
generate: make_dir go python ts common

# 生成 Go 代码
go: $(GO_PROTO_FILES) | common
	$(PROTOC) -I=$(PROTO_DIR) $(GO_ARGS) $?
	protoc-go-inject-tag -input="$(OUTPUT_DIR)/*.pb.go"
	@$(MAKE) clean

