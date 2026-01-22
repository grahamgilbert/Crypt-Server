package app

// HashPassword exposes the Argon2id hash for CLI utilities.
func HashPassword(plaintext string) (string, error) {
	return hashPassword(plaintext)
}
