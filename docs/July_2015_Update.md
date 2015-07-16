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
If you modified your `crypt.wsgi`, keep in mind that this [new version of django requires mod_wsgi to be compiled against python2.7](http://stackoverflow.com/questions/28681398/re-building-python-and-mod-wsgi), which RHEL wouldn't have by default.

It is recommended that you start with a fresh ``fvserver/settings.py`` (copied from ``fvserver/example_settings.py``), as there have been quite a few changes. Plug in your database and time zone/language information as needed.


**Before** the last upgrade step, `python manage.py migrate`, (while still in the virtualenv,) run the following:

```
$ python manage.py migrate --fake-initial
```
You may see another warning that `changes are not yet reflected in a migration`, you would do the provided command and then run the final migrate 