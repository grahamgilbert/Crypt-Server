#!/bin/bash

CWD=`pwd`
docker rm -f crypt
docker build -t macadmins/crypt --no-cache .
docker run -d \
    -e ADMIN_PASS=pass \
    -e DEBUG=false \
    -e DB_NAME=crypt \
    -e DB_USER=admin \
    -e DB_PASS=password \
    --name=crypt \
    --link postgres-crypt:db \
    --restart="always" \
    -v "$CWD/keyset":/home/docker/crypt/keyset \
    -e FIELD_ENCRYPTION_KEY=jKAv1Sde8m6jCYFnmps0iXkUfAilweNVjbvoebBrDwg= \
    -p 8000-8050:8000-8050 \
    macadmins/crypt
