# Development

## Prerequisites

- Go 1.22 or later
- SQLite or PostgreSQL for database

## Building

```bash
# Build the server
make build

# Build the migration tool
make cryptctl
```

## Running locally

### Quick start with SQLite

```bash
# Generate keys (save these for reuse)
export FIELD_ENCRYPTION_KEY=$(./cryptctl gen-key)
export SESSION_KEY=$(./cryptctl gen-key)

# Run with SQLite
make run-sqlite
```

Or manually:

```bash
export FIELD_ENCRYPTION_KEY=$(./cryptctl gen-key)
export SESSION_KEY=$(./cryptctl gen-key)
export SQLITE_PATH=./crypt.db

./crypt-server
```

### With PostgreSQL

```bash
export FIELD_ENCRYPTION_KEY=$(./cryptctl gen-key)
export SESSION_KEY=$(./cryptctl gen-key)
export DATABASE_URL="postgres://user:pass@localhost:5432/crypt"

./crypt-server
```

### Create admin user

```bash
./crypt-server -create-admin -username=admin -password='password'
```

## Testing

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/store/...
```

## Project structure

```
.
├── cmd/
│   ├── crypt-server/    # Main server binary
│   └── cryptctl/        # Migration/utility tool
├── internal/
│   ├── app/             # HTTP handlers, routing, sessions
│   ├── crypto/          # Encryption (AES-GCM)
│   ├── fixture/         # Migration fixture types
│   ├── migrate/         # Database migrations
│   └── store/           # Database layer (PostgreSQL, SQLite)
├── web/
│   └── templates/       # HTML templates
└── docs/                # Documentation
```

## Database migrations

Migrations are embedded in the binary and run automatically on startup.

```bash
# Validate migrations
./crypt-server -validate-migrations

# Print migrations
./crypt-server -print-migrations

# Target specific driver
./crypt-server -validate-migrations -migrations-driver=postgres
```

Migration files are in `internal/migrate/migrations/{postgres,sqlite}/`.

## Docker

```bash
# Build image
docker build -t crypt-server .

# Run
docker run -p 8080:8080 \
  -e FIELD_ENCRYPTION_KEY="..." \
  -e SESSION_KEY="..." \
  -e SQLITE_PATH=/data/crypt.db \
  -v ./data:/data \
  crypt-server
```
