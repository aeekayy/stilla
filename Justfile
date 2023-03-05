name := "stilla"
api_spec := "api/openapi.yaml"
api_path := "./pkg/api"
svc_db := trim(`psql "postgresql://postgres:postgres@localhost:5432/postgres" -c "select exists(SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower('stilla'));" -t`)

set shell := ["bash", "-uc"]

build:
	go build -o {{name}}

gen-api:
	java -jar ~/openapi-generator-cli.jar generate -i {{api_spec}} -g go-gin-server -o {{api_path}} --skip-validate-spec

migrate:
	[[ "{{svc_db}}" == "t" ]] || PGPASSWORD=postgres createdb -h localhost -U postgres stilla
	psql "postgresql://postgres:postgres@localhost:5432/stilla" -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	atlas schema apply --url "postgres://postgres:postgres@localhost:5432/stilla?sslmode=disable" --to "file://sql/schema.hcl" --auto-approve

setup:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2