Upgrading on Ubuntu 12.04 LTS
=====================
This document assumes Ubuntu 12.04 LTS and that you have an existing installation of Crypt-Server, installed using the [instructions provided](https://github.com/grahamgilbert/Crypt-Server/blob/master/docs/Installation_on_Ubuntu_12.md). If you don't have an existing installation, you just need to follow the installation instructions.

##Upgrade guide
Switch to the service account

	su cryptuser

Start bash

	bash

Change into the Crypt virtualenv directory

	cd /usr/local/crypt_env

Activate the virtualenv

	source bin/activate

Change into the Crypt directory and update the code from GitHub

	cd crypt
	git pull

Run the migration so your database is up to date

	python manage.py migrate

Now we need to generate some encryption keys:

	python generate_keyczart.py

Finally, as root (not cryptuser) restart Apache

	service apache2 restart
