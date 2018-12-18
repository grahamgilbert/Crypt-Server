import keyczar
import subprocess
import os

directory = os.path.join(os.path.dirname(os.path.abspath(__file__)), "keyset")

if not os.path.exists(directory):
    os.makedirs(directory)

if not os.listdir(directory):
    location_string = "--location={}".format(directory)
    cmd = ["keyczart", "create", location_string, "--purpose=crypt", "--name=crypt"]
    subprocess.check_call(cmd)
    cmd = ["keyczart", "addkey", location_string, "--status=primary"]
    subprocess.check_call(cmd)
else:
    print("Keyset directory already has something in there. Skipping key generation.")
