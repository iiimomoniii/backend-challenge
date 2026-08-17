package mocks

import (
	"context"

	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
)

type TokenProvider struct {
	GenerateFunc func(ctx context.Context, payload domainauth.AuthPayload) (string, error)
	VerifyFunc   func(ctx context.Context, token string) (*domainauth.AuthPayload, error)
}

var _ domainauth.TokenProvider = (*TokenProvider)(nil)

func (m *TokenProvider) Generate(ctx context.Context, payload domainauth.AuthPayload) (string, error) {
	if m.GenerateFunc == nil {
		panic("mocks.TokenProvider: GenerateFunc not set")
	}
	return m.GenerateFunc(ctx, payload)
}

func (m *TokenProvider) Verify(ctx context.Context, token string) (*domainauth.AuthPayload, error) {
	if m.VerifyFunc == nil {
		panic("mocks.TokenProvider: VerifyFunc not set")
	}
	return m.VerifyFunc(ctx, token)
}
