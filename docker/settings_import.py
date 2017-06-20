#!/usr/bin/python
from os import getenv
import locale

# Read the DEBUG setting from env var
try:
    if getenv('DOCKER_CRYPT_DEBUG').lower() == 'true':
        DEBUG = True
    else:
        DEBUG = False
except:
    DEBUG = False

try:
    if getenv('DOCKER_CRYPT_APPROVE_OWN').lower() == 'false':
        APPROVE_OWN = False
    else:
        APPROVE_OWN = True
except:
    APPROVE_OWN = True

try:
    if getenv('DOCKER_CRYPT_ALL_APPROVE').lower() == 'true':
        ALL_APPROVE = True
    else:
        ALL_APPROVE = False
except:
    ALL_APPROVE = False

# Read list of admins from $DOCKER_CRYPT_ADMINS env var
admin_list = []
if getenv('DOCKER_CRYPT_ADMINS'):
    admins_var = getenv('DOCKER_CRYPT_ADMINS')
    if ',' in admins_var and ':' in admins_var:
        for admin in admins_var.split(':'):
            admin_list.append(tuple(admin.split(',')))
        ADMINS = tuple(admin_list)
    elif ',' in admins_var:
        admin_list.append(tuple(admins_var.split(',')))
        ADMINS = tuple(admin_list)
else:
    ADMINS = (
                ('Admin User', 'admin@test.com')
             )

# Read the preferred time zone from $DOCKER_CRYPT_TZ, use system locale or
# set to 'America/New_York' if neither are set
if getenv('DOCKER_CRYPT_TZ'):
    if '/' in getenv('DOCKER_CRYPT_TZ'):
        TIME_ZONE = getenv('DOCKER_CRYPT_TZ')
    else: TIME_ZONE = 'America/New_York'
elif getenv('TZ'):
    TIME_ZONE = getenv('TZ')
else:
    TIME_ZONE = 'America/New_York'

# Read the preferred language code from $DOCKER_CRYPT_LANG, use system locale or
# set to 'en_US' if neither are set
if getenv('DOCKER_CRYPT_LANG'):
    if '_' in getenv('DOCKER_CRYPT_LANG'):
        LANGUAGE_CODE = getenv('DOCKER_CRYPT_LANG')
    else:
        LANGUAGE_CODE = 'en_US'
elif locale.getdefaultlocale():
    LANGUAGE_CODE = locale.getdefaultlocale()[0]
else:
    LANGUAGE_CODE = 'en_US'

# Set the display name from the $DOCKER_CRYPT_DISPLAY_NAME env var, or
# use the default
if getenv('DOCKER_CRYPT_DISPLAY_NAME'):
    DISPLAY_NAME = getenv('DOCKER_CRYPT_DISPLAY_NAME')
else:
    DISPLAY_NAME = 'Crypt'    

if getenv('DOCKER_CRYPT_EMAIL_HOST'):
    EMAIL_HOST = getenv('DOCKER_CRYPT_EMAIL_HOST')

if getenv('DOCKER_CRYPT_EMAIL_PORT'):
    EMAIL_PORT = getenv('DOCKER_CRYPT_EMAIL_PORT')

if getenv('DOCKER_CRYPT_EMAIL_USER'):
    EMAIL_USER = getenv('DOCKER_CRYPT_EMAIL_USER')

if getenv('DOCKER_CRYPT_EMAIL_PASSWORD'):
    EMAIL_PASSWORD = getenv('DOCKER_CRYPT_EMAIL_PASSWORD')

if getenv('DOCKER_CRYPT_HOST_NAME'):
    HOST_NAME = getenv('DOCKER_CRYPT_HOST_NAME')

# Read the list of allowed hosts from the $DOCKER_CRYPT_ALLOWED env var, or
# allow all hosts if none was set.
if getenv('DOCKER_CRYPT_ALLOWED'):
    ALLOWED_HOSTS = getenv('DOCKER_CRYPT_ALLOWED').split(',')
else:
    ALLOWED_HOSTS = ['*']
