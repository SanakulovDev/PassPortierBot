// Package crypto provides production-ready encryption using Zero-Knowledge Architecture.
// The user's password is NEVER stored anywhere; verification happens via decryption only.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

// Security Constants
// These values are aligned with OWASP recommendations for backend systems.
const (
	SaltSize   = 16       // 128-bit salt for Argon2id
	NonceSize  = 12       // 96-bit nonce for AES-GCM (standard)
	KeySize    = 32       // 256-bit key for AES-256
	Argon2Time = 1        // Number of iterations
	Argon2Mem  = 64 * 1024 // Memory in KiB (64 MB)
	Argon2Threads = 4     // Parallelism factor
)

// Predefined errors for Zero-Knowledge verification
var (
	ErrInvalidPassword = errors.New("invalid password or data corrupted")
	ErrDataTooShort    = errors.New("encrypted data is malformed or too short")
)

// CryptoManager handles all encryption/decryption operations.
// It is stateless and safe for concurrent use.
type CryptoManager struct{}

// NewCryptoManager creates a new CryptoManager instance.
func NewCryptoManager() *CryptoManager {
	return &CryptoManager{}
}

// Encrypt encrypts plaintext using AES-256-GCM with Argon2id key derivation.
//
// Security Design:
// 1. A unique random salt is generated for EVERY encryption operation.
// 2. A unique random nonce is generated for EVERY encryption operation.
// 3. The salt is embedded in the output, enabling stateless decryption.
// 4. Output format: base64(Salt[16] + Nonce[12] + Ciphertext[N+16])
//
// The authentication tag (16 bytes) is appended to ciphertext by GCM.
func (cm *CryptoManager) Encrypt(plainText, userKey string) (string, error) {
	// Generate cryptographically secure random salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// Derive AES-256 key using Argon2id (memory-hard, resistant to GPU/ASIC)
	aesKey := argon2.IDKey([]byte(userKey), salt, Argon2Time, Argon2Mem, Argon2Threads, KeySize)

	// Create AES cipher block
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode (provides both confidentiality and authenticity)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate random nonce (NEVER reuse with same key)
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate the plaintext
	// GCM appends a 16-byte authentication tag to the ciphertext
	ciphertext := gcm.Seal(nil, nonce, []byte(plainText), nil)

	// Combine: Salt + Nonce + Ciphertext into single blob
	// This allows stateless decryption without external salt storage
	combined := make([]byte, SaltSize+NonceSize+len(ciphertext))
	copy(combined[0:SaltSize], salt)
	copy(combined[SaltSize:SaltSize+NonceSize], nonce)
	copy(combined[SaltSize+NonceSize:], ciphertext)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt decrypts data encrypted by Encrypt() method.
//
// Zero-Knowledge Verification:
// The password is validated ONLY by attempting decryption.
// If GCM authentication fails, it means the password is wrong.
// This is the core of Zero-Knowledge: we never store password hashes.
func (cm *CryptoManager) Decrypt(encryptedData, userKey string) (string, error) {
	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", ErrDataTooShort
	}

	// Validate minimum length: Salt + Nonce + (at least 1 byte ciphertext + 16 byte tag)
	minLen := SaltSize + NonceSize + 17
	if len(combined) < minLen {
		return "", ErrDataTooShort
	}

	// Extract components
	salt := combined[0:SaltSize]
	nonce := combined[SaltSize : SaltSize+NonceSize]
	ciphertext := combined[SaltSize+NonceSize:]

	// Re-derive the AES key using extracted salt and provided password
	aesKey := argon2.IDKey([]byte(userKey), salt, Argon2Time, Argon2Mem, Argon2Threads, KeySize)

	// Create AES cipher block
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", ErrInvalidPassword
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrInvalidPassword
	}

	// Attempt decryption and authentication
	// CRITICAL: If password is wrong, the derived key is wrong, and GCM will fail
	// GCM.Open returns an error if the authentication tag doesn't match
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// This is the Zero-Knowledge verification point
		// Wrong password = wrong key = authentication failure
		return "", ErrInvalidPassword
	}

	return string(plaintext), nil
}
