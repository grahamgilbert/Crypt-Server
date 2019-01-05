#!/usr/bin/python
from os import getenv
import locale

# Read the DEBUG setting from env var
try:
    if getenv("DEBUG").lower() == "true":
        DEBUG = True
    else:
        DEBUG = False
except:
    DEBUG = False

try:
    if getenv("APPROVE_OWN").lower() == "false":
        APPROVE_OWN = False
    else:
        APPROVE_OWN = True
except:
    APPROVE_OWN = True

try:
    if getenv("ROTATE_VIEWED_SECRETS").lower() == "false":
        ROTATE_VIEWED_SECRETS = False
    else:
        ROTATE_VIEWED_SECRETS = True
except:
    ROTATE_VIEWED_SECRETS = True

try:
    if getenv("ALL_APPROVE").lower() == "true":
        ALL_APPROVE = True
    else:
        ALL_APPROVE = False
except:
    ALL_APPROVE = False

# Read list of admins from $ADMINS env var
admin_list = []
if getenv("ADMINS"):
    admins_var = getenv("ADMINS")
    if "," in admins_var and ":" in admins_var:
        for admin in admins_var.split(":"):
            admin_list.append(tuple(admin.split(",")))
        ADMINS = tuple(admin_list)
    elif "," in admins_var:
        admin_list.append(tuple(admins_var.split(",")))
        ADMINS = tuple(admin_list)
else:
    ADMINS = ("Admin User", "admin@test.com")

# Read the preferred time zone from $TZ, use system locale or
# set to 'America/New_York' if neither are set
if getenv("TZ"):
    if "/" in getenv("TZ"):
        TIME_ZONE = getenv("TZ")
    else:
        TIME_ZONE = "America/New_York"
elif getenv("TZ"):
    TIME_ZONE = getenv("TZ")
else:
    TIME_ZONE = "America/New_York"

# Read the preferred language code from $LANG, use system locale or
# set to 'en_US' if neither are set
if getenv("LANG"):
    if "_" in getenv("LANG"):
        LANGUAGE_CODE = getenv("LANG")
    else:
        LANGUAGE_CODE = "en_US"
elif locale.getdefaultlocale():
    LANGUAGE_CODE = locale.getdefaultlocale()[0]
else:
    LANGUAGE_CODE = "en_US"

# Set the display name from the $DISPLAY_NAME env var, or
# use the default
if getenv("DISPLAY_NAME"):
    DISPLAY_NAME = getenv("DISPLAY_NAME")
else:
    DISPLAY_NAME = "Crypt"

if getenv("EMAIL_HOST"):
    EMAIL_HOST = getenv("EMAIL_HOST")

if getenv("EMAIL_PORT"):
    EMAIL_PORT = getenv("EMAIL_PORT")

if getenv("EMAIL_USER"):
    EMAIL_USER = getenv("EMAIL_USER")

if getenv("EMAIL_PASSWORD"):
    EMAIL_PASSWORD = getenv("EMAIL_PASSWORD")

if getenv("HOST_NAME"):
    HOST_NAME = getenv("HOST_NAME")

# Read the list of allowed hosts from the $DOCKER_CRYPT_ALLOWED env var, or
# allow all hosts if none was set.
if getenv("ALLOWED_HOSTS"):
    ALLOWED_HOSTS = getenv("ALLOWED_HOSTS").split(",")
else:
    ALLOWED_HOSTS = ["*"]
