#!/bin/bash

info() {
    echo "[info] $@"
}

info "Seeding the database"
psql "postgresql://postgres:postgres@postgres:5432/postgres" -c "INSERT INTO api_keys(id, name, role) VALUES('9923d21c-dbac-421d-a31a-649a849d4c85','master', 'e3f01984-8185-4829-affe-56b84a9913eb')"
info "Creting the configuration"
curl -H "Content-Type: application/json" -H "Authorization: Bearer 9923d21c-dbac-421d-a31a-649a849d4c85" http://localhost:8080/api/v1/config/ -X POST --data '{ "config_name": "kubernetes", "owner": "testuser", "config": { "enabled": true } }'
info "Seeding complete"