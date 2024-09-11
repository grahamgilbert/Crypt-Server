import os

# Django settings for fvserver project.

PROJECT_DIR = os.path.abspath(
	os.path.join(os.path.dirname(os.path.abspath(__file__)), os.path.pardir)
)
ENCRYPTED_FIELD_KEYS_DIR = os.path.join(PROJECT_DIR, "keyset")
DEBUG = False

ROTATE_VIEWED_SECRETS = True

DATE_FORMAT = "Y-m-d H:i:s"
DATETIME_FORMAT = "Y-m-d H:i:s"

ADMINS = ()

FIELD_ENCRYPTION_KEY = os.environ.get("FIELD_ENCRYPTION_KEY", "")

MANAGERS = ADMINS

# Local time zone for this installation. Choices can be found here:
# http://en.wikipedia.org/wiki/List_of_tz_zones_by_name
# although not all choices may be available on all operating systems.
# In a Windows environment this must be set to your system time zone.
TIME_ZONE = "Europe/London"

# Language code for this installation. All choices can be found here:
# http://www.i18nguy.com/unicode/language-identifiers.html
LANGUAGE_CODE = "en-us"

SITE_ID = 1

# If you set this to False, Django will make some optimizations so as not
# to load the internationalization machinery.
USE_I18N = True

# If you set this to False, Django will not format dates, numbers and
# calendars according to the current locale.
USE_L10N = False

# If you set this to False, Django will not use timezone-aware datetimes.
USE_TZ = True

# Absolute filesystem path to the directory that will hold user-uploaded files.
# Example: "/home/media/media.lawrence.com/media/"
MEDIA_ROOT = ""

# URL that handles the media served from MEDIA_ROOT. Make sure to use a
# trailing slash.
# Examples: "http://media.lawrence.com/media/", "http://example.com/media/"
MEDIA_URL = ""

# Absolute path to the directory static files should be collected to.
# Don't put anything in this directory yourself; store your static files
# in apps' "static/" subdirectories and in STATICFILES_DIRS.
# Example: "/home/media/media.lawrence.com/static/"
STATIC_ROOT = os.path.join(PROJECT_DIR, "static")

# URL prefix for static files.
# Example: "http://media.lawrence.com/static/"
STATIC_URL = "/static/"

# URL prefix for admin static files -- CSS, JavaScript and images.
# Make sure to use a trailing slash.
# Examples: "http://foo.com/static/admin/", "/static/admin/".
# deprecated in Django 1.4, but django_wsgiserver still looks for it
# when serving admin media
ADMIN_MEDIA_PREFIX = "/static_admin/"

# Additional locations of static files
STATICFILES_DIRS = (
	# Put strings here, like "/home/html/static" or "C:/www/django/static".
	# Always use forward slashes, even on Windows.
	# Don't forget to use absolute paths, not relative paths.
	os.path.join(PROJECT_DIR, "site_static"),
)

LOGIN_URL = "/login/"
LOGIN_REDIRECT_URL = "/"

ALLOWED_HOSTS = ["*"]

# List of finder classes that know how to find static files in
# various locations.
STATICFILES_FINDERS = (
	"django.contrib.staticfiles.finders.FileSystemFinder",
	"django.contrib.staticfiles.finders.AppDirectoriesFinder",
	#    'django.contrib.staticfiles.finders.DefaultStorageFinder',
)

# Make this unique, and don't share it with anybody.
SECRET_KEY = "6%y8=x5(#ufxd*+d+-ohwy0b$5z^cla@7tvl@n55_h_cex0qat"

TEMPLATES = [
	{
		"BACKEND": "django.template.backends.django.DjangoTemplates",
		"DIRS": [
			# insert your TEMPLATE_DIRS here
			os.path.join(PROJECT_DIR, "templates")
		],
		"APP_DIRS": True,
		"OPTIONS": {
			"context_processors": [
				# Insert your TEMPLATE_CONTEXT_PROCESSORS here or use this
				# list if you haven't customized them:
				"django.contrib.auth.context_processors.auth",
				"django.template.context_processors.debug",
				"django.template.context_processors.i18n",
				"django.contrib.messages.context_processors.messages",
				"django.template.context_processors.media",
				"django.template.context_processors.static",
				"django.template.context_processors.tz",
				"django.template.context_processors.request",
				"fvserver.context_processors.crypt_version",
				"fvserver.context_processors.oidc_enabled"
			],
			"debug": DEBUG,
		},
	}
]

MIDDLEWARE = [
	"django.middleware.security.SecurityMiddleware",
	"whitenoise.middleware.WhiteNoiseMiddleware",
	"django.contrib.sessions.middleware.SessionMiddleware",
	"django.middleware.common.CommonMiddleware",
	"django.middleware.csrf.CsrfViewMiddleware",
	"django.contrib.auth.middleware.AuthenticationMiddleware",
	"django.contrib.messages.middleware.MessageMiddleware",
	"django.middleware.clickjacking.XFrameOptionsMiddleware",
]

ROOT_URLCONF = "fvserver.urls"

# Python dotted path to the WSGI application used by Django's runserver.
WSGI_APPLICATION = "fvserver.wsgi.application"

INSTALLED_APPS = [
	"whitenoise.runserver_nostatic",
	"django.contrib.auth",
	"django.contrib.contenttypes",
	"django.contrib.sessions",
	"django.contrib.sites",
	"django.contrib.messages",
	"django.contrib.staticfiles",
	# Uncomment the next line to enable the admin:
	"django.contrib.admin",
	# Uncomment the next line to enable admin documentation:
	"django.contrib.admindocs",
	"server",
	"bootstrap4",
	"django_extensions",
]

LOGGING = {
	"version": 1,
	"disable_existing_loggers": False,
	"formatters": {
		"default": {
			"format": "[DJANGO] %(levelname)s %(asctime)s %(module)s "
					  "%(name)s.%(funcName)s:%(lineno)s: %(message)s"
		},
	},
	"handlers": {
		"console": {
			"level": "DEBUG",
			"class": "logging.StreamHandler",
			"formatter": "default",
		}
	},
	"loggers": {
		"*": {
			"handlers": ["console"],
			"level": "DEBUG",
			"propagate": True,
		},
	},
}

DEFAULT_AUTO_FIELD = "django.db.models.AutoField"

# OIDC stuff
AUTHENTICATION_BACKENDS = [
	"django.contrib.auth.backends.ModelBackend"
]

OIDC_ENABLED = int(os.environ.get("OIDC_ENABLED", default=0))

if OIDC_ENABLED:
	OIDC_VERIFY_SSL = True
	OIDC_CREATE_USER = True
	OIDC_RENEW_ID_TOKEN_EXPIRY_SECONDS = int(os.environ.get('OIDC_RENEW_ID_TOKEN_EXPIRY_SECONDS', default=900))
	# OIDC_RP_IDP_SIGN_KEY = "<OP signing key in PEM or DER format>"
	OIDC_RP_CLIENT_ID = os.environ.get("OIDC_RP_CLIENT_ID", None)
	OIDC_RP_CLIENT_SECRET = os.environ.get("OIDC_RP_CLIENT_SECRET", None)
	OIDC_RP_SCOPES = "openid profile email groups"
	OIDC_RP_SIGN_ALGO = os.environ.get("OIDC_RP_SIGN_ALGO", None)
	OIDC_OP_AUTHORIZATION_ENDPOINT = os.environ.get("OIDC_OP_AUTHORIZATION_ENDPOINT", None)
	OIDC_OP_JWKS_ENDPOINT = os.environ.get("OIDC_OP_JWKS_ENDPOINT", None)
	OIDC_OP_TOKEN_ENDPOINT = os.environ.get("OIDC_OP_TOKEN_ENDPOINT", None)
	OIDC_OP_USER_ENDPOINT = os.environ.get("OIDC_OP_USER_ENDPOINT", None)
	OIDC_REDIRECT_URL = os.environ.get("OIDC_REDIRECT_URL", None)

	INSTALLED_APPS.append("mozilla_django_oidc")
	MIDDLEWARE.append("mozilla_django_oidc.middleware.SessionRefresh")

	# overwrite the existing backends to disable local auth
	AUTHENTICATION_BACKENDS = ["fvserver.oidc.CustomOIDC"]

INSTALLED_APPS = tuple(INSTALLED_APPS)
