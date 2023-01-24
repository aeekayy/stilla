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

.PHONY: gen-protobuf
gen-protobuf:
	protoc -I=$(current_dir)/api/protobuf --go_out=$(go_dir) $(current_dir)/api/protobuf/messages.proto
