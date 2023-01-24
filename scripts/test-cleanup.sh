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

info "Remove stilla-test Docker network"
docker network rm stilla-test
info "Cleaning up Docker images"
readarray images < <(yqa e -o=j -I=0 '.images[]' scripts/settings.yml )
for images in "${images[@]}"; do
    image=$(echo "$images" | jq -r '.image' -)
    info "Removing $image"
    cleanup_container $image
done