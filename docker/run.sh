#!/bin/bash

cd $APP_DIR
ADMIN_PASS=${ADMIN_PASS:-}
python generate_keyczart.py
# python manage.py syncdb --noinput
# python manage.py migrate server
python manage.py migrate --noinput
python manage.py collectstatic --noinput

if [ ! -z "$ADMIN_PASS" ] ; then
  python manage.py update_admin_user --username=admin --password=$ADMIN_PASS
else
  python manage.py update_admin_user --username=admin --password=password
fi

chown -R www-data:www-data $APP_DIR
chmod go+x $APP_DIR
mkdir -p /var/log/gunicorn
export PYTHONPATH=$PYTHONPATH:$APP_DIR
export DJANGO_SETTINGS_MODULE='fvserver.settings'

if [ "${DOCKER_CRYPT_DEBUG}" = "true" ] || [ "${DOCKER_CRYPT_DEBUG}" = "True" ] || [ "${DOCKER_CRYPT_DEBUG}" = "TRUE" ] ; then
    service nginx stop
    echo "RUNNING IN DEBUG MODE"
    python manage.py runserver 0.0.0.0:8000
else
    supervisord --nodaemon -c $APP_DIR/supervisord.conf
fi
