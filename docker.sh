#!/bin/bash
CWD=`pwd`


# Clean up

docker ps -aq | xargs docker rm -f

KEYSET=$CWD/keyset
docker build -t "macadmins/crypt-server" $CWD
docker run -d \
  --name="crypt" \
  -e ADMIN_PASS='password' \
  -v $KEYSET:/home/app/crypt/keyset \
  --restart="always" \
  -p 8000:8000 \
  macadmins/crypt-server
