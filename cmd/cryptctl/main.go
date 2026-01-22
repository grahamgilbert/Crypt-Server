package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"crypt-server/internal/crypto"
	"github.com/fernet/fernet-go"
)

type fixtureEntry struct {
	Model  string                 `json:"model"`
	PK     interface{}            `json:"pk"`
	Fields map[string]interface{} `json:"fields"`
}

// pkInt returns the integer PK value, or 0 if not an int/float (e.g., string session keys).
func (e fixtureEntry) pkInt() int {
	switch v := e.PK.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

type migrationOutput struct {
	Computers []computerOut `json:"computers"`
	Secrets   []secretOut   `json:"secrets"`
	Requests  []requestOut  `json:"requests"`
	Users     []userOut     `json:"users"`
}

type computerOut struct {
	ID           int    `json:"id"`
	Serial       string `json:"serial"`
	Username     string `json:"username"`
	ComputerName string `json:"computername"`
	LastCheckin  string `json:"last_checkin"`
}

type secretOut struct {
	ID               int    `json:"id"`
	ComputerID       int    `json:"computer_id"`
	SecretType       string `json:"secret_type"`
	Secret           string `json:"secret"`
	DateEscrowed     string `json:"date_escrowed"`
	RotationRequired bool   `json:"rotation_required"`
}

type requestOut struct {
	ID                int    `json:"id"`
	SecretID          int    `json:"secret_id"`
	RequestingUser    string `json:"requesting_user"`
	Approved          *bool  `json:"approved"`
	AuthUser          string `json:"auth_user"`
	ReasonForRequest  string `json:"reason_for_request"`
	ReasonForApproval string `json:"reason_for_approval"`
	DateRequested     string `json:"date_requested"`
	DateApproved      string `json:"date_approved"`
	Current           bool   `json:"current"`
}

type userOut struct {
	ID                int      `json:"id"`
	Username          string   `json:"username"`
	Email             string   `json:"email"`
	IsStaff           bool     `json:"is_staff"`
	IsSuper           bool     `json:"is_superuser"`
	CanApprove        bool     `json:"can_approve"`
	Groups            []string `json:"groups"`
	PasswordHash      string   `json:"password_hash"`
	MustResetPassword bool     `json:"must_reset_password"`
	LocalLoginEnabled bool     `json:"local_login_enabled"`
	AuthSource        string   `json:"auth_source"`
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: cryptctl <command>")
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintln(flag.CommandLine.Output(), "Commands:")
		fmt.Fprintln(flag.CommandLine.Output(), "  gen-key           Generate a base64-encoded 32-byte FIELD_ENCRYPTION_KEY")
		fmt.Fprintln(flag.CommandLine.Output(), "  import-fixture    Convert Django JSON fixtures into an encrypted migration export")
		fmt.Fprintln(flag.CommandLine.Output(), "  integration-test  Run integration tests against a database")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	switch flag.Arg(0) {
	case "gen-key":
		if err := runGenKey(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "import-fixture":
		if err := runImportFixture(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "integration-test":
		if err := runIntegrationTest(os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", flag.Arg(0))
		flag.Usage()
		os.Exit(2)
	}
}

func runGenKey(w io.Writer) error {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	_, err := fmt.Fprintln(w, encoded)
	return err
}

func runImportFixture(args []string) error {
	fs := flag.NewFlagSet("import-fixture", flag.ExitOnError)
	inputPath := fs.String("input", "", "Path to Django JSON fixture file")
	outputPath := fs.String("output", "", "Path to write migration export JSON")
	legacyKey := fs.String("legacy-key", "", "Base64 legacy FIELD_ENCRYPTION_KEY")
	legacyKeyFile := fs.String("legacy-key-file", "", "Path to file containing legacy FIELD_ENCRYPTION_KEY")
	newKey := fs.String("new-key", "", "Base64 new FIELD_ENCRYPTION_KEY")
	newKeyFile := fs.String("new-key-file", "", "Path to file containing new FIELD_ENCRYPTION_KEY")
	passwordMapPath := fs.String("password-map", "", "Path to CSV file mapping usernames/emails to passwords")
	fs.Parse(args)

	if *inputPath == "" || *outputPath == "" {
		return errors.New("input and output paths are required")
	}

	legacyKeyValue, err := loadKey(*legacyKey, *legacyKeyFile, "LEGACY_FIELD_ENCRYPTION_KEY")
	if err != nil {
		return fmt.Errorf("load legacy key: %w", err)
	}
	newKeyValue, err := loadKey(*newKey, *newKeyFile, "FIELD_ENCRYPTION_KEY")
	if err != nil {
		return fmt.Errorf("load new key: %w", err)
	}

	legacyFernetKey, err := fernet.DecodeKey(legacyKeyValue)
	if err != nil {
		return fmt.Errorf("decode legacy key: %w", err)
	}
	newCodec, err := crypto.NewAesGcmCodecFromBase64Key(newKeyValue)
	if err != nil {
		return fmt.Errorf("invalid new key: %w", err)
	}

	fixtureBytes, err := os.ReadFile(*inputPath)
	if err != nil {
		return fmt.Errorf("read fixture: %w", err)
	}

	entries, err := parseFixture(fixtureBytes)
	if err != nil {
		return fmt.Errorf("parse fixture: %w", err)
	}

	passwordMap, err := loadPasswordMap(*passwordMapPath)
	if err != nil {
		return fmt.Errorf("load password map: %w", err)
	}

	output, err := convertFixture(entries, legacyFernetKey, newCodec, passwordMap)
	if err != nil {
		return fmt.Errorf("convert fixture: %w", err)
	}

	payload, err := marshalOutput(output)
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	if err := os.WriteFile(*outputPath, payload, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func loadKey(value, path, env string) (string, error) {
	if value != "" {
		return strings.TrimSpace(value), nil
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	if envValue := os.Getenv(env); envValue != "" {
		return strings.TrimSpace(envValue), nil
	}
	return "", fmt.Errorf("missing key: provide --key, --key-file, or %s", env)
}
