Applying the July 2015 Update
======================

If you have a new installation of Crypt, or you started with Crypt after the July 2015 update, you **do not need to do this. Stop now.**

The July 2015 update contains a significant upgrade to Django (the core framework that runs Crypt) - this contains lots of bug fixes and security patches (a 'good thing'). But you do need to perform one extra step after performing your normal update procedure (either following [this guide](https://github.com/grahamgilert/crypt-server/blob/master/docs/Upgrading_on_Ubuntu_12.md) if you are running using the older method of deploying Crypt, or after ``docker pull macadmins/crypt-server`` if you are using Docker).

As ever, **back up your data**. The easiest method is to use [Django's built in method](https://coderwall.com/p/mvsoyg/django-dumpdata-and-loaddata).

It is recommended that you start with a fresh ``fvserver/settings.py`` (copied from ``fvserver/example_settings.py``), as there have been quite a few changes. Plug in your database information as needed - if you're using Docker and not using your own ``settings.py``, you don't need to do anything.

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
$ python manage.py migrate --fake-initial
$ python manage.py migrate
$ python generate_keyczart.py
$ exit
```

# Legacy (Apache)

```
$ source /usr/local/crypt_env/bin/activate
$ cd /usr/local/crypt_env/sal
$ python manage.py migrate --fake-initial
$ python manage.py migrate
$ python generate_keyczart.py
```
