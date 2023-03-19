#!/bin/bash

info() {
    echo "[info] $@"
}

info "Seeding the database"
psql "postgresql://postgres:postgres@postgres:5432/stilla" -c "INSERT INTO api_keys(id, name, role, token) VALUES('9923d21c-dbac-421d-a31a-649a849d4c85','master', 'e3f01984-8185-4829-affe-56b84a9913eb', 'cfacd739-4a13-47ae-82c3-13d6d7ffeb2e')"
info "Creting the configuration"
curl -H "Content-Type: application/json" -H "Authorization: Bearer cfacd739-4a13-47ae-82c3-13d6d7ffeb2e" -H "HostID: 9923d21c-dbac-421d-a31a-649a849d4c85" http://localhost:8080/api/v1/config/ -X POST --data '{ "config_name": "kubernetes", "owner": "testuser", "config": { "enabled": true } }'
info "Seeding complete"