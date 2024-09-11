from django.conf import settings
import plistlib
import os


def crypt_version(request):
	# return the value you want as a dictionary. you may add multiple values in there.
	current_dir = os.path.dirname(os.path.realpath(__file__))
	with open(
			os.path.join(os.path.dirname(current_dir), "fvserver", "version.plist"), "rb"
	) as f:
		version = plistlib.load(f)
	return {"CRYPT_VERSION": version["version"]}


def oidc_enabled(request):
	return {
		"OIDC_ENABLED": settings.OIDC_ENABLED
	}
