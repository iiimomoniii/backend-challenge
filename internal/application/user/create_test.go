package user

import (
	"context"
	"errors"
	"testing"

	"github.com/iiimomoniii/backend-challenge/internal/mocks"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

func TestCreateUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("missing email returns ErrUsernameRequired", func(t *testing.T) {
		uc := NewCreateUseCase(&mocks.UserRepository{}, &mocks.PasswordHasher{})

		_, err := uc.Execute(ctx, CreateRequest{Name: "Ada", Password: "secret1"})

		if !errors.Is(err, domainuser.ErrUsernameRequired) {
			t.Fatalf("want ErrUsernameRequired, got %v", err)
		}
	})

	t.Run("missing name returns ErrNameRequired", func(t *testing.T) {
		uc := NewCreateUseCase(&mocks.UserRepository{}, &mocks.PasswordHasher{})

		_, err := uc.Execute(ctx, CreateRequest{Email: "ada@example.com", Password: "secret1"})

		if !errors.Is(err, domainuser.ErrNameRequired) {
			t.Fatalf("want ErrNameRequired, got %v", err)
		}
	})

	t.Run("missing password returns ErrPasswordRequired", func(t *testing.T) {
		uc := NewCreateUseCase(&mocks.UserRepository{}, &mocks.PasswordHasher{})

		_, err := uc.Execute(ctx, CreateRequest{Name: "Ada", Email: "ada@example.com"})

		if !errors.Is(err, domainuser.ErrPasswordRequired) {
			t.Fatalf("want ErrPasswordRequired, got %v", err)
		}
	})

	t.Run("short password returns ErrPasswordTooShort", func(t *testing.T) {
		uc := NewCreateUseCase(&mocks.UserRepository{}, &mocks.PasswordHasher{})

		_, err := uc.Execute(ctx, CreateRequest{Name: "Ada", Email: "ada@example.com", Password: "123"})

		if !errors.Is(err, domainuser.ErrPasswordTooShort) {
			t.Fatalf("want ErrPasswordTooShort, got %v", err)
		}
	})

	t.Run("duplicate email returns ErrUsernameAlreadyExists", func(t *testing.T) {
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return &domainuser.User{ID: "existing-id", Email: email}, nil
			},
		}
		uc := NewCreateUseCase(repo, &mocks.PasswordHasher{})

		_, err := uc.Execute(ctx, CreateRequest{Name: "Ada", Email: "ada@example.com", Password: "secret1"})

		if !errors.Is(err, domainuser.ErrUsernameAlreadyExists) {
			t.Fatalf("want ErrUsernameAlreadyExists, got %v", err)
		}
	})

	t.Run("repository error other than not-found is propagated", func(t *testing.T) {
		wantErr := errors.New("connection refused")
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return nil, wantErr
			},
		}
		uc := NewCreateUseCase(repo, &mocks.PasswordHasher{})

		_, err := uc.Execute(ctx, CreateRequest{Name: "Ada", Email: "ada@example.com", Password: "secret1"})

		if !errors.Is(err, wantErr) {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("hasher error is propagated", func(t *testing.T) {
		wantErr := errors.New("hash failed")
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return nil, domainuser.ErrUserNotFound
			},
		}
		hasher := &mocks.PasswordHasher{
			HashFunc: func(password string) (string, error) {
				return "", wantErr
			},
		}
		uc := NewCreateUseCase(repo, hasher)

		_, err := uc.Execute(ctx, CreateRequest{Name: "Ada", Email: "ada@example.com", Password: "secret1"})

		if !errors.Is(err, wantErr) {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("happy path creates and returns the user", func(t *testing.T) {
		var created *domainuser.User

		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return nil, domainuser.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, u *domainuser.User) error {
				created = u
				return nil
			},
		}
		hasher := &mocks.PasswordHasher{
			HashFunc: func(password string) (string, error) {
				return "hashed:" + password, nil
			},
		}
		uc := NewCreateUseCase(repo, hasher)

		got, err := uc.Execute(ctx, CreateRequest{Name: "  Ada  ", Email: "  ada@example.com  ", Password: "secret1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name != "Ada" {
			t.Errorf("want trimmed name %q, got %q", "Ada", got.Name)
		}
		if got.Email != "ada@example.com" {
			t.Errorf("want trimmed email %q, got %q", "ada@example.com", got.Email)
		}
		if got.Password != "hashed:secret1" {
			t.Errorf("want hashed password %q, got %q", "hashed:secret1", got.Password)
		}
		if got.ID == "" {
			t.Error("want a generated ID, got empty string")
		}
		if created == nil {
			t.Fatal("expected repo.Create to be called")
		}
		if created.ID != got.ID {
			t.Errorf("want repo.Create called with returned user (ID %q), got ID %q", got.ID, created.ID)
		}
	})
}
