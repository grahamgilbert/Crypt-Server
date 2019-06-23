Installation on Ubuntu 18.04.2 LTS
=====================
This document assumes a bare install of Ubuntu 18.04.2 LTS server. 

All commands should be run as root, unless specified

##Install Prerequisites
###Install Apache and the Apache modules

 	apt-get install apache2 libapache2-mod-wsgi-py3

###Install GCC (Needed for the encryption library)

	apt-get install gcc

###Install git

	apt-get install git

###Install the python C headers (so you can compile  the encryption library)

	apt-get install python3.6-dev 

###If you want to use MySQL, you the following

	apt-get install libmysqlclient-dev
 	apt-get install python3-pip
 	pip3 install mysqlclient

###Install the python dev tools

	apt-get install python3-setuptools

###Verify virtual env is installed

	virtualenv --version

###If is isn't, install it with

	apt-get install python3-venv

##Create a non-admin service account and group
Create the Crypt user:

	useradd -d /usr/local/crypt_env cryptuser

Create the Crypt group:

	groupadd cryptgroup

Add cryptuser to the cryptgroup group:

	usermod -g cryptgroup cryptuser

##Create the virtual environment
When a virtualenv is created, pip will also be installed to manage a
virtualenv's local packages. Create a virtualenv which will handle
installing Django in a contained environment. In this example we'll
create a virtualenv for Crypt at /usr/local. This should be run from
Bash, as this is what the virtualenv activate script expects.

Go to where we're going to install the virtualenv:

	 cd /usr/local

Create the virtualenv for Crypt:

	python3 -m venv crypt_env

Make sure cryptuser has permissions to the new crypt_env folder:

	chown -R cryptuser crypt_env

Switch to the service account:

	su cryptuser

Virtualenv needs to be run from a bash prompt, so let's switch to one:

	bash

Now we can activate the virtualenv:

	cd crypt_env
	source bin/activate

##Install and configure Crypt
Still inside the crypt_env virtualenv, use git to clone the current
version of Crypt-Server

	git clone https://github.com/grahamgilbert/Crypt-Server.git crypt

Now we need to get the other components for Crypt

	pip3 install -r crypt/setup/requirements.txt

Now we need to generate some encryption keys (make sure these go in crypt/keyset):

	cd crypt
	python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key())"

You will need the key output here to be set as the variable FIELD_ENCRYPTION_KEY, in the settings.py file below. 
Next we need to make a copy of the example_settings.py file and put
in your info:

	cd ./fvserver
	cp example_settings.py settings.py

Edit settings.py:

* Set FIELD_ENCRYPTION_KEY to the encryption key generated above
* Set ADMINS to an administrative name and email
* Set TIME_ZONE to the appropriate timezone
* Change ALLOWED_HOSTS to be a list of hosts that the server will be
accessible from (e.g. ``ALLOWED_HOSTS=['crypt.grahamgilbert.dev']``

If you wish to use email notifications, add the following to your settings.py:

``` python
# This is the host and port you are sending email on
EMAIL_HOST = 'localhost'
EMAIL_PORT = '25'

# If your email server requires Authentication
EMAIL_HOST_USER = 'youruser'
EMAIL_HOST_PASSWORD = 'yourpassword'
# This is the URL at the front of any links in the emails
HOST_NAME = 'http://localhost'
```

## Using with MySQL
In order to use Crypt-Server with MySQL, you need to configure it to connect to
a MySQL server instead of the default sqlite3. To do this, locate the DATABASES
section of settings.py, and change ENGINE to 'django.db.backends.mysql'. Set the
NAME as the database name, USER and PASSWORD to your user and password, and
either leave HOST as blank for localhost, or insert an IP or hostname of your
MySQL server. You will also need to install the correct python and apt packages.

	apt-get install libmysqlclient-dev
 	apt-get install pip
 	pip3 install mysqlclient

## More Setup
We need to use Django's manage.py to initialise the app's database and
create an admin user. Running the syncdb command will ask you to create
an admin user - make sure you do this!

	cd ..
	python3 manage.py syncdb
	python3 manage.py migrate

Stage the static files (type yes when prompted)

	python3 manage.py collectstatic

##Installing mod_wsgi and configuring Apache
To run Crypt in a production environment, we need to set up a suitable
webserver. Make sure you exit out of the crypt_env virtualenv and the
cryptuser user (back to root) before continuing).

##Set up an Apache virtualhost
You will probably need to edit most of these bits to suit your
environment, especially to add SSL encryption. There are many different
options, especially if you prefer nginx, the below example is for apache
with an internal puppet CA. Make a new file at
/etc/apache2/sites-available (call it whatever you want)

	vim /etc/apache2/sites-available/crypt.conf

And then enter something like:

	<VirtualHost *:443>
        ServerName crypt.yourdomain.com
        WSGIScriptAlias / /usr/local/crypt_env/crypt/crypt.wsgi
        WSGIDaemonProcess cryptuser user=cryptuser group=cryptgroup
        Alias /static/ /usr/local/crypt_env/crypt/static/
        SSLEngine on
        SSLCertificateFile      "/etc/puppetlabs/puppet/ssl/certs/cryptserver.yourdomain.com.pem"
        SSLCertificateKeyFile   "/etc/puppetlabs/puppet/ssl/private_keys/cryptserver.yourdomain.com.pem"
        SSLCACertificatePath    "/etc/puppetlabs/puppet/ssl/certs"
        SSLCACertificateFile    "/etc/puppetlabs/puppet/ssl/certs/ca.pem"
        SSLCARevocationFile     "/etc/puppetlabs/puppet/ssl/crl.pem"
        SSLProtocol             +TLSv1
        SSLCipherSuite          ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA:ECDHE-RSA-AES128-SHA:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-SHA256:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA
        SSLHonorCipherOrder     On
        <Directory /usr/local/crypt_env/crypt>
            WSGIProcessGroup cryptuser
            WSGIApplicationGroup %{GLOBAL}
            Options FollowSymLinks
            AllowOverride None
            Require all granted
        </Directory>
    </VirtualHost>
    WSGISocketPrefix /var/run/wsgi
    WSGIPythonHome /usr/local/crypt_env

Now we just need to enable our site, and then your can go and configure
your clients:

	a2ensite crypt.conf
	service apache2 reload
	
	
Upgrade on Ubuntu 18.04.2 LTS from Crypt 2 to Crypt 3
=====================
This document assumes that you have Ubuntu 18.04.2 LTS with Python 2 and non upgraded versions of Apache etc.. used for Crypt 2 installs.

All commands should be run as root, unless specified

##Upgrade Prerequisites
###Upgrade Apache and the Apache modules. This is critical. If you do not update apache WSGI to compile against python 3, your site will not load. 

	apt-get update
	apt-get upgrade
	
	apt-get install apache2 libapache2-mod-wsgi-py3

###Install GCC (Needed for the encryption library)

	apt-get update gcc

###Install git

	apt-get update git

###Install the python C headers (so you can compile  the encryption library)

	apt-get install python3.6-dev 

###If you want to use MySQL, you the following

	apt-get update libmysqlclient-dev
 	apt-get install python3-pip
 	pip3 install mysqlclient

###Install the python dev tools

	apt-get install python3-setuptools

###Verify virtual env is installed

	virtualenv --version

###If is isn't, install it with

	apt-get install python3-venv
	
##Create a non-admin service account and group

We are assuming that you already have a user if you were running Crypt 2 so if you need to create a new user refer to the above core install instructions otherwise skip.

##Update the virtual environment
When a virtualenv is created, pip will also be installed to manage a
virtualenv's local packages. Create a virtualenv which will handle
installing Django in a contained environment. In this example we'll
create a virtualenv for Crypt at /usr/local. This should be run from
Bash, as this is what the virtualenv activate script expects. 

For the update this process simply rebuilds the virtual environment. It will 
not overwrite it completely nor the files inside it. 

Go to where we're going to install the virtualenv:

	 cd /usr/local

Create the virtualenv for Crypt:

	python3 -m venv crypt_env

Make sure cryptuser has permissions to the new crypt_env folder:

	chown -R cryptuser crypt_env

Switch to the service account:

	su cryptuser

Virtualenv needs to be run from a bash prompt, so let's switch to one:

	bash

Now we can activate the virtualenv:

	cd crypt_env
	source bin/activate


##Update and configure Crypt
Still inside the crypt_env virtualenv, use git to clone the current
version of Crypt-Server

	cd crypt
	git pull

Now we need to get the other components for Crypt

	pip3 install -r setup/requirements.txt

Now we need to generate some encryption keys (make sure these go in crypt/keyset):

	python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key())"

You will need the key output here to be set as the variable FIELD_ENCRYPTION_KEY, in the settings.py file below. 
Next we need to make a copy of the example_settings.py file and put
in your info:

	cd ./fvserver
	nano settings.py
	
There are 2 blocks that have changed

MIDDLEWARE_SCRIPTS has changed to MIDDLEWARE and a new variable STATICFILES_STORAGE has been added underneath that block as shown below. 
	
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

	STATICFILES_STORAGE = "whitenoise.storage.CompressedManifestStaticFilesStorage"

TEMPLATES should be replaced fully by 
	
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
				"django.template.context_processors.media",
				"django.template.context_processors.static",
				"django.template.context_processors.tz",
				"django.contrib.messages.context_processors.messages",
				"fvserver.context_processors.crypt_version",
			],
			"debug": DEBUG,
		},
	}
	]
	
Finally INSTALLED_APPS has been updated as follows
	
	INSTALLED_APPS = (
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
	)

Edit settings.py:

* Set FIELD_ENCRYPTION_KEY to the encryption key generated above
* Set ADMINS to an administrative name and email
* Set TIME_ZONE to the appropriate timezone
* Change ALLOWED_HOSTS to be a list of hosts that the server will be
accessible from (e.g. ``ALLOWED_HOSTS=['crypt.grahamgilbert.dev']``

## Update the WSGI
The WSGI is hard coded with version 2.7 of Python and it needs to be modified before the WSGI will load. 

	nano /usr/local/crypt_env/crypt/crypt.wsgi
	
Modify the code to reflect the current version of your python 3. In this case I am using 3.6.

	site.addsitedir(os.path.join(CRYPT_ENV_DIR, 'lib/python3.6/site-packages'))


## Using with MySQL
In order to use Crypt-Server with MySQL, you need to configure it to connect to
a MySQL server instead of the default sqlite3. To do this, locate the DATABASES
section of settings.py, and change ENGINE to 'django.db.backends.mysql'. Set the
NAME as the database name, USER and PASSWORD to your user and password, and
either leave HOST as blank for localhost, or insert an IP or hostname of your
MySQL server. You will also need to install the correct python and apt packages.

	apt-get install libmysqlclient-dev
 	apt-get install pip
 	pip3 install mysqlclient

## More Setup
We need to use Django's manage.py to initialise the app's database and
create an admin user. Running the syncdb command will ask you to create
an admin user - make sure you do this!

	cd ..
	python3 manage.py migrate

Stage the static files (type yes when prompted)

	python3 manage.py collectstatic

##Installing mod_wsgi and configuring Apache
To run Crypt in a production environment, we need to set up a suitable
webserver. Make sure you exit out of the crypt_env virtualenv and the
cryptuser user (back to root) before continuing).

	service apache2 reload
	
	or
	
	service apache2 restart