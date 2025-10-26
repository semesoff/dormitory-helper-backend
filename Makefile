PB_OUT=./generated/proto
PROTO_SRC=./proto

.PHONY: proto-user
proto-user:
	rm -rf $(PB_OUT)/user
	mkdir -p $(PB_OUT)/user
	# запускаем protoc из директории `proto`, тогда не добавится лишний `proto/`
	cd $(PROTO_SRC) && \
		protoc -I . \
		--go_out=paths=source_relative:../$(PB_OUT) \
		--go-grpc_out=paths=source_relative:../$(PB_OUT) \
		user/*.proto

.PHONY: proto
proto: proto-user

.PHONY: install-proto-tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
