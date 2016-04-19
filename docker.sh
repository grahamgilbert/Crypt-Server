#!/bin/bash
CWD=`pwd`
# if [ ! -d "$CWD/Data" ] ; then
#     echo "Script must be run from root of imagr directory"
#     exit 1
# fi


# Check that Docker Machine exists

if [ -z "$(docker-machine ls | grep imagr)" ]; then
  docker-machine create -d vmwarefusion --vmwarefusion-disk-size=500000 --vmwarefusion-boot2docker-url "https://github.com/boot2docker/boot2docker/releases/download/v1.8.2/boot2docker.iso" --vmwarefusion-memory-size=2048 --vmwarefusion-cpu-count=4 imagr
  docker-machine env imagr
  eval "$(docker-machine env imagr)"
fi
# Check that Docker Machine is running
if [ "$(docker-machine status imagr)" != "Running" ]; then
  docker-machine start imagr
  docker-machine env imagr
  eval "$(docker-machine env imagr)"
fi

# Get the IP address of the machine
IP=`docker-machine ip imagr`

# Clean up

docker ps -aq | xargs docker rm -f
DATA_DIR=$CWD/Data
docker build -t "xbstech/puppet-certmgr" /Users/grahamgilbert/src/Work/docker-puppet-certmgr
docker run -d \
  --name="certmgr" \
  -e PUPPETCERTMGR_USERNAME='username' \
  -e PUPPETCERTMGR_PASSWORD='password' \
  -e PUPPETCERTMGR_SALURL='https://sal.xbstech.com' \
  -e PUPPETCERTMGR_SALPUBLIC='1nes2xbb89zgmx0gnj66puk8' \
  -e PUPPETCERTMGR_SALPRIVATE='ig0jaak80xmvk4fra3g3xll8uwvkvl8tjle9qazd4ayn40y5rn2aghkq3jrkkhq5' \
  --restart="always" \
  -p 5000:5000 \
  xbstech/puppet-certmgr

docker run -d \
-v "$DATA_DIR/web_root":/usr/local/nginx/html \
-v "$DATA_DIR/Docker/sites-templates":/etc/nginx/sites-templates \
-v /Users/grahamgilbert/src/Munki/work_munki:/munki \
--name="web" \
  --restart="always" \
  -p 80:80 \
  grahamgilbert/proxy


docker run -d \
  -p 0.0.0.0:69:69/udp \
  -v "$DATA_DIR/web_root":/nbi \
  --name tftpd \
  --restart="always" \
  macadmins/tftpd


docker run -d \
-p 8000:8000 \
--name imagr_server \
--restart="always" \
grahamgilbert/imagr-server

docker run -d \
    -p 0.0.0.0:67:67/udp \
    -v "$DATA_DIR/web_root":/nbi \
    -e BSDPY_IFACE=eth0 \
    -e BSDPY_NBI_URL=http://$IP \
    -e BSDPY_IP=$IP \
    --name bsdpy \
    --restart=always \
    bruienne/bsdpy:1.0
echo
echo "### Your Docker Machine IP is: $IP"
echo
echo `docker-machine env imagr`
