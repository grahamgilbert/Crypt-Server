# Using Docker

## Server Initialization
This was last tested on Ubuntu 24.04 x86

``` bash
git clone https://github.com/grahamgilbert/Crypt-Server.git
```

Install Docker and Docker Compose plugin following instructions here:
``` bash
https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository
```

Restart the Docker services
``` bash
sudo systemctl restart docker
```

Ensure docker permissions are set. Log out then back in after running this command: 
``` bash
sudo usermod -aG docker $USER
```

We will now install the Docker buildx plugin. 
``` bash
sudo apt install docker-buildx
```

Verify plugin state: 
``` bash
docker info | head -n8
```

Response should be similar to: 
``` bash
Client:
 Version:    24.0.7
 Context:    default
 Debug Mode: false
 Plugins:
  buildx: Docker Buildx (Docker Inc.)
    Version:  0.12.1
    Path:     /usr/libexec/docker/cli-plugins/docker-buildx
```

Build the Dockerfile:
``` bash
docker buildx build -t crypt-server .
```

## Prepare for first use
When starting from scratch, create a new empty file on the docker host to hold the sqlite3 secrets database
``` bash
touch /somewhere/else/on/the/host
```

## Basic usage
``` bash
docker run -d --name="Crypt" \
--restart="always" \
-v /somewhere/else/on/the/host:/home/docker/crypt/crypt.db \
-e FIELD_ENCRYPTION_KEY='yourencryptionkey' \
-p 8000:8000 \
crypt-server
```

## Verify Operation
``` bash
docker logs Crypt
```

## Upgrading from Crypt Server 2

The encryption method has changed in Crypt Server. You should pass in both your old encryption keys (e.g. `-v /somewhere/on/the/host:/home/docker/crypt/keyset`) and the new one (see below) for the first run to migrate your keys. After the migration you no longer need your old encryption keys. Crypt 3 is a major update, you should ensure any custom settings you pass are still valid.



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
-e CSRF_TRUSTED_ORIGINS='https://FirstServer.com,https://SecondServer.com' \
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
-e CSRF_TRUSTED_ORIGINS='https://FirstServer.com,https://SecondServer.com' \
macadmins/crypt-server
```

If your SMTP server doesn't need a setting (username and password for example), you should omit it. The `HOST_NAME` setting should be the hostname of your server - this will be used to generate links in emails.

## SSL

It is recommended to use either an Nginx proxy in front of the Crypt app for SSL termination (outside of the scope of this document, see [here](https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-18-04) and [here](https://www.linode.com/docs/web-servers/nginx/use-nginx-reverse-proxy/) for more information), or to use Caddy. Caddy will also handle setting up letsencrypt SSL certificates for you. An example Caddyfile is included in `docker/Caddyfile`. Using Crypt without SSL __will__ result in your secrets being compromised.

_Note Caddy is only free for personal use. For commercial deployments you should build from source yourself or use Nginx._

## X-Frame-Options

The nginx config included with the docker container configures the X-Frame-Options as sameorigin. This protects against a potential attacker using iframes to do bad stuff with Crypt.

Depending on your environment you may need to also configure X-Frame-Options on any proxies in front of Crypt.

## docker-compose

An example `docker-compose.yml` is included. For basic usuage, you should only need to edit the `FIELD_ENCRYPTION_KEY`.
