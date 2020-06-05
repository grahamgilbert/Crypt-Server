FROM python:3.7.2-alpine3.7

LABEL maintainer="graham@grahamgilbert.com"

ENV APP_DIR /home/docker/crypt
ENV DEBUG false
ENV LANG en-us
ENV TZ America/New_York
ENV LC_ALL en_US.UTF-8

COPY setup/requirements.txt /tmp/requirements.txt

# This is gross, but needed until we get pip patched in the upstream image
RUN LIBRARY_PATH=/lib:/usr/lib /bin/sh -c "pip install --upgrade pip==19.0.3"

RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
    gcc \
    git \
    libressl-dev \
    libffi-dev \
    libc-dev \
    musl-dev \
    linux-headers \
    pcre-dev \
    postgresql-dev \
    xmlsec-dev \
    tzdata \
    && LIBRARY_PATH=/lib:/usr/lib /bin/sh -c "pip install --no-cache-dir -r /tmp/requirements.txt" \
    && rm /tmp/requirements.txt

COPY / $APP_DIR
COPY docker/settings.py $APP_DIR/fvserver/
COPY docker/settings_import.py $APP_DIR/fvserver/
COPY docker/gunicorn_config.py $APP_DIR/
COPY docker/django/management/ $APP_DIR/server/management/
COPY docker/run.sh /run.sh

RUN chmod +x /run.sh \
    && mkdir -p /home/app \
    && ln -s ${APP_DIR} /home/app/crypt

WORKDIR ${APP_DIR}
# don't use this key anywhere else, this is just for collectstatic to run
RUN export FIELD_ENCRYPTION_KEY="jKAv1Sde8m6jCYFnmps0iXkUfAilweNVjbvoebBrDwg="; python manage.py collectstatic --noinput; export FIELD_ENCRYPTION_KEY=""

EXPOSE 8000

VOLUME $APP_DIR/keyset

CMD ["/run.sh"]
