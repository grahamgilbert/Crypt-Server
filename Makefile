VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X crypt-server/internal/app.Version=$(VERSION)"

.PHONY: build cryptctl clean test run-sqlite run-sqlite-saml

build:
	go build $(LDFLAGS) -o crypt-server ./cmd/crypt-server

cryptctl:
	go build -o cryptctl ./cmd/cryptctl

test:
	go test ./...

clean:
	rm -f crypt-server cryptctl

run-sqlite: build cryptctl
	@echo "Starting crypt-server with SQLite..."
	@echo "Server will be available at http://localhost:8080"
	@test -f .field-encryption-key || ./cryptctl gen-key > .field-encryption-key
	SQLITE_PATH=./crypt.db \
	FIELD_ENCRYPTION_KEY=$$(cat .field-encryption-key) \
	SESSION_KEY=$${SESSION_KEY:-$$(./cryptctl gen-key)} \
	./crypt-server

run-sqlite-saml: build cryptctl
	@echo "Starting crypt-server with SQLite and SAML..."
	@echo "Server will be available at http://localhost:8080"
	@test -f .field-encryption-key || ./cryptctl gen-key > .field-encryption-key
	@test -f saml-config.yaml || (echo "Error: saml-config.yaml not found" && exit 1)
	SQLITE_PATH=./crypt.db \
	FIELD_ENCRYPTION_KEY=$$(cat .field-encryption-key) \
	SESSION_KEY=$${SESSION_KEY:-$$(./cryptctl gen-key)} \
	SAML_CONFIG_FILE=./saml-config.yaml \
	./crypt-server
