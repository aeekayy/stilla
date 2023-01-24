#!/bin/bash

info() {
    echo "[info] " $@
}

error() {
    echo "[err] " $@
    exit 1
}

yqa() {
    if ! [ -x "$(command -v yq)" ]; then
        /snap/bin/yq $@
    else
        yq $@
    fi
}

docker_run() {
    info "Starting $1"
    local container_name="stilla-${1//[:\^\&\*\(\)\#@!\/]/-}"
    if [ -f ${container_name}.pid ]; then
        error "The container $container_name exists. Run scripts/test-cleanup.sh. Exiting."
    fi
    docker pull $1
    if [ "$2" != "" ]; then
        ports="-p $2"
    fi
    if [ "$3" != "" ]; then
        env="-e $3"
    fi
    docker run $ports -ti --detach $env --name $container_name $1 && touch ${container_name}.pid
}

info "Setting up the dependencies for tests"
readarray images < <(yqa e -o=j -I=0 '.images[]' scripts/settings.yml )
for images in "${images[@]}"; do
    image=$(echo "$images" | jq -r '.image' -)
    ports=$(echo "$images" | jq -r '.ports | join("-p ")' -)
    env=$(echo "$images" | jq -r 'select(.env != null) | .env | [ .[] | keys_unsorted[] as $key | "\($key)=\(.[$key])" ] | join (" -e ")' -)
    echo "image: $image"
    echo "ports: $ports"
    echo "env: $env"
    docker_run $image $ports $env
done