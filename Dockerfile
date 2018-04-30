# Crypt Server Dockerfile
FROM ubuntu:16.04
MAINTAINER Graham Gilbert <graham@grahamgilbert.com>

# Basic env vars for apt and Passenger
ENV HOME /root
ENV DEBIAN_FRONTEND noninteractive
ENV TZ America/New_York
ENV APP_DIR /home/docker/crypt
ENV DOCKER_SAL_DEBUG false
# DOCKER_CRYPT_* env vars configure the following possible settings in Crypt's
# settings.py:
# DOCKER_CRYPT_ADMINS = A list of lists (tuples) of names and email addresses of authorized admins
# DOCKER_CRYPT_ALLOWED = A list of allowed hosts or IP addresses, defaults to *
# DOCKER_CRYPT_LANG = Preferred language
# DOCKER_CRYPT_TZ = Time zone to use for GUI and logging

# To define multiple admins: Some Name,some@where.com:Another One,another@host.net
ENV DOCKER_CRYPT_ADMINS Admin User,admin@test.com
# ENV DOCKER_CRYPT_ALLOWED myhost,1.2.3.4,anotherhost.fqdn.com
ENV DOCKER_CRYPT_LANG en_US
ENV DOCKER_CRYPT_TZ America/New_York

RUN apt-get -y update && \
    apt-get install -y software-properties-common && \
    apt-get install -y python-software-properties && \
    apt-get -y update && \
    add-apt-repository -y ppa:jonathonf/python-3.6 && \
    add-apt-repository -y ppa:nginx/stable && \
    apt-get -y update && \
    apt-get -y install \
    git \
    nginx \
    postgresql \
    postgresql-contrib \
    libpq-dev \
    python3.6-dev \
    python3-pip \
    supervisor \
    libffi-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ADD / $APP_DIR

RUN pip3 install -r $APP_DIR/setup/requirements.txt && \
    pip3 install psycopg2==2.5.3 && \
    pip3 install gunicorn && \
    pip3 install setproctitle

ADD docker/nginx/nginx-env.conf /etc/nginx/main.d/
ADD docker/nginx/crypt.conf /etc/nginx/sites-enabled/crypt.conf
ADD docker/nginx/nginx.conf /etc/nginx/nginx.conf
ADD docker/settings.py $APP_DIR/fvserver/
ADD docker/settings_import.py $APP_DIR/fvserver/
ADD docker/wsgi.py $APP_DIR/
ADD docker/gunicorn_config.py $APP_DIR/
ADD docker/django/management/ $APP_DIR/server/management/
ADD docker/run.sh /run.sh
ADD docker/supervisord.conf $APP_DIR/supervisord.conf

RUN update-rc.d -f postgresql remove && \
    update-rc.d -f nginx remove && \
    rm -f /etc/nginx/sites-enabled/default && \
    mkdir -p /home/app && \
    mkdir -p /home/backup && \
    ln -s $APP_DIR /home/app/crypt

EXPOSE 8000

VOLUME $APP_DIR/keyset

CMD ["/run.sh"]
