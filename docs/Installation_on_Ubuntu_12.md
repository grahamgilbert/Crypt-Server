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

	pip install -r setup/requirements.txt
	
Next we need to make a copy of the example_settings.py file and put in your info:

	cd crypt/fvserver
	cp example_settings.py settings.py
	
Edit settings.py:
* Set ADMINS to an administrative name and email
* Set TIME_ZONE to the appropriate timezone

###More Setup
We need to use Django's manage.py to initialise the app's database and create an admin user. Running the syncdb command will ask you to create an admin user - make sure you do this!

	cd ..
	python manage.py syncdb
	python manage.py migrate
	
Stage the staic files (type yes when prompted)
	
	python manage.py collectstatic

##Installing mod_wsgi and configuring Apache
To run Crypt in a production environment, we need to set up a suitable webserver. Make sure you exit out of the crypt_env virtualenv and the cryptuser user (back to root) before continuing).

	apt-get install libapache2-mod-wsgi
	
##Set up an Apache virtualhost
You will probably need to edit some of these bits to suit your environment, but this works for me. Make a new file at /etc/apache2/sites-available (call it whatever you want)

	nano /etc/apache2/sites-available/crypt.conf
	
And then enter something like:

	<VirtualHost *:80>
	ServerName crypt.yourdomain.com
   	WSGIScriptAlias / /usr/local/crypt_env/crypt/crypt.wsgi
   	WSGIDaemonProcess crypt user=cryptuser group=cryptgroup
   	Alias /static/ /usr/local/crypt_env/crypt/static/
   	<Directory /usr/local/crypt_env/crypt>
    	   WSGIProcessGroup crypt
       		WSGIApplicationGroup %{GLOBAL}
       		Order deny,allow
       		Allow from all
   	</Directory>
	</VirtualHost>
	
We're nearly done. Switch back to your cryptuser user and create /usr/local/crypt_env/crypt/crypt.wsgi with the following contents:
	
	su cryptuser
	bash
	nano /usr/local/crypt_env/crypt/crypt.wsgi
	
And the contents of the file:

	import os, sys
	import site

	CRYPT_ENV_DIR = '/usr/local/crypt_env'

	# Use site to load the site-packages directory of our virtualenv
	site.addsitedir(os.path.join(CRYPT_ENV_DIR, 'lib/python2.7/site-packages'))

	# Make sure we have the virtualenv and the Django app itself added to our path
	sys.path.append(CRYPT_ENV_DIR)
	sys.path.append(os.path.join(CRYPT_ENV_DIR, 'crypt'))
	os.environ.setdefault("DJANGO_SETTINGS_MODULE", "fvserver.settings")
	import django.core.handlers.wsgi
	application = django.core.handlers.wsgi.WSGIHandler()
	
Now we just need to enable our site, and then your can go and configure your clients (exit back to root for this):

	a2ensite crypt.conf
	service apache2 reload