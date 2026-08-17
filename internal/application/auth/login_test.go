package auth

import (
	"context"
	"errors"
	"testing"

	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/mocks"
)

func TestLoginUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("missing email returns ErrInvalidCredentials", func(t *testing.T) {
		uc := NewLoginUseCase(&mocks.UserRepository{}, &mocks.PasswordHasher{}, &mocks.TokenProvider{})
		_, err := uc.Execute(ctx, LoginRequest{Password: "secret1"})
		if !errors.Is(err, domainuser.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("missing password returns ErrInvalidCredentials", func(t *testing.T) {
		uc := NewLoginUseCase(&mocks.UserRepository{}, &mocks.PasswordHasher{}, &mocks.TokenProvider{})
		_, err := uc.Execute(ctx, LoginRequest{Email: "ada@example.com"})
		if !errors.Is(err, domainuser.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("unknown email returns ErrInvalidCredentials, not ErrUserNotFound", func(t *testing.T) {
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return nil, domainuser.ErrUserNotFound
			},
		}
		uc := NewLoginUseCase(repo, &mocks.PasswordHasher{}, &mocks.TokenProvider{})
		_, err := uc.Execute(ctx, LoginRequest{Email: "ghost@example.com", Password: "secret1"})
		if !errors.Is(err, domainuser.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
		if errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatal("must not leak ErrUserNotFound to the caller")
		}
	})

	t.Run("wrong password returns ErrInvalidCredentials", func(t *testing.T) {
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return &domainuser.User{ID: "u1", Email: email, Password: "hashed"}, nil
			},
		}
		hasher := &mocks.PasswordHasher{
			VerifyFunc: func(hashedPassword, password string) error {
				return errors.New("mismatch")
			},
		}
		uc := NewLoginUseCase(repo, hasher, &mocks.TokenProvider{})
		_, err := uc.Execute(ctx, LoginRequest{Email: "ada@example.com", Password: "wrong"})
		if !errors.Is(err, domainuser.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("repository error other than not-found is propagated", func(t *testing.T) {
		wantErr := errors.New("connection refused")
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return nil, wantErr
			},
		}
		uc := NewLoginUseCase(repo, &mocks.PasswordHasher{}, &mocks.TokenProvider{})
		_, err := uc.Execute(ctx, LoginRequest{Email: "ada@example.com", Password: "secret1"})
		if !errors.Is(err, wantErr) {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("token generation error is propagated", func(t *testing.T) {
		wantErr := errors.New("signing failed")
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return &domainuser.User{ID: "u1", Email: email, Password: "hashed"}, nil
			},
		}
		hasher := &mocks.PasswordHasher{
			VerifyFunc: func(hashedPassword, password string) error { return nil },
		}
		tokens := &mocks.TokenProvider{
			GenerateFunc: func(ctx context.Context, payload domainauth.AuthPayload) (string, error) {
				return "", wantErr
			},
		}
		uc := NewLoginUseCase(repo, hasher, tokens)
		_, err := uc.Execute(ctx, LoginRequest{Email: "ada@example.com", Password: "secret1"})
		if !errors.Is(err, wantErr) {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("happy path returns token and user", func(t *testing.T) {
		user := &domainuser.User{ID: "u1", Name: "Ada", Email: "ada@example.com", Password: "hashed"}

		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return user, nil
			},
		}
		hasher := &mocks.PasswordHasher{
			VerifyFunc: func(hashedPassword, password string) error {
				if hashedPassword != "hashed" || password != "secret1" {
					t.Errorf("Verify called with unexpected args: (%q, %q)", hashedPassword, password)
				}
				return nil
			},
		}
		var generatedFor domainauth.AuthPayload
		tokens := &mocks.TokenProvider{
			GenerateFunc: func(ctx context.Context, payload domainauth.AuthPayload) (string, error) {
				generatedFor = payload
				return "signed.jwt.token", nil
			},
		}
		uc := NewLoginUseCase(repo, hasher, tokens)

		got, err := uc.Execute(ctx, LoginRequest{Email: "ada@example.com", Password: "secret1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Token != "signed.jwt.token" {
			t.Errorf("want token %q, got %q", "signed.jwt.token", got.Token)
		}
		if got.User != user {
			t.Error("want returned user to be the one found by SearchByEmail")
		}
		if generatedFor.UserID != user.ID || generatedFor.Email != user.Email {
			t.Errorf("want token generated for user %+v, got payload %+v", user, generatedFor)
		}
	})
}
