from fvserver.system_settings import *
from fvserver.settings_import import *
from django.utils.log import DEFAULT_LOGGING
import os


# Django settings for fvserver project.

DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.sqlite3",  # Add 'postgresql_psycopg2', 'mysql', 'sqlite3' or 'oracle'.
        "NAME": os.path.join(
            PROJECT_DIR, "crypt.db"
        ),  # Or path to database file if using sqlite3.
        "USER": "",  # Not used with sqlite3.
        "PASSWORD": "",  # Not used with sqlite3.
        "HOST": "",  # Set to empty string for localhost. Not used with sqlite3.
        "PORT": "",  # Set to empty string for default. Not used with sqlite3.
    }
}

host = None
port = None

if "DB_HOST" in os.environ:
    host = os.environ.get("DB_HOST")
    port = os.environ.get("DB_PORT")

elif "DB_PORT_5432_TCP_ADDR" in os.environ:
    host = os.environ.get("DB_PORT_5432_TCP_ADDR")
    port = os.environ.get("DB_PORT_5432_TCP_PORT", "5432")

if host and port:
    DATABASES = {
        "default": {
            "ENGINE": "django.db.backends.postgresql_psycopg2",
            "NAME": os.environ["DB_NAME"],
            "USER": os.environ["DB_USER"],
            "PASSWORD": os.environ["DB_PASS"],
            "HOST": host,
            "PORT": port,
        }
    }

# Don't filter anything going to console
DEFAULT_LOGGING["handlers"]["console"]["filters"] = []

DEFAULT_LOGGING["loggers"][""] = {
    "handlers": ["console"],
    "level": "INFO",
    "propagate": True,
}
