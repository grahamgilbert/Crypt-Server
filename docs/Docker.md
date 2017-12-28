# Using Docker

## Basic usage

``` bash
docker run -d --name="Crypt" \
--restart="always" \
-v /somewhere/on/the/host:/home/docker/crypt/keyset \
-v /somewhere/else/on/the/host:/home/docker/crypt/crypt.db \
-p 8000:8000 \
macadmins/crypt-server
```

The secrets are encrypted, with the encryption keys stored at ``/home/docker/crypt/keyset``. You should back this up as the keys are not recoverable without them.

## Using an external database

Crypt, by default, uses a sqlite3 database for the django db backend.  If you would like to use an external database, such as Postgres, you need to set the following environment variables:
```
docker run -d --name="Crypt" \
--restart="always" \
-v /somewhere/on/the/host:/home/docker/crypt/keyset \
-p 8000:8000 \
-e DB_HOST='db.example.com' \
-e DB_PORT='5432' \
-e DB_NAME='postgres_dbname' \
-e DB_USER='postgres_user' \
-e DB_PASS='postgres_user_pass' \
macadmins/crypt-server
```

## Emails

If you would like Crypt to send emails when keys are requested and approved, you should set the following environment variables:

```
docker run -d --name="Crypt" \
--restart="always" \
-v /somewhere/on/the/host:/home/docker/crypt/keyset \
-v /somewhere/else/on/the/host:/home/docker/crypt/crypt.db \
-p 8000:8000 \
-e DOCKER_CRYPT_EMAIL_HOST='mail.yourdomain.com' \
-e DOCKER_CRYPT_EMAIL_PORT='25' \
-e DOCKER_CRYPT_EMAIL_USER='youruser' \
-e DOCKER_CRYPT_EMAIL_PASSWORD='yourpassword' \
-e DOCKER_CRYPT_HOST_NAME='https://crypt.myorg.com' \
macadmins/crypt-server
```

If your SMTP server doesn't need a setting (username and password for example), you should omit it. The `DOCKER_CRYPT_HOST_NAME` setting should be the hostname of your server - this will be used to generate links in emails.
