package mocks

import (
	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
)

type PasswordHasher struct {
	HashFunc   func(password string) (string, error)
	VerifyFunc func(hashedPassword, password string) error
}

var _ domainauth.PasswordHasher = (*PasswordHasher)(nil)

func (m *PasswordHasher) Hash(password string) (string, error) {
	if m.HashFunc == nil {
		panic("mocks.PasswordHasher: HashFunc not set")
	}
	return m.HashFunc(password)
}

func (m *PasswordHasher) Verify(hashedPassword, password string) error {
	if m.VerifyFunc == nil {
		panic("mocks.PasswordHasher: VerifyFunc not set")
	}
	return m.VerifyFunc(hashedPassword, password)
}
