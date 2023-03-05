name := "stilla"
api_spec := "api/openapi.yaml"
api_path := "./pkg/api"
svc_db := trim(`psql "postgresql://postgres:postgres@${POSTGRES_HOST:-localhost}:5432/postgres" -c "select exists(SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower('stilla'));" -t`)
protoc_ver := "22.0"
protoc_zip := "protoc-" + protoc_ver + "-linux-x86_64.zip"


set shell := ["bash", "-uc"]

build:
	# Solve the buildvcs flag issue later
	go build -ldflags="-X 'main.Version=v0.0.1' -X 'main.BuildTime=$(date)' -X 'main.CommitHash=$(git rev-parse HEAD)'" -buildvcs=false -o {{name}}

gen-api:
	java -jar ~/openapi-generator-cli.jar generate -i {{api_spec}} -g go-gin-server -o {{api_path}} --skip-validate-spec

migrate:
	[[ "{{svc_db}}" == "t" ]] || PGPASSWORD=postgres createdb -h ${POSTGRES_HOST:-localhost} -U postgres stilla
	psql "postgresql://postgres:postgres@${POSTGRES_HOST:-localhost}:5432/stilla" -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	atlas schema apply --url "postgres://postgres:postgres@${POSTGRES_HOST:-localhost}:5432/stilla?sslmode=disable" --to "file://sql/schema.hcl" --auto-approve

setup:
	curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v{{protoc_ver}}/{{protoc_zip}}
	unzip -o {{protoc_zip}} -d /usr/local bin/protoc
	unzip -o {{protoc_zip}} -d /usr/local 'include/*'
	rm -f {{protoc_zip}}
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2