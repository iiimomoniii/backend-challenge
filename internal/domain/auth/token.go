package auth

import (
	"context"
	"time"
)

type AuthPayload struct {
	UserID    string
	Email     string
	ExpiresAt time.Time
}

type TokenIssuer interface {
	Generate(ctx context.Context, claims AuthPayload) (string, error)
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*AuthPayload, error)
}

// TokenProvider is implemented by local token providers such as JWT.
type TokenProvider interface {
	TokenIssuer
	TokenVerifier
}
