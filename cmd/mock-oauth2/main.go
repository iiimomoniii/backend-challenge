// Command mock-oauth2 starts a local token introspection mock for development.
// It accepts one fixed access token and does not implement token issuance or OAuth flows.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	Addr         string
	ClientID     string
	ClientSecret string
	AccessToken  string
	Subject      string
	Email        string
}

type response struct {
	Active  bool   `json:"active"`
	Subject string `json:"sub,omitempty"`
	Email   string `json:"email,omitempty"`
	Expires int64  `json:"exp,omitempty"`
}

func main() {
	cfg := config{
		Addr:         getEnv("MOCK_OAUTH2_ADDR", "127.0.0.1:8090"),
		ClientID:     getEnv("MOCK_OAUTH2_CLIENT_ID", "local-mock-client"),
		ClientSecret: getEnv("MOCK_OAUTH2_CLIENT_SECRET", "local-mock-secret"),
		AccessToken:  getEnv("MOCK_OAUTH2_ACCESS_TOKEN", "local-test-token"),
		Subject:      getEnv("MOCK_OAUTH2_SUBJECT", "local-user-1"),
		Email:        getEnv("MOCK_OAUTH2_EMAIL", "demo@example.test"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		if !validClient(r, cfg) {
			http.Error(w, "invalid client credentials", http.StatusUnauthorized)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form body", http.StatusBadRequest)
			return
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": cfg.AccessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	mux.HandleFunc("POST /introspect", func(w http.ResponseWriter, r *http.Request) {
		if !validClient(r, cfg) {
			http.Error(w, "invalid introspection client credentials", http.StatusUnauthorized)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form body", http.StatusBadRequest)
			return
		}

		result := response{Active: subtle.ConstantTimeCompare([]byte(r.Form.Get("token")), []byte(cfg.AccessToken)) == 1}
		if result.Active {
			result.Subject = cfg.Subject
			result.Email = cfg.Email
			result.Expires = time.Now().Add(time.Hour).Unix()
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(result)
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("local OAuth2 mock listening on http://%s (token and introspection endpoints)", cfg.Addr)
	log.Printf("test access token: %s", cfg.AccessToken)
	log.Fatal(server.ListenAndServe())
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func validClient(r *http.Request, cfg config) bool {
	clientID, clientSecret, ok := r.BasicAuth()
	return ok &&
		subtle.ConstantTimeCompare([]byte(clientID), []byte(cfg.ClientID)) == 1 &&
		subtle.ConstantTimeCompare([]byte(clientSecret), []byte(cfg.ClientSecret)) == 1
}

func writeOAuthError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
