# Installation on CentOS 7

This document has not been updated for several years and should only be used for version 2 of Crypt server. Pull requests to update this are gratefully accepted.

All commands should be run as root, unless specified.

## Install Prerequisites

### Setup and Virtual Environment

Install needed packages:

`yum install git python-setuptools gcc libffi-devel python-devel openssl-devel
postgresql-libs postgresql-devel`

Check if `virtualenv` is installed via `virtualenv --version` and install it if
needed:

`easy_install virtualenv`

### Create a non-admin service account and group

Create a new group:

`groupadd cryptgroup`

and add a new user in the cryptgroup with a home directory:

`useradd -g cryptgroup -m cryptuser`

### Create the virtual environment

When a virtualenv is created, pip will also be installed to manage a
virtualenv's local packages. Create a virtualenv which will handle installing
Django in a contained environment. In this example we'll create a virtualenv for
Crypt at /usr/local. This should be run from Bash, as this is what the
virtualenv activate script expects.

Switch to bash if needed: `/usr/bin/bash` and get into the local folder:

`cd /usr/local`

Create the virtialenv for Crypt `virtualenv crypt_env` and change folder
permissions: `chown -R cryptuser:cryptgroup crypt_env`.

Switch to the newly created service account `su cryptuser` and make sure to use
the bash shell: `bash`.

Now let's activate the virtualenv:

```
cd crypt_env
source bin/activate
```

### Copy and configure Crypt

Still inside the crypt_env virtualenv, use git to clone the current version of
Crypt-Server:

`git clone https://github.com/grahamgilbert/Crypt-Server.git crypt`


We could also get the 1.6.8 version via git without touching
the `requirements.txt`-file: `pip install git+https://github.com/django-extensions/django-extensions@243abe93451c3b53a5f562023afcd809b79c9b7f`.

Also install these aditional packages:

```
pip install psycopg2==2.5.3
pip install gunicorn
pip install setproctitle
```

Now we need to get the other missing components for Crypt via pip:

`pip install -r crypt/setup/requirements.txt`

Now we need to generate some encryption keys (dont forget to change directory!):

```
cd crypt
python ./generate_keyczart.py
```

Next we need to make a copy of the example_settings.py file and put in your
info:

```
cd fvserver
cp example_settings.py settings.py
vim settings.py
```

Atleast change the following:
- Set ADMINS to an administrative name and email
- Set TIME_ZONE to the appropriate timezone
- Change ALLOWED_HOSTS to be a list of hosts that the server will be accessible
  from.
- Take a look at the `DATABASES` and email settings.

### DB Setup

We need to use Django's manage.py to initialise the app's database and create an
admin user. Running the syncdb command will ask you to create an admin user -
make sure you do this!

```
cd ..
python manage.py syncdb
python manage.py migrate
```

If you used an external DB like Postgres you dont need to run `pyton manage.py syncdb`.

And stage the static files (type yes when prompted):

```
python manage.py collectstatic
```

Also create a new superuser to auth on the webinterface:

```
python manage.py createsuperuser --username $USERNAME
```

## Set up an Apache virtualhost

Exit out of the virtualenv and also switch back to root user. After that install
the Apache Modification `mod_wsgi`: `yum install mod_wsgi`.

Create the wsgi directory and give the cryptuser the needed rights:

Create a new VirtualHost `vim /etc/httpd/conf.d/crypt.conf`:

```
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
        Require all granted
    </Directory>
</VirtualHost>
WSGISocketPrefix /var/run/wsgi
WSGIPythonHome /home/app/crypt_env
```

### Configure SELinux to work with Apache

On CentOS SELinux is activated and needs to be configured so Apache can do it's work:

```
yum install -y policycoreutils-python
semanage fcontext -a -t httpd_sys_content_t "/usr/local/crypt_env/crypt(/.*)?"
semanage fcontext -a -t httpd_sys_rw_content_t "/usr/local/crypt_env/crypt(/.*)?"
setsebool -P httpd_can_sendmail on
setsebool -P httpd_can_network_connect_db on
restorecon -Rv /usr/local/crypt_env/crypt
```

If you enabled SSL also grant access to the key files:

```
semanage fcontext -a -t httpd_sys_rw_content_t "/etc/pki/tls/private/KEY.key"
restorecon -Rv /etc/pki/tls/private/KEY.key
```

### Open needed ports in the firewall

```
firewall-cmd --zone=public --add-service=https --permanent
firewall-cmd --reload
```

### Activate Apache and start the httpd-server

```
systemctl enable httpd
systemctl start httpd
```
