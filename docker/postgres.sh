#!/bin/bash

docker rm -f postgres-crypt

docker run -d --name="postgres-crypt" \
    -e DB_NAME=crypt \
    -e DB_USER=admin \
    -e DB_PASS=password \
    -v /Users/Shared/test-pg-db:/var/lib/postgresql/data \
    -p 5432:5432 \
    grahamgilbert/postgres:9.4.5

sleep 30
