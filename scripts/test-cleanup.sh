#!/bin/bash

info() {
    echo "[info] " $@
}

cleanup_container() {
    local container_name="stilla-${1//[:\^\&\*\(\)\#@!\/]/-}"
    docker rm -f $container_name
    rm -f $container_name.pid
}

info "Removing postgres"
cleanup_container postgres:15
info "Removing redis"
cleanup_container redis:7
info "Removing kafka"
cleanup_container bitnami/kafka:3
info "Removing mongo"
cleanup_container mongo:6