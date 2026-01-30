// Package repository provides data access layer for PassPortierBot.
// It implements the Repository Pattern for clean separation of concerns.
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Secret represents a user's encrypted secret stored in the database.
type Secret struct {
	ID             uint           `gorm:"primaryKey"`
	UserID         int64          `gorm:"index;not null"`
	KeyName        string         `gorm:"column:service;not null"`    // Maps to existing "service" column
	EncryptedValue string         `gorm:"column:encrypted_data;not null"` // Maps to existing column
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name to use existing table.
func (Secret) TableName() string {
	return "password_entries"
}

// SecretRepository handles all database operations for secrets.
// It provides clean abstraction over GORM with explicit SQL-like operations.
type SecretRepository struct {
	db *gorm.DB
}

// NewSecretRepository creates a new repository instance.
func NewSecretRepository(db *gorm.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

// Upsert inserts a new secret or updates existing one if key exists.
//
// SQL equivalent:
//
//	INSERT INTO password_entries (user_id, service, encrypted_data, updated_at)
//	VALUES ($1, $2, $3, NOW())
//	ON CONFLICT (user_id, service)
//	DO UPDATE SET
//	    encrypted_data = EXCLUDED.encrypted_data,
//	    updated_at = NOW();
//
// This is atomic and thread-safe.
func (r *SecretRepository) Upsert(ctx context.Context, userID int64, keyName, encryptedValue string) error {
	secret := Secret{
		UserID:         userID,
		KeyName:        keyName,
		EncryptedValue: encryptedValue,
	}

	// GORM's upsert using ON CONFLICT clause
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "service"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted_data", "updated_at"}),
	}).Create(&secret).Error
}

// GetAll retrieves all secrets for a user.
// Returns a slice of secrets ordered by key name for consistent display.
func (r *SecretRepository) GetAll(ctx context.Context, userID int64) ([]Secret, error) {
	var secrets []Secret
	
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("service ASC").
		Find(&secrets).Error
		
	return secrets, err
}

// GetByKey retrieves a single secret by user ID and key name.
// Uses case-insensitive partial matching for user convenience.
func (r *SecretRepository) GetByKey(ctx context.Context, userID int64, keyName string) (*Secret, error) {
	var secret Secret
	
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND LOWER(service) LIKE ?", userID, "%"+keyName+"%").
		First(&secret).Error
		
	if err != nil {
		return nil, err
	}
	
	return &secret, nil
}

// Delete removes a secret by user ID and key name.
// Uses soft delete if DeletedAt field is present.
func (r *SecretRepository) Delete(ctx context.Context, userID int64, keyName string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND LOWER(service) = ?", userID, keyName).
		Delete(&Secret{}).Error
}

// Count returns the number of secrets stored for a user.
func (r *SecretRepository) Count(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Secret{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
