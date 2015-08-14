Applying the July 2015 Update
======================

If you have a new installation of Crypt, or you started with Crypt after the July 2015 update, you **do not need to do this. Stop now.**

The July 2015 update contains a significant upgrade to Django (the core framework that runs Crypt) - this contains lots of bug fixes and security patches (a 'good thing'). But you do need to perform one extra step after performing your normal update procedure (either following [this guide](https://github.com/grahamgilbert/crypt-server/blob/master/docs/Upgrading_on_Ubuntu_12.md) if you are running using the older method of deploying Crypt, or after ``docker pull macadmins/crypt-server`` if you are using Docker).

As ever, **back up your data**. The easiest method is to use [Django's built in method](https://coderwall.com/p/mvsoyg/django-dumpdata-and-loaddata). If you get an error, you may just be able to copy the crypt.db file if you're using sqlite.

You only need to perform the following once.

# Docker

Stop and remove your Crypt container:

```
$ docker stop crypt
$ docker rm crypt
```

We're now going to run a temporary container to update the database - if you have any custom mounts (e.g. ``settings.py``), you should include them as you normally would, and replace the DB_* environment variables to match what you have used:
```
$ docker run -t -i --rm --link postgres-crypt:db -e DB_NAME=crypt -e DB_USER=admin -e DB_PASS=password macadmins/crypt-server /bin/bash
# We're in the container now
$ cd /home/app/crypt
$ python generate_keyczart.py
$ python manage.py migrate --fake-initial
$ python manage.py migrate
$ exit
```

# Legacy (Apache)

It is recommended that you start with a fresh ``fvserver/settings.py`` (copied from ``fvserver/example_settings.py``), as there have been quite a few changes. Plug in your database and time zone/language information as needed.


**Before** the last upgrade step, `python manage.py migrate`, (while still in the virtualenv,) run the following:

```
$ python manage.py migrate --fake-initial
```
You may see another warning that `changes are not yet reflected in a migration`, you would do the prompted command (`makemigration`) and then run the final migrate

If you're on RHEL or CENTOS 6, the [new version of django requires mod_wsgi to be compiled against python2.7](http://stackoverflow.com/questions/28681398/re-building-python-and-mod-wsgi), and other tweaks too numerous to mention may make this a upgrade a world of pain. But here's a few while the wounds are still fresh:  
You would be building python 2.7 as instructed in the link above(substitute in more recent versions as applicable), the cryptuser probably wants a home folder(with the .profile containing `export LD_LIBRARY_ROOT=/usr/local/lib`), the vitualenv may need to be rebuilt if you didn't explicitly use 2.7, the vhost wants the WSGIPythonHome to point to your venv, the cryptuser of course needs access to python2.7's complete lib and bin paths, and of course double-back `chcon`'ing applicable paths as necessary(e.g. `-R --reference=/var/www /path/to/allow/apache`).