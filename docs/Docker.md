# Using Docker

## Prepare for first use
When starting from scratch, create a new empty file on the docker host to hold the sqlite3 secrets database
``` bash
touch /somewhere/else/on/the/host
```

## Upgrading from Crypt Server 2

The encryption method has changed in Crypt Server. You should pass in both your old encryption keys (e.g. `-v /somewhere/on/the/host:/home/docker/crypt/keyset`) and the new one (see below) for the first run to migrate your keys. After the migration you no longer need your old encryption keys. Crypt 3 is a major update, you should ensure any custom settings you pass are still valid.

## Basic usage
``` bash
docker run -d --name="Crypt" \
--restart="always" \
-v /somewhere/else/on/the/host:/home/docker/crypt/crypt.db \
-e FIELD_ENCRYPTION_KEY='yourencryptionkey' \
-p 8000:8000 \
macadmins/crypt-server
```

The secrets are encrypted, with the encryption key passed in as an environment variable. You should back this up as the keys are not recoverable without them.

### Generating an encryption key

Run the following command to generate an encryption key (you should specify the string only):

```
docker run --rm -ti macadmins/crypt-server \
python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key())"
```

## Backing up the database with a data dump
``` bash
docker exec -it Crypt bash
cd /home/docker/crypt/
python manage.py dumpdata > db.json
exit
docker cp Crypt:/home/docker/crypt/db.json .
```
Optionally
``` bash
docker exec -it Crypt bash
rm /home/docker/crypt/db.json
exit
```

## Using Postgres as an external database

Crypt, by default, uses a sqlite3 database for the django db backend.  Crypt also supports using Postgres as the django db backend.  If you would like to use an external Postgres server, you need to set the following environment variables:

```
docker run -d --name="Crypt" \
--restart="always" \
-p 8000:8000 \
-e DB_HOST='db.example.com' \
-e DB_PORT='5432' \
-e DB_NAME='postgres_dbname' \
-e DB_USER='postgres_user' \
-e DB_PASS='postgres_user_pass' \
-e FIELD_ENCRYPTION_KEY='yourencryptionkey' \
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
-e EMAIL_HOST='mail.yourdomain.com' \
-e EMAIL_PORT='25' \
-e EMAIL_USER='youruser' \
-e EMAIL_PASSWORD='yourpassword' \
-e HOST_NAME='https://crypt.myorg.com' \
-e FIELD_ENCRYPTION_KEY='yourencryptionkey' \
macadmins/crypt-server
```

If your SMTP server doesn't need a setting (username and password for example), you should omit it. The `HOST_NAME` setting should be the hostname of your server - this will be used to generate links in emails.
