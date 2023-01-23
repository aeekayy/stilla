#!/bin/bash

info() {
    echo "[info] " $@
}

error() {
    echo "[err] " $@
    exit 1
}

docker_run() {
    info "Starting $1 on port $2"
    local container_name="stilla-${1//[:\^\&\*\(\)\#@!\/]/-}"
    if [ -f ${container_name}.pid ]; then
        error "The container $container_name exists. Run scripts/test-cleanup.sh. Exiting."
    fi
    touch ${container_name}.pid
    docker pull $1
    docker run -p $2:$2 -ti --detach -e POSTGRES_PASSWORD=postgres --name $container_name $1
}

info "Setting up the dependencies for tests"
docker_run postgres:15 5432
docker_run redis:7 6379
docker_run bitnami/kafka:3 9093
docker_run mongo:6 27017