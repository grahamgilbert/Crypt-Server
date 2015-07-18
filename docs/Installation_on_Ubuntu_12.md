Installation on Ubuntu 12.04 LTS
=====================
This document assumes Ubuntu 12.04 LTS. The instructions are largely based on the [CentOS MunkiWebAdmin setup instructions](https://code.google.com/p/munki/wiki/MunkiWebAdminLinuxSetup) by Timothy Sutton.

All commands should be run as root, unless specified

##Install Prerequisites
###Setup the Virtual Environment
Make sure git is installed:

	which git

If it isn't, isntall it:

	apt-get install git

Install python setup tools:

	apt-get install python-setuptools

Make sure virtualenv is installed

	virtualenv --version

If it's not, install it:

	easy_install virtualenv

###Create a non-admin service account and group
Create the Crypt user:

	useradd cryptuser

Create the Crypt group:

	groupadd cryptgroup

Add cryptuser to the cryptgroup group:

	usermod -g cryptgroup cryptuser

(You may also want a home folder for cryptuser, if it barks when spinning up your wsgi script)

##Create the virtual environment
When a virtualenv is created, pip will also be installed to manage a virtualenv's local packages. Create a virtualenv which will handle installing Django in a contained environment. In this example we'll create a virtualenv for Crypt at /usr/local. This should be run from Bash, as this is what the virtualenv activate script expects.

Go to where we're going to install the virtualenv:

	 cd /usr/local

Create the virtualenv for Crypt:

	virtualenv crypt_env

Make sure cryptuser has permissions to the new crypt_env folder:

	chown -R cryptuser crypt_env

Switch to the service account:

	su cryptuser

Virtualenv needs to be run from a bash prompt, so let's switch to one:

	bash

Now we can activate the virtualenv:

	cd crypt_env
	source bin/activate

##Copy and configure Crypt
Still inside the crypt_env virtualenv, use git to clone the current version of Crypt-Server

	git clone https://github.com/grahamgilbert/Crypt-Server.git crypt

Now we need to get the other components for Crypt

	pip install -r crypt/setup/requirements.txt

Now we need to generate some encryption keys:

	python generate_keyczart.py

Next we need to make a copy of the example_settings.py file and put in your info:

	cd crypt/fvserver
	cp example_settings.py settings.py

Edit settings.py:

* Set ADMINS to an administrative name and email
* Set TIME_ZONE to the appropriate timezone
* Change ALLOWED_HOSTS to be a list of hosts that the server will be accessible from (e.g. ``ALLOWED_HOSTS=['crypt.grahamgilbert.dev']``

###More Setup
We need to use Django's manage.py to initialise the app's database and create an admin user. Running the syncdb command will ask you to create an admin user - make sure you do this!

	cd ..
	python manage.py syncdb
	python manage.py migrate

Stage the static files (type yes when prompted)

	python manage.py collectstatic

##Installing mod_wsgi and configuring Apache
To run Crypt in a production environment, we need to set up a suitable webserver. Make sure you exit out of the crypt_env virtualenv and the cryptuser user (back to root) before continuing).

	apt-get install libapache2-mod-wsgi

##Set up an Apache virtualhost
You will probably need to edit most of these bits to suit your environment, especially to add SSL encryption. There are many different options, especially if you prefer nginx, the below example is for apache with an internal puppet CA. Make a new file at /etc/apache2/sites-available (call it whatever you want)

	vim /etc/apache2/sites-available/crypt.conf

And then enter something like:

	<VirtualHost *:443>
        ServerName crypt.yourdomain.com
        WSGIScriptAlias / /home/app/crypt_env/crypt/crypt.wsgi
        WSGIDaemonProcess cryptuser user=cryptuser group=cryptgroup
        Alias /static/ /home/app/crypt_env/crypt/static/
        SSLEngine on
        SSLCertificateFile      "/etc/puppetlabs/puppet/ssl/certs/cryptserver.yourdomain.com.pem"
        SSLCertificateKeyFile   "/etc/puppetlabs/puppet/ssl/private_keys/cryptserver.yourdomain.com.pem"
        SSLCACertificatePath    "/etc/puppetlabs/puppet/ssl/certs"
        SSLCACertificateFile    "/etc/puppetlabs/puppet/ssl/certs/ca.pem"
        SSLCARevocationFile     "/etc/puppetlabs/puppet/ssl/crl.pem"
        SSLProtocol             +TLSv1
        SSLCipherSuite          ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA:ECDHE-RSA-AES128-SHA:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-SHA256:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA
        SSLHonorCipherOrder     On
        <Directory /home/app/crypt_env/crypt>
            WSGIProcessGroup cryptuser
            WSGIApplicationGroup %{GLOBAL}
            Order deny,allow
            Allow from all
        </Directory>
    </VirtualHost>
    WSGISocketPrefix /var/run/wsgi
    WSGIPythonHome /home/app/crypt_env

Now we just need to enable our site, and then your can go and configure your clients (exit back to root for this):

	a2ensite crypt.conf
	service apache2 reload
