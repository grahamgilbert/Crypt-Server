#!/bin/sh

set -e

cd $APP_DIR
ADMIN_PASS=${ADMIN_PASS:-}
# python3 generate_keyczart.py
python3 manage.py migrate --noinput

if [ ! -z "$ADMIN_PASS" ] ; then
  python3 manage.py update_admin_user --username=admin --password=$ADMIN_PASS
else
  python3 manage.py update_admin_user --username=admin --password=password
fi


export PYTHONPATH=$PYTHONPATH:$APP_DIR
export DJANGO_SETTINGS_MODULE='fvserver.settings'

if [ "${DOCKER_CRYPT_DEBUG}" = "true" ] || [ "${DOCKER_CRYPT_DEBUG}" = "True" ] || [ "${DOCKER_CRYPT_DEBUG}" = "TRUE" ] ; then
    echo "RUNNING IN DEBUG MODE"
    python3 manage.py runserver 0.0.0.0:8000
else
    gunicorn -c $APP_DIR/gunicorn_config.py fvserver.wsgi:application
fi
