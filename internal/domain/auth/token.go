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

type TokenProvider interface {
	Generate(ctx context.Context, claims AuthPayload) (string, error)
	Verify(ctx context.Context, token string) (*AuthPayload, error)
}