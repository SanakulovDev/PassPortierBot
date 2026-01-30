// Package services provides business logic layer for PassPortierBot.
// It coordinates between handlers, repository, and crypto operations.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/repository"
	"passportier-bot/internal/vault"
)

// SecretService handles all business logic for user secrets.
// It's responsible for encryption/decryption and coordinating with repository.
type SecretService struct {
	repo   *repository.SecretRepository
	crypto *crypto.CryptoManager
}

// DecryptedSecret represents a secret after decryption.
type DecryptedSecret struct {
	KeyName string
	Value   string
}

// NewSecretService creates a new secret service instance.
func NewSecretService(repo *repository.SecretRepository) *SecretService {
	return &SecretService{
		repo:   repo,
		crypto: crypto.NewCryptoManager(),
	}
}

// SaveSecret encrypts and stores a secret.
// If key already exists, it will be updated (upsert behavior).
func (s *SecretService) SaveSecret(ctx context.Context, userID int64, keyName, plainValue string) error {
	// Get session passphrase from vault (RAM only)
	passphrase, ok := vault.GetKey(userID)
	if !ok {
		return errors.New("session not found - please unlock first")
	}

	// Encrypt with unique salt per encryption
	encrypted, err := s.crypto.Encrypt(plainValue, passphrase)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Upsert to database
	return s.repo.Upsert(ctx, userID, keyName, encrypted)
}

// GetSecret retrieves and decrypts a single secret.
func (s *SecretService) GetSecret(ctx context.Context, userID int64, keyName string) (*DecryptedSecret, error) {
	passphrase, ok := vault.GetKey(userID)
	if !ok {
		return nil, errors.New("session not found - please unlock first")
	}

	secret, err := s.repo.GetByKey(ctx, userID, keyName)
	if err != nil {
		return nil, fmt.Errorf("secret not found: %w", err)
	}

	// Decrypt - this also validates the password (Zero-Knowledge)
	plainValue, err := s.crypto.Decrypt(secret.EncryptedValue, passphrase)
	if err != nil {
		return nil, crypto.ErrInvalidPassword
	}

	return &DecryptedSecret{
		KeyName: secret.KeyName,
		Value:   plainValue,
	}, nil
}

// ListAllSecrets retrieves and decrypts ALL secrets for a user.
// Returns formatted output and slice of decrypted secrets.
func (s *SecretService) ListAllSecrets(ctx context.Context, userID int64) ([]DecryptedSecret, error) {
	passphrase, ok := vault.GetKey(userID)
	if !ok {
		return nil, errors.New("session not found - please unlock first")
	}

	// Fetch all encrypted secrets from DB
	secrets, err := s.repo.GetAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secrets: %w", err)
	}

	if len(secrets) == 0 {
		return nil, nil
	}

	// Decrypt each secret
	result := make([]DecryptedSecret, 0, len(secrets))
	for _, secret := range secrets {
		plainValue, err := s.crypto.Decrypt(secret.EncryptedValue, passphrase)
		if err != nil {
			// Mark as decryption error but continue with others
			result = append(result, DecryptedSecret{
				KeyName: secret.KeyName,
				Value:   "[decryption error]",
			})
			continue
		}
		result = append(result, DecryptedSecret{
			KeyName: secret.KeyName,
			Value:   plainValue,
		})
	}

	return result, nil
}

// FormatSecretsForDisplay creates a formatted string for Telegram display.
func (s *SecretService) FormatSecretsForDisplay(secrets []DecryptedSecret) string {
	if len(secrets) == 0 {
		return "📭 No secrets stored."
	}

	var sb strings.Builder
	sb.WriteString("📋 *Your Secrets:*\n\n")

	for i, secret := range secrets {
		if secret.Value == "[decryption error]" {
			sb.WriteString(fmt.Sprintf("%d. *%s*: ❌ _error_\n", i+1, secret.KeyName))
		} else {
			sb.WriteString(fmt.Sprintf("%d. *%s*: `%s`\n", i+1, secret.KeyName, secret.Value))
		}
	}

	sb.WriteString("\n⚠️ _Expires in 10 seconds_")
	return sb.String()
}

// DeleteSecret removes a secret by key name.
func (s *SecretService) DeleteSecret(ctx context.Context, userID int64, keyName string) error {
	// Verify session exists
	_, ok := vault.GetKey(userID)
	if !ok {
		return errors.New("session not found - please unlock first")
	}

	return s.repo.Delete(ctx, userID, keyName)
}

// CountSecrets returns the number of secrets stored by a user.
func (s *SecretService) CountSecrets(ctx context.Context, userID int64) (int64, error) {
	return s.repo.Count(ctx, userID)
}
