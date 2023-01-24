current_dir = $(shell pwd)
go_dir = /home/ubuntu/go/src

.PHONY: build
build:
	go mod tidy
	$(MAKE) gen-protobuf
	go build -o stilla

.PHONY: gen-api
gen-api:
	java -jar ~/openapi-generator-cli.jar generate -i api/openapi.yaml -g go-gin-server -o ./pkg/api --skip-validate-spec

.PHONY: install-proto-go
install-proto-go:
	go get google.golang.org/protobuf/cmd/protoc-gen-go
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc

.PHONY: gen-protobuf
gen-protobuf:
ifdef GITHUB_EVENT_NAME
	$(MAKE) install-proto-go
endif
	protoc -I=$(current_dir)/api/protobuf --go_out=$(go_dir) $(current_dir)/api/protobuf/messages.proto
