# Crypt-Server

**[Crypt][1]** is a tool for securely storing secrets such as FileVault 2 recovery keys. It is made up of a client app, and a web app for storing the keys.

## Features

- Secrets are encrypted in the database
- All access is audited - all reasons for retrieval and approval are logged along side the users performing the actions
- Two step approval for retrieval of secrets is enabled by default
- Approval permission can be given to all users (so just any two users need to approve the retrieval) or a specific group of users

  [1]: https://github.com/grahamgilbert/Crypt

## Migration from Django

### Step 1: Export data from Django

Export your Django database to a JSON fixture:

```bash
# If running Django directly:
cd /path/to/legacy/crypt-server
./manage.py dumpdata > legacy.json

# If running Django in Docker:
docker exec <container_name> python manage.py dumpdata > legacy.json
```

### Step 2: Generate a new encryption key

Generate a new AES-GCM encryption key for the Go backend:

```bash
./cryptctl gen-key > new-field-encryption-key.txt
```

### Step 3: Convert the fixture

Convert the Django JSON fixture into the new format. This re-encrypts all secrets from Django's Fernet encryption to the new AES-GCM format:

```bash
./cryptctl import-fixture \
  -input legacy.json \
  -output migration-export.json \
  -legacy-key-file legacy-field-encryption-key.txt \
  -new-key-file new-field-encryption-key.txt \
  -password-map password-map.csv
```

The optional password map CSV allows you to set passwords for users who should have local login enabled. Any users not in this map will be configured for SAML authentication only. The CSV should have the following format (including header row):

```csv
username_or_email,password,must_reset_password
admin@example.com,Str0ng!Passw0rd,false
```

Users not in the password map will be configured for SAML authentication only.

### Step 4: Import into the new server

Import the converted fixture into the Go server. **The database must be empty** (no existing computers, secrets, requests, or users).

First, set the required environment variables:

```bash
export FIELD_ENCRYPTION_KEY=$(cat new-field-encryption-key.txt)
export SESSION_KEY=$(./cryptctl gen-key)
```

Then run the import:

```bash
./crypt-server -import-fixture migration-export.json
```

The import will:

- Verify the database is empty (fails if any data exists)
- Import all computers with their original IDs
- Import all secrets (already re-encrypted with the new key)
- Import all users with their authentication settings
- Import all requests with their approval status

After import completes, start the server normally (the environment variables are already set):

```bash
./crypt-server
```

## Installation instructions

It is recommended that you use [Docker](docs/Docker.md) to run this. See the Docker documentation for complete setup instructions.

### Migrating from the Django version

If you are migrating from the Django version of Crypt Server, follow the "Migration from Django" steps above. If you are running a version earlier than Crypt 3.0, you should first upgrade to Django Crypt 3.2.0 to migrate from the legacy encryption format, then follow the migration steps to move to the Go version.

## Settings

All settings are configured via environment variables.

### Required

- `FIELD_ENCRYPTION_KEY` - Base64-encoded 32-byte key for encrypting secrets. Generate with `./cryptctl gen-key`.

- `SESSION_KEY` - A random string (at least 32 bytes) used to sign session cookies. Generate with `./cryptctl gen-key`.

### Database (one required)

- `DATABASE_URL` - PostgreSQL connection string (e.g., `postgres://user:pass@host:5432/dbname`). Mutually exclusive with `SQLITE_PATH`.

- `SQLITE_PATH` - SQLite database file path. Must be a file path (not `:memory:`). Mutually exclusive with `DATABASE_URL`.

### Optional

- `SESSION_COOKIE_SECURE` - Set to `true` to mark session cookies as secure (recommended when using HTTPS). Default: `false`.

- `SAML_CONFIG_FILE` - Path to a YAML file containing SAML configuration. See `docs/saml-config.sample.yaml` for all supported fields.

- `APPROVE_OWN` - Allow users with approval permissions to approve their own key requests. Default: `false`.

- `ALL_APPROVE` - Grant all users approval permissions when they log in. Default: `false`.

- `ROTATE_VIEWED_SECRETS` - Instruct compatible clients (Crypt 3.2.0+) to rotate and re-escrow secrets after viewing. Default: `false`.

## Database migrations

The Go server applies embedded SQL migrations on startup and records applied versions in `schema_migrations`.

Migration file naming: `NNN_description.sql` (for example, `002_add_requests.sql`).

Flags:

- `-validate-migrations` - Validate embedded migrations and exit.
- `-print-migrations` - Print embedded migrations and exit.
- `-migrations-driver` - Limit the validation/print target to `postgres` or `sqlite` (default: both).

Example:

``` bash
./crypt-server -validate-migrations -migrations-driver=postgres
```

## First admin creation

Create the initial admin user (only works when no users exist yet):

``` bash
./crypt-server -create-admin -username=admin -password='your-password'
```

## Password reset

Reset a user's password from the command line:

``` bash
./crypt-server -reset-password -username=admin -password='new-password'
```

## Screenshots

Main Page:
![Crypt Main Page](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/home.png)

Computer Info:
![Computer info](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/admin_computer_info.png)

User Key Request:
![Userkey request](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/user_key_request.png)

Manage Requests:
![Manage Requests](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/manage_requests.png)

Approve Request:
![Approve Request](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/approve_request.png)

Key Retrieval:
![Key Retrieval](https://raw.github.com/grahamgilbert/Crypt-Server/master/docs/images/key_retrieval.png)
