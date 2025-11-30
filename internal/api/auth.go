package api

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// BasicAuth middleware for API authentication
func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get credentials from environment or use defaults
		username := os.Getenv("AUTH_USERNAME")
		password := os.Getenv("AUTH_PASSWORD")

		// Default credentials if not set
		if username == "" {
			username = "admin"
		}
		if password == "" {
			password = "proxmox2024"
		}

		// Get credentials from request
		reqUsername, reqPassword, ok := r.BasicAuth()

		// Check if credentials are provided and valid
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(password)) != 1 {

			w.Header().Set("WWW-Authenticate", `Basic realm="Proxmox Auto-Restart API"`)
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Credentials are valid, proceed
		next.ServeHTTP(w, r)
	})
}
