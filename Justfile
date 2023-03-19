name := "stilla"
cwd := `pwd`

# Service variables
# Language: Go 1.20
api_spec := "service/api/openapi.yaml"
api_path := "./service/pkg/api"
svc_db := trim(`psql "postgresql://postgres:postgres@${POSTGRES_HOST:-localhost}:5432/postgres" -c "select exists(SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower('stilla'));" -t`)
protoc_ver := "22.0"
protoc_zip := "protoc-" + protoc_ver + "-linux-x86_64.zip"

# SDK variables 
# Language: Python
virtualenvsh := "~/.local/bin/virtualenvwrapper.sh"
python_venv := "stilla-client"
python_dir := "~/.virtualenvs/" + python_venv
system_python := if os_family() == "windows" { "py.exe -3.8" } else { "python3" }


# This is currently failing for go 1.20
bazel:
	set shell := ["bash", "-uc"]
	bazel run //:gazelle
	bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies

build: build-go 

build-go:
	set shell := ["bash", "-uc"]
	# Solve the buildvcs flag issue later
	go build -ldflags="-X 'main.Version=v0.0.1' -X 'main.BuildTime=$(date)' -X 'main.CommitHash=$(git rev-parse HEAD)'" -buildvcs=false -o {{name}} ./service

build-python: setup-python
	#!/bin/bash
	cd sdk/python
	{{ python_dir }}/bin/python -m build 

# Generates Go files from openapi specification
gen-api:
	java -jar ~/openapi-generator-cli.jar generate -i {{api_spec}} -g go-gin-server -o {{api_path}} --skip-validate-spec

migrate:
	[[ "{{svc_db}}" == "t" ]] || PGPASSWORD=postgres createdb -h ${POSTGRES_HOST:-localhost} -U postgres stilla
	psql "postgresql://postgres:postgres@${POSTGRES_HOST:-localhost}:5432/stilla" -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	atlas schema apply --url "postgres://postgres:postgres@${POSTGRES_HOST:-localhost}:5432/stilla?sslmode=disable" --to "file://service/sql/schema.hcl" --auto-approve

# setup dependencies for Github Actions
setup:
	curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v{{protoc_ver}}/{{protoc_zip}}
	unzip -o {{protoc_zip}} -d /usr/local bin/protoc
	unzip -o {{protoc_zip}} -d /usr/local 'include/*'
	rm -f {{protoc_zip}}
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	# install atlas
	curl -sSf https://atlasgo.sh | sh
	apt install -y python3 python3-pip
	pip3 install locust

# Setup Python virtual environment
# {{ python_dir }}/bin/pip3 install sdk/python/requirements.txt
setup-python:
	#!/bin/bash
	export VIRTUALENVWRAPPER_PYTHON=/usr/bin/python3
	export WORKON_HOME=./.virtualenvs
	source $HOME/.local/bin/virtualenvwrapper.sh
	cd sdk/python
	echo -e "python: ${VIRTUALENVWRAPPER_PYTHON}\nworkon_home=${WORKON_HOME}"
	if test ! -e {{ python_dir }}; then mkvirtualenv -p {{ system_python }} -a {{ cwd }}/sdk/python {{ python_venv }} && echo "Created stilla-client"; fi
	source {{ python_dir }}/bin/activate
	{{ python_dir }}/bin/pip3 install -r requirements.txt
	
# linting
lint: lint-python lint-go 

# Lint all Go files
lint-go:
	golint ./...

lint-python:
	black ./...

fmt:
	go fmt ./...

unit-test: unit-test-go unit-test-python

unit-test-go:
	go test -v ./...

unit-test-python:
	cd sdk/python && pytest

test: unit-test

performance:
	cd {{cwd}}/service/lib/db && go test -bench=.
	cd {{cwd}}

load-test:
	locust -f tests/locustfile.py --headless --skip-log -u 100 -r 3 --host http://localhost:8080 -t 300s -L ERROR

# Run the Stilla service
run: build
	cp service/stilla.gh.yaml stilla.yaml
	./stilla &

seed:
	service/scripts/seed.sh

prepare-commit: lint fmt unit-test performance