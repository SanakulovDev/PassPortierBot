package crypto

import (
	"golang.org/x/crypto/argon2"
)

// DeriveKey generates a 32-byte AES key from a passphrase and a salt using Argon2id.
// This is significantly more secure than simple SHA-256 hashing as it resists GPU/ASIC attacks.
//
// Parameters aligned with OWASP recommendations for backend systems:
// - Time: 1
// - Memory: 64 MB (64 * 1024)
// - Threads: 4
// - KeyLen: 32 bytes (for AES-256)
func DeriveKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 32)
}
