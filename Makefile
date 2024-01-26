OUTPUT_DIR=./bin

create_output:
	mkdir -p $(OUTPUT_DIR)

make_proto:
	@echo "Running make_proto..."
	$(MAKE) -C protos go

make_server:
	@echo "Building the server..."
	go build -o $(OUTPUT_DIR)/gateway ./main.go

all: create_output make_proto make_server

.DEFAULT_GOAL := all