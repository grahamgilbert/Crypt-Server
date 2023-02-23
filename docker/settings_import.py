#!/usr/bin/env python
from os import getenv
import pytz
import locale


if getenv("DEBUG"):
  if getenv("DEBUG").lower() == "true":
    DEBUG = True
else:
    DEBUG = False

if getenv("APPROVE_OWN"):
  if getenv("APPROVE_OWN").lower() == "false":
    APPROVE_OWN = False
else:
   APPROVE_OWN = True

if getenv("ROTATE_VIEWED_SECRETS"):
  if getenv("ROTATE_VIEWED_SECRETS").lower() == "false":
      ROTATE_VIEWED_SECRETS = False
else:
    ROTATE_VIEWED_SECRETS = True

if getenv("ROTATE_VIEWED_SECRETS_DAYS"):
  if getenv("ALL_APPROVE").lower() == "true":
      ALL_APPROVE = True
else:
    ALL_APPROVE = False

# Read list of admins from $ADMINS env var
admin_list = []
if getenv("ADMINS"):
    admins_var = getenv("ADMINS")
    if "," in admins_var and ":" in admins_var:
        for admin in admins_var.split(":"):
            admin_list.append(tuple(admin.split(",")))
        ADMINS = admin_list
    elif "," in admins_var:
        admin_list.append(tuple(admins_var.split(",")))
        ADMINS = admin_list
else:
    ADMINS = [("Admin User", "admin@test.com")]

# Read the preferred time zone from $TZ, use system locale or
# set to 'America/New_York' if neither are set
if getenv("TZ") and getenv("TZ") in pytz.all_timezones:
    TIME_ZONE = getenv("TZ")
else:
    TIME_ZONE = "America/New_York"

# Read the preferred language code from $LANG & default to en-us if not set
# note django does not support locale-format for LANG
if getenv("LANG"):
    LANGUAGE_CODE = getenv("LANG")
else:
    LANGUAGE_CODE = "en-us"

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
    CSRF_TRUSTED_ORIGINS = [getenv("HOST_NAME")]
else:
    HOST_NAME = "https://cryptexample.com"
    CSRF_TRUSTED_ORIGINS = []

if getenv("EMAIL_SENDER"):
    EMAIL_SENDER = getenv("EMAIL_SENDER")
else:
    EMAIL_SENDER = "crypt@cryptexample.com"

# Read the list of allowed hosts from the $DOCKER_CRYPT_ALLOWED env var, or
# allow all hosts if none was set.
if getenv("ALLOWED_HOSTS"):
    ALLOWED_HOSTS = getenv("ALLOWED_HOSTS").split(",")
else:
    ALLOWED_HOSTS = ["*"]

if getenv("SEND_EMAIL") and getenv("SEND_EMAIL").lower() == "true":
    SEND_EMAIL = True
else:
    SEND_EMAIL = False

if getenv("EMAIL_USE_TLS") and getenv("EMAIL_USE_TLS").lower() == "true":
    EMAIL_USE_TLS = True

if getenv("EMAIL_USE_SSL") and getenv("EMAIL_USE_SSL").lower() == "true":
    EMAIL_USE_SSL = True
