FROM python:3.7-alpine

LABEL maintainer="graham@grahamgilbert.com"

ENV APP_DIR /home/docker/crypt
ENV DOCKER_SAL_DEBUG false
ENV DOCKER_CRYPT_LANG en_US
ENV DOCKER_CRYPT_TZ America/New_York

COPY setup/requirements.txt /tmp/requirements.txt

RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
            gcc \
            git \
            libffi-dev \
            libc-dev \
            musl-dev \
            linux-headers \
            pcre-dev \
            postgresql-dev \
    && LIBRARY_PATH=/lib:/usr/lib /bin/sh -c "pip install --no-cache-dir -r /tmp/requirements.txt" \
    && rm /tmp/requirements.txt

    # && runDeps="$( \
    #         scanelf --needed --nobanner --recursive /venv \
    #                 | awk '{ gsub(/,/, "\nso:", $2); print "so:" $2 }' \
    #                 | sort -u \
    #                 | xargs -r apk info --installed \
    #                 | sort -u \
    # )" \
    # && apk add --virtual .python-rundeps $runDeps \
    # && apk del .build-deps \

COPY / $APP_DIR
COPY docker/settings.py $APP_DIR/fvserver/
COPY docker/settings_import.py $APP_DIR/fvserver/
# COPY docker/wsgi.py $APP_DIR
COPY docker/gunicorn_config.py $APP_DIR/
COPY docker/django/management/ $APP_DIR/server/management/
COPY docker/run.sh /run.sh

RUN chmod +x /run.sh \
    && mkdir -p /home/app \
    && ln -s ${APP_DIR} /home/app/crypt

WORKDIR ${APP_DIR}
RUN export FIELD_ENCRYPTION_KEY="jKAv1Sde8m6jCYFnmps0iXkUfAilweNVjbvoebBrDwg="; python manage.py collectstatic --noinput; export FIELD_ENCRYPTION_KEY=""

EXPOSE 8000

VOLUME $APP_DIR/keyset

CMD ["/run.sh"]