import keyczar
from keyczar import keyczart
import os
directory = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'keyset')

if not os.path.exists(directory):
    os.makedirs(directory)

if not os.listdir(directory):
    keyczart.main(['create','--location=keyset','--purpose=crypt','--name=crypt'])
    keyczart.main(['addkey','--location=keyset' ,'--status=primary'])
else:
    print 'Keyset directory already has something in there. Skipping key generation.'
