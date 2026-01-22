package main

import (
	"crypt-server/internal/app"
	"crypt-server/internal/crypto"
	"crypt-server/internal/migrate"
	"crypt-server/internal/store"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/crewjam/saml/samlsp"
)

func main() {
	printMigrations := flag.Bool("print-migrations", false, "Print embedded migrations and exit")
	validateMigrations := flag.Bool("validate-migrations", false, "Validate embedded migrations and exit")
	migrationsDriver := flag.String("migrations-driver", "", "Migrations driver to target (postgres or sqlite)")
	createAdmin := flag.Bool("create-admin", false, "Create the first admin user and exit")
	resetPassword := flag.Bool("reset-password", false, "Reset a user's password and exit")
	adminUsername := flag.String("username", "", "Username for admin operations")
	adminPassword := flag.String("password", "", "Password for admin operations")
	importFixturePath := flag.String("import-fixture", "", "Path to fixture JSON file to import (database must be empty)")
	flag.Parse()

	logger := log.New(os.Stdout, "crypt-server ", log.LstdFlags)
	if *printMigrations || *validateMigrations {
		if err := runMigrationCommand(os.Stdout, migrate.EmbeddedFS, *migrationsDriver, *validateMigrations, *printMigrations); err != nil {
			logger.Fatalf("migration command failed: %v", err)
		}
		return
	}

	encryptionKey := os.Getenv("FIELD_ENCRYPTION_KEY")
	codec, err := crypto.NewAesGcmCodecFromBase64Key(encryptionKey)
	if err != nil {
		logger.Fatalf("invalid encryption key: %v", err)
	}

	dbConfig, err := loadDatabaseConfig()
	if err != nil {
		logger.Fatal(err)
	}
	var dataStore store.Store
	switch dbConfig.driver {
	case "postgres":
		postgresStore, err := store.NewPostgresStore(dbConfig.dsn, codec)
		if err != nil {
			logger.Fatalf("database connection failed: %v", err)
		}
		pgFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, "postgres")
		if err != nil {
			logger.Fatalf("database migration failed: %v", err)
		}
		if err := migrate.Apply(postgresStore.DB(), "postgres", pgFS); err != nil {
			logger.Fatalf("database migration failed: %v", err)
		}
		dataStore = postgresStore
		logger.Printf("using postgres store")
	case "sqlite":
		sqliteStore, err := store.NewSQLiteStore(dbConfig.dsn, codec)
		if err != nil {
			logger.Fatalf("database connection failed: %v", err)
		}
		sqliteFS, err := migrate.SubMigrationsFS(migrate.EmbeddedFS, "sqlite")
		if err != nil {
			logger.Fatalf("database migration failed: %v", err)
		}
		if err := migrate.Apply(sqliteStore.DB(), "sqlite", sqliteFS); err != nil {
			logger.Fatalf("database migration failed: %v", err)
		}
		dataStore = sqliteStore
		logger.Printf("using sqlite store")
	default:
		logger.Fatalf("unsupported database driver: %s", dbConfig.driver)
	}

	// Wrap store with logging
	dataStore = store.NewLoggingStore(dataStore, logger)

	if *createAdmin {
		if err := createFirstAdmin(dataStore, *adminUsername, *adminPassword); err != nil {
			logger.Fatalf("create admin failed: %v", err)
		}
		logger.Printf("created first admin user: %s", *adminUsername)
		return
	}

	if *resetPassword {
		if err := resetUserPassword(dataStore, *adminUsername, *adminPassword); err != nil {
			logger.Fatalf("reset password failed: %v", err)
		}
		logger.Printf("password reset for user: %s", *adminUsername)
		return
	}

	if *importFixturePath != "" {
		logger.Printf("importing fixture from %s", *importFixturePath)
		if err := importFixture(dataStore, *importFixturePath); err != nil {
			logger.Fatalf("import fixture failed: %v", err)
		}
		logger.Printf("fixture imported successfully")
		return
	}
	renderer := app.NewRenderer("web/templates/layouts/base.html", "web/templates/pages")
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		logger.Fatal("SESSION_KEY is required")
	}
	sessionTTL := 24 * time.Hour
	sessionManager, err := app.NewSessionManager([]byte(sessionKey), "crypt_session", sessionTTL)
	if err != nil {
		logger.Fatalf("invalid session configuration: %v", err)
	}
	settings := app.Settings{
		ApproveOwn:             envBool("APPROVE_OWN", false),
		AllApprove:             envBool("ALL_APPROVE", false),
		SessionTTL:             sessionTTL,
		CookieSecure:           envBool("SESSION_COOKIE_SECURE", false),
		RequestCleanupInterval: time.Hour,
		RotateViewedSecrets:    envBool("ROTATE_VIEWED_SECRETS", false),
	}
	csrfManager := app.NewCSRFManager("crypt_csrf", 32)

	var samlSP *samlsp.Middleware
	var samlConfig *app.SAMLConfig
	samlConfigPath := os.Getenv("SAML_CONFIG_FILE")
	if samlConfigPath != "" {
		cfg, err := app.LoadSAMLConfig(samlConfigPath)
		if err != nil {
			logger.Fatalf("invalid saml config: %v", err)
		}
		samlProvider, err := app.BuildSAMLProvider(cfg)
		if err != nil {
			logger.Fatalf("saml setup failed: %v", err)
		}
		samlSP = samlProvider
		samlConfig = cfg
		logger.Printf("saml enabled")
	}

	server := app.NewServer(dataStore, renderer, logger, sessionManager, csrfManager, samlSP, samlConfig, settings)

	addr := ":8080"
	logger.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		logger.Fatalf("server stopped: %v", err)
	}
}

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
