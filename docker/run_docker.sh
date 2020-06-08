CWD=`pwd`
docker rm -f crypt
docker build -t macadmins/crypt .
docker run -d \
    -e ADMIN_PASS=pass \
    -e DEBUG=true \
    -e PROMETHEUS=true \
    --name=crypt \
    --restart="always" \
    -v "$CWD/crypt.db":/home/docker/crypt/crypt.db \
    -v "$CWD/keyset":/home/docker/crypt/keyset \
    -e FIELD_ENCRYPTION_KEY=jKAv1Sde8m6jCYFnmps0iXkUfAilweNVjbvoebBrDwg= \
    -p 8000-8050:8000-8050 \
    macadmins/crypt
