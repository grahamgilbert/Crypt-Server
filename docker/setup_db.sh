#!/bin/sh
DB_NAME=crypt
DB_USER=admin
DB_PASS=password

echo "CREATE ROLE $DB_USER WITH LOGIN ENCRYPTED PASSWORD '${DB_PASS}' CREATEDB;" | docker run \
  --rm \
  --interactive \
  --link postgres-crypt:postgres \
  postgres:9.3.4 \
  bash -c 'exec psql -h "$POSTGRES_PORT_5432_TCP_ADDR" -p "$POSTGRES_PORT_5432_TCP_PORT" -U postgres'

echo "CREATE DATABASE $DB_NAME WITH OWNER $DB_USER TEMPLATE template0 ENCODING 'UTF8';" | docker run \
  --rm \
  --interactive \
  --link postgres-crypt:postgres \
  postgres:9.3.4 \
  bash -c 'exec psql -h "$POSTGRES_PORT_5432_TCP_ADDR" -p "$POSTGRES_PORT_5432_TCP_PORT" -U postgres'

echo "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" | docker run \
  --rm \
  --interactive \
  --link postgres-crypt:postgres \
  postgres:9.3.4 \
  bash -c 'exec psql -h "$POSTGRES_PORT_5432_TCP_ADDR" -p "$POSTGRES_PORT_5432_TCP_PORT" -U postgres'


