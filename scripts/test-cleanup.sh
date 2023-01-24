#!/bin/bash

info() {
    echo "[info] " $@
}

yqa() {
    if ! [ -x "$(command -v yq)" ]; then
        /snap/bin/yq $@
    else
        yq $@
    fi
}

cleanup_container() {
    local container_name="stilla-${1//[:\^\&\*\(\)\#@!\/]/-}"
    docker rm -f $container_name
    rm -f $container_name.pid
}

readarray images < <(yqa e -o=j -I=0 '.images[]' scripts/settings.yml )
for images in "${images[@]}"; do
    image=$(echo "$images" | jq -r '.image' -)
    info "Removing $image"
    cleanup_container $image
done
#info "Removing postgres"
#cleanup_container postgres:15
#info "Removing redis"
#cleanup_container redis:7
#info "Removing kafka"
#cleanup_container bitnami/kafka:3
#info "Removing mongo"
#cleanup_container mongo:6