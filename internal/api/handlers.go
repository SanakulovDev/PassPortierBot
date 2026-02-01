package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"passportier-bot/internal/crypto"
	"passportier-bot/internal/vault"
)

// handlePasswords returns user's passwords.
func (s *Server) handlePasswords(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from query (for now, simple approach)
	// In production, validate Telegram initData
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	// Check if session is active
	userKey, err := s.sm.GetSession(context.Background(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "session_locked",
			"message": "Session yopiq. /unlock qiling.",
		})
		return
	}

	// Get passwords
	entries, err := vault.ListEntries(s.db, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Decrypt and build response
	cm := crypto.NewCryptoManager()
	var passwords []PasswordResponse

	for _, entry := range entries {
		decrypted, err := cm.Decrypt(entry.EncryptedData, userKey)
		if err != nil {
			continue
		}

		passwords = append(passwords, PasswordResponse{
			ID:      entry.ID,
			Service: entry.Service,
			Data:    decrypted,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"passwords": passwords,
		"count":     len(passwords),
	})
}

// handleDelete deletes a password entry.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID  int64  `json:"user_id"`
		Service string `json:"service"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check session
	if _, err := s.sm.GetSession(context.Background(), req.UserID); err != nil {
		http.Error(w, "Session locked", http.StatusUnauthorized)
		return
	}

	// Delete entry
	if err := vault.DeleteEntry(s.db, req.UserID, req.Service); err != nil {
		log.Printf("Delete error: %v", err)
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGetOne returns a single password entry.
func (s *Server) handleGetOne(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	service := r.URL.Query().Get("service")

	if userIDStr == "" || service == "" {
		http.Error(w, "user_id and service required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	userKey, err := s.sm.GetSession(context.Background(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "session_locked",
		})
		return
	}

	entry, err := vault.GetEntry(s.db, userID, service)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	cm := crypto.NewCryptoManager()
	decrypted, err := cm.Decrypt(entry.EncryptedData, userKey)
	if err != nil {
		http.Error(w, "Decrypt error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      entry.ID,
		"service": entry.Service,
		"data":    decrypted,
	})
}

// handleUpdate updates a password entry.
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID     int64  `json:"user_id"`
		OldService string `json:"old_service"`
		NewService string `json:"new_service"`
		Data       string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userKey, err := s.sm.GetSession(context.Background(), req.UserID)
	if err != nil {
		http.Error(w, "Session locked", http.StatusUnauthorized)
		return
	}

	// Delete old entry if service name changed
	if req.OldService != req.NewService {
		if err := vault.DeleteEntry(s.db, req.UserID, req.OldService); err != nil {
			log.Printf("Delete old entry error: %v", err)
		}
	}

	// Upsert with new data (UpsertCredential handles encryption internally)
	if err := vault.UpsertCredential(s.db, req.UserID, req.NewService, req.Data, userKey); err != nil {
		http.Error(w, "Save error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
