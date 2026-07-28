FROM python:3.10.11-alpine3.16

LABEL maintainer="graham@grahamgilbert.com"

ENV APP_DIR /home/docker/crypt
ENV DEBUG false
ENV LANG en
ENV TZ Etc/UTC
ENV LC_ALL en_US.UTF-8



RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
    gcc \
    git \
    openssl-dev \
    build-base \
    libffi-dev \
    libc-dev \
    musl-dev \
    linux-headers \
    pcre-dev \
    postgresql-dev \
    xmlsec-dev \
    tzdata \
    postgresql-libs \
    libpq

COPY setup/requirements.txt /tmp/requirements.txt

RUN set -ex \
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
# collectstatic imports the app, which needs a key present to load the encrypted
# fields. Generate a throwaway one for this build step only, so no key literal
# ends up in the image, the layer, or `docker history`.
RUN FIELD_ENCRYPTION_KEY="$(python -c 'import base64, os; print(base64.urlsafe_b64encode(os.urandom(32)).decode())')" \
    python manage.py collectstatic --noinput

EXPOSE 8000

VOLUME $APP_DIR/keyset

CMD ["/run.sh"]
