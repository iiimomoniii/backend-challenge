package oauth2

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"
)

type mockIntrospectionResponse struct {
	Active  bool   `json:"active"`
	Subject string `json:"sub,omitempty"`
	Email   string `json:"email,omitempty"`
	Expires int64  `json:"exp,omitempty"`
}

// NewLocalMockServer creates a development-only OAuth2 token and introspection server.
func NewLocalMockServer(addr, clientID, clientSecret, accessToken, subject, email string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		if !validMockClient(r, clientID, clientSecret) {
			http.Error(w, "invalid client credentials", http.StatusUnauthorized)
			return
		}
		if err := r.ParseForm(); err != nil {
			writeMockOAuthError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			writeMockOAuthError(w, http.StatusBadRequest, "unsupported_grant_type")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	mux.HandleFunc("POST /introspect", func(w http.ResponseWriter, r *http.Request) {
		if !validMockClient(r, clientID, clientSecret) {
			http.Error(w, "invalid introspection client credentials", http.StatusUnauthorized)
			return
		}
		if err := r.ParseForm(); err != nil {
			writeMockOAuthError(w, http.StatusBadRequest, "invalid_request")
			return
		}

		result := mockIntrospectionResponse{
			Active: subtle.ConstantTimeCompare([]byte(r.Form.Get("token")), []byte(accessToken)) == 1,
		}
		if result.Active {
			result.Subject = subject
			result.Email = email
			result.Expires = time.Now().Add(time.Hour).Unix()
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(result)
	})

	return &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
}

func validMockClient(r *http.Request, clientID, clientSecret string) bool {
	gotID, gotSecret, ok := r.BasicAuth()
	return ok &&
		subtle.ConstantTimeCompare([]byte(gotID), []byte(clientID)) == 1 &&
		subtle.ConstantTimeCompare([]byte(gotSecret), []byte(clientSecret)) == 1
}

func writeMockOAuthError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
