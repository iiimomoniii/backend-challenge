package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Provider struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) *Provider {
	return &Provider{secret: []byte(secret), ttl: ttl}
}

type authPayload struct {
	UserID    string
	Email     string
	ExpiresAt time.Time
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (p *Provider) Generate(ctx context.Context, payload authPayload) (string, error) {
	expiresAt := payload.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(p.ttl)
	}

	c := jwtClaims{
		UserID: payload.UserID,
		Email:  payload.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(p.secret)
}

func (p *Provider) Verify(ctx context.Context, tokenString string) (*authPayload, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return &authPayload{UserID: c.UserID, Email: c.Email, ExpiresAt: c.ExpiresAt.Time}, nil
}