# Using Docker

## Quick Start

### 1. Generate encryption keys

```bash
# Generate field encryption key
docker run --rm ghcr.io/grahamgilbert/crypt-server ./cryptctl gen-key > field-encryption-key.txt

# Generate session key
docker run --rm ghcr.io/grahamgilbert/crypt-server ./cryptctl gen-key > session-key.txt
```

### 2. Create database file

For SQLite (simplest setup):

```bash
touch /path/to/crypt.db
```

### 3. Run the container

```bash
docker run -d --name="crypt" \
  --restart="always" \
  -v /path/to/crypt.db:/data/crypt.db \
  -e FIELD_ENCRYPTION_KEY="$(cat field-encryption-key.txt)" \
  -e SESSION_KEY="$(cat session-key.txt)" \
  -e SQLITE_PATH=/data/crypt.db \
  -p 8080:8080 \
  ghcr.io/grahamgilbert/crypt-server
```

### 4. Create admin user

```bash
docker exec crypt ./crypt-server \
  -create-admin \
  -username=admin \
  -password='your-secure-password'
```

### 5. Verify operation

```bash
docker logs crypt
```

Access the web interface at `http://localhost:8080`.

## Using PostgreSQL

For production deployments, PostgreSQL is recommended:

```bash
docker run -d --name="crypt" \
  --restart="always" \
  -e FIELD_ENCRYPTION_KEY="$(cat field-encryption-key.txt)" \
  -e SESSION_KEY="$(cat session-key.txt)" \
  -e DATABASE_URL="postgres://user:pass@db.example.com:5432/crypt" \
  -p 8080:8080 \
  ghcr.io/grahamgilbert/crypt-server
```

## Environment Variables

### Required

| Variable | Description |
|----------|-------------|
| `FIELD_ENCRYPTION_KEY` | Base64-encoded 32-byte key for encrypting secrets |
| `SESSION_KEY` | Random string (at least 32 bytes) for signing session cookies |

### Database (one required)

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `SQLITE_PATH` | Path to SQLite database file |

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `SESSION_COOKIE_SECURE` | `false` | Set to `true` when using HTTPS |
| `SAML_CONFIG_FILE` | - | Path to SAML configuration YAML file |
| `APPROVE_OWN` | `false` | Allow users to approve their own requests |
| `ALL_APPROVE` | `false` | Grant all users approval permissions |
| `ROTATE_VIEWED_SECRETS` | `false` | Instruct clients to rotate secrets after viewing |

## Docker Compose

Create a `docker-compose.yml`:

```yaml
services:
  crypt:
    image: ghcr.io/grahamgilbert/crypt-server
    restart: always
    ports:
      - "8080:8080"
    environment:
      - FIELD_ENCRYPTION_KEY=${FIELD_ENCRYPTION_KEY}
      - SESSION_KEY=${SESSION_KEY}
      - SQLITE_PATH=/data/crypt.db
      - SESSION_COOKIE_SECURE=true
    volumes:
      - ./data:/data
```

Create a `.env` file:

```bash
FIELD_ENCRYPTION_KEY=your-base64-key-here
SESSION_KEY=your-session-key-here
```

Run:

```bash
mkdir -p data && touch data/crypt.db
docker compose up -d
```

## SSL/TLS

Using Crypt without SSL **will** result in your secrets being compromised. Options:

1. **Reverse proxy** (recommended): Use nginx, Caddy, or Traefik in front of Crypt
2. **Load balancer**: Terminate SSL at your load balancer

Example with Caddy:

```yaml
services:
  crypt:
    image: ghcr.io/grahamgilbert/crypt-server
    environment:
      - FIELD_ENCRYPTION_KEY=${FIELD_ENCRYPTION_KEY}
      - SESSION_KEY=${SESSION_KEY}
      - SQLITE_PATH=/data/crypt.db
      - SESSION_COOKIE_SECURE=true
    volumes:
      - ./data:/data

  caddy:
    image: caddy:2
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
    depends_on:
      - crypt

volumes:
  caddy_data:
```

Example `Caddyfile`:

```
crypt.example.com {
    reverse_proxy crypt:8080
}
```

## Backing Up

### SQLite

```bash
# Stop container first for consistent backup
docker stop crypt
cp /path/to/crypt.db /path/to/backup/crypt.db.backup
docker start crypt
```

### PostgreSQL

Use standard PostgreSQL backup tools (`pg_dump`).

### Important

Always back up your `FIELD_ENCRYPTION_KEY`. Secrets cannot be recovered without it.

## Password Reset

```bash
docker exec crypt ./crypt-server \
  -reset-password \
  -username=admin \
  -password='new-password'
```
