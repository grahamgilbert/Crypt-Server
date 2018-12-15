#!/bin/bash
CWD=`pwd`


# Check that Docker Machine exists
# --vmwarefusion-boot2docker-url "https://github.com/boot2docker/boot2docker/releases/download/v1.8.2/boot2docker.iso"
# if [ -z "$(docker-machine ls | grep crypt)" ]; then
#   docker-machine create -d vmwarefusion --vmwarefusion-disk-size=500000 --vmwarefusion-memory-size=2048 --vmwarefusion-cpu-count=4 crypt
#   docker-machine env crypt
#   eval "$(docker-machine env crypt)"
# fi
# Check that Docker Machine is running
# if [ "$(docker-machine status crypt)" != "Running" ]; then
#   docker-machine start crypt
#   docker-machine env crypt
#   eval "$(docker-machine env crypt)"
# fi

# Get the IP address of the machine
# IP=`docker-machine ip crypt`

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
