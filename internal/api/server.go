package api

import (
	"log"
	"net/http"
	"os"

	"passportier-bot/internal/security"

	"gorm.io/gorm"
)

// PasswordResponse represents a password entry for API response.
type PasswordResponse struct {
	ID      uint   `json:"id"`
	Service string `json:"service"`
	Data    string `json:"data"`
}

// Server handles HTTP API requests.
type Server struct {
	db       *gorm.DB
	sm       *security.SessionManager
	botToken string
}

// NewServer creates a new API server.
func NewServer(db *gorm.DB, sm *security.SessionManager) *Server {
	return &Server{
		db:       db,
		sm:       sm,
		botToken: os.Getenv("BOT_TOKEN"),
	}
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	http.HandleFunc("/api/passwords", s.corsMiddleware(s.handlePasswords))
	http.HandleFunc("/api/password", s.corsMiddleware(s.handleGetOne))
	http.HandleFunc("/api/delete", s.corsMiddleware(s.handleDelete))
	http.HandleFunc("/api/update", s.corsMiddleware(s.handleUpdate))
	
	log.Printf("[API] Starting server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// corsMiddleware adds CORS headers.
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
