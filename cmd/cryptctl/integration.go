package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"crypt-server/internal/app"
	"crypt-server/internal/crypto"
	"crypt-server/internal/migrate"
	"crypt-server/internal/store"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func runIntegrationTest(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("integration-test", flag.ExitOnError)
	dbType := fs.String("db", "sqlite", "Database type: sqlite or postgres")
	dbURL := fs.String("db-url", "", "Database URL (for postgres) or file path (for sqlite)")
	encryptionKey := fs.String("key", "", "Base64 FIELD_ENCRYPTION_KEY")
	encryptionKeyFile := fs.String("key-file", "", "Path to file containing FIELD_ENCRYPTION_KEY")
	fs.Parse(args)

	keyValue, err := loadKey(*encryptionKey, *encryptionKeyFile, "FIELD_ENCRYPTION_KEY")
	if err != nil {
		return fmt.Errorf("load encryption key: %w", err)
	}

	codec, err := crypto.NewAesGcmCodecFromBase64Key(keyValue)
	if err != nil {
		return fmt.Errorf("create codec: %w", err)
	}

	var st store.Store
	var db *sql.DB

	switch *dbType {
	case "sqlite":
		dsn := *dbURL
		if dsn == "" {
			dsn = ":memory:"
		}
		sqliteStore, err := store.NewSQLiteStore(dsn, codec)
		if err != nil {
			return fmt.Errorf("open sqlite: %w", err)
		}
		st = sqliteStore
		db = sqliteStore.DB()
	case "postgres":
		if *dbURL == "" {
			return errors.New("db-url is required for postgres")
		}
		pgStore, err := store.NewPostgresStore(*dbURL, codec)
		if err != nil {
			return fmt.Errorf("open postgres: %w", err)
		}
		st = pgStore
		db = pgStore.DB()
	default:
		return fmt.Errorf("unsupported database type: %s", *dbType)
	}
	defer db.Close()

	// Run migrations using the embedded migration files
	fmt.Fprintln(stdout, "Running migrations...")
	migrationsFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, *dbType)
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}
	if err := migrate.Apply(db, *dbType, migrationsFS); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// Create test server with minimal configuration
	fmt.Fprintln(stdout, "Creating test server...")
	logger := log.New(io.Discard, "", 0)
	renderer := app.NewRenderer("web/templates/base.html", "web/templates")
	sessionKey := make([]byte, 32)
	sessionManager, err := app.NewSessionManager(sessionKey, "crypt_session", 24*time.Hour)
	if err != nil {
		return fmt.Errorf("create session manager: %w", err)
	}
	csrfManager := app.NewCSRFManager("csrf_token", 32)
	settings := app.Settings{}

	server := app.NewServer(st, renderer, logger, sessionManager, csrfManager, nil, nil, settings)
	handler := server.Routes()

	testSerial := "TEST-SERIAL-001"
	testUsername := "testuser"
	testMacName := "Test Mac"
	testSecret := "test-recovery-key-12345"
	testSecretType := "recovery_key"

	// Test 1: Send initial checkin
	fmt.Fprintln(stdout, "\n=== Test 1: Send initial checkin ===")
	if err := testCheckin(handler, stdout, testSerial, testUsername, testMacName, testSecret, testSecretType); err != nil {
		return fmt.Errorf("test 1 (initial checkin): %w", err)
	}

	// Test 2: Verify secret is stored and encrypted correctly
	fmt.Fprintln(stdout, "\n=== Test 2: Verify secret retrieval ===")
	secretCount, err := testSecretRetrieval(st, stdout, testSerial, testSecretType, testSecret)
	if err != nil {
		return fmt.Errorf("test 2 (secret retrieval): %w", err)
	}
	if secretCount != 1 {
		return fmt.Errorf("expected 1 secret, got %d", secretCount)
	}

	// Test 3: Send duplicate checkin (should NOT create new secret)
	fmt.Fprintln(stdout, "\n=== Test 3: Send duplicate checkin ===")
	if err := testCheckin(handler, stdout, testSerial, testUsername, testMacName, testSecret, testSecretType); err != nil {
		return fmt.Errorf("test 3 (duplicate checkin): %w", err)
	}

	// Test 4: Verify no duplicate was created
	fmt.Fprintln(stdout, "\n=== Test 4: Verify no duplicate ===")
	secretCount, err = testSecretRetrieval(st, stdout, testSerial, testSecretType, testSecret)
	if err != nil {
		return fmt.Errorf("test 4 (verify no duplicate): %w", err)
	}
	if secretCount != 1 {
		return fmt.Errorf("duplicate secret created! expected 1 secret, got %d", secretCount)
	}
	fmt.Fprintln(stdout, "PASS: No duplicate secret created")

	// Test 5: Send different secret (should create new entry)
	fmt.Fprintln(stdout, "\n=== Test 5: Send different secret ===")
	newSecret := "different-recovery-key-67890"
	if err := testCheckin(handler, stdout, testSerial, testUsername, testMacName, newSecret, testSecretType); err != nil {
		return fmt.Errorf("test 5 (different secret): %w", err)
	}

	// Test 6: Verify new secret was created
	fmt.Fprintln(stdout, "\n=== Test 6: Verify new secret created ===")
	secretCount, err = countSecretsForComputer(st, testSerial, testSecretType)
	if err != nil {
		return fmt.Errorf("test 6 (count secrets): %w", err)
	}
	if secretCount != 2 {
		return fmt.Errorf("expected 2 secrets after sending different secret, got %d", secretCount)
	}
	fmt.Fprintln(stdout, "PASS: New secret created for different value")

	fmt.Fprintln(stdout, "\n=== All integration tests passed! ===")
	return nil
}

func testCheckin(handler http.Handler, stdout io.Writer, serial, username, macName, secret, secretType string) error {
	form := url.Values{}
	form.Set("serial", serial)
	form.Set("username", username)
	form.Set("macname", macName)
	form.Set("recovery_password", secret)
	form.Set("secret_type", secretType)

	req := httptest.NewRequest(http.MethodPost, "/checkin/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return fmt.Errorf("checkin failed with status %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	fmt.Fprintf(stdout, "Checkin response: serial=%v, username=%v, rotation_required=%v\n",
		response["serial"], response["username"], response["rotation_required"])

	return nil
}

func testSecretRetrieval(st store.Store, stdout io.Writer, serial, secretType, expectedSecret string) (int, error) {
	computer, err := st.GetComputerBySerial(serial)
	if err != nil {
		return 0, fmt.Errorf("get computer: %w", err)
	}
	fmt.Fprintf(stdout, "Found computer: ID=%d, Serial=%s, Name=%s\n",
		computer.ID, computer.Serial, computer.ComputerName)

	secrets, err := st.ListSecretsByComputer(computer.ID)
	if err != nil {
		return 0, fmt.Errorf("list secrets: %w", err)
	}

	count := 0
	for _, s := range secrets {
		if s.SecretType == secretType {
			count++
			if s.Secret == expectedSecret {
				fmt.Fprintf(stdout, "PASS: Secret decrypted correctly: ID=%d, Type=%s\n", s.ID, s.SecretType)
			} else if s.Secret != "" {
				fmt.Fprintf(stdout, "Secret found: ID=%d, Type=%s\n", s.ID, s.SecretType)
			}
		}
	}

	if count == 0 {
		return 0, fmt.Errorf("no secrets found for type %s", secretType)
	}

	return count, nil
}

func countSecretsForComputer(st store.Store, serial, secretType string) (int, error) {
	computer, err := st.GetComputerBySerial(serial)
	if err != nil {
		return 0, fmt.Errorf("get computer: %w", err)
	}

	secrets, err := st.ListSecretsByComputer(computer.ID)
	if err != nil {
		return 0, fmt.Errorf("list secrets: %w", err)
	}

	count := 0
	for _, s := range secrets {
		if s.SecretType == secretType {
			count++
		}
	}
	return count, nil
}

