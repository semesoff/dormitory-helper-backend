PB_OUT=./generated/proto
PROTO_SRC=./proto
GOOGLEAPIS_DIR ?= ./proto:$(HOME)/proto/protoc-gen-validate:$(HOME)/proto/googleapis:$(HOME)/proto/grpc-gateway:$(GOPATH)

.PHONY: all
all: proto

.PHONY: proto-user
proto-user:
	rm -rf $(PB_OUT)/user
	mkdir -p $(PB_OUT)/user
	# запускаем protoc из директории `proto`, тогда не добавится лишний `proto/`
	cd $(PROTO_SRC) && \
		protoc -I . -I ../$(GOOGLEAPIS_DIR) \
		--go_out=paths=source_relative:../$(PB_OUT) \
		--go-grpc_out=paths=source_relative:../$(PB_OUT) \
		--grpc-gateway_out=paths=source_relative:../$(PB_OUT) \
		user/*.proto

.PHONY: proto-laundry
proto-laundry:
	rm -rf $(PB_OUT)/laundry
	mkdir -p $(PB_OUT)/laundry
	cd $(PROTO_SRC) && \
		protoc -I . -I ../$(GOOGLEAPIS_DIR) \
		--go_out=paths=source_relative:../$(PB_OUT) \
		--go-grpc_out=paths=source_relative:../$(PB_OUT) \
		--grpc-gateway_out=paths=source_relative:../$(PB_OUT) \
		laundry/*.proto

.PHONY: proto-kitchen
proto-kitchen:
	rm -rf $(PB_OUT)/kitchen
	mkdir -p $(PB_OUT)/kitchen
	cd $(PROTO_SRC) && \
		protoc -I . -I ../$(GOOGLEAPIS_DIR) \
		--go_out=paths=source_relative:../$(PB_OUT) \
		--go-grpc_out=paths=source_relative:../$(PB_OUT) \
		--grpc-gateway_out=paths=source_relative:../$(PB_OUT) \
		kitchen/*.proto

.PHONY: proto
proto: proto-user proto-laundry proto-kitchen

.PHONY: setup-googleapis
setup-googleapis:
	@echo "Cloning googleapis..."
	@if [ ! -d "$(GOOGLEAPIS_DIR)" ]; then \
		git clone https://github.com/googleapis/googleapis.git $(GOOGLEAPIS_DIR); \
	else \
		echo "googleapis already exists"; \
	fi

.PHONY: install-proto-tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

.PHONY: run-backend
run-backend:
	go run cmd/app/app.go

.PHONY: run
run: run-backend
