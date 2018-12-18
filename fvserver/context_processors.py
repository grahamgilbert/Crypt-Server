import plistlib
import os


def crypt_version(request):
    # return the value you want as a dictionary. you may add multiple values in there.
    current_dir = os.path.dirname(os.path.realpath(__file__))
    version = plistlib.readPlist(os.path.join(current_dir, "version.plist"))
    return {"CRYPT_VERSION": version["version"]}
