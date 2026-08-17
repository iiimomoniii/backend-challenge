package user

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/mocks"
)

func strPtr(s string) *string { return &s }

func TestUpdateUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("empty id returns ErrUserNotFound without calling repo", func(t *testing.T) {
		uc := NewUpdateUseCase(&mocks.UserRepository{})
		_, err := uc.Execute(ctx, "", domainuser.UpdateRequest{Name: strPtr("Ada")})
		if !errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})

	t.Run("blank name after trim returns ErrNameRequired", func(t *testing.T) {
		uc := NewUpdateUseCase(&mocks.UserRepository{})
		_, err := uc.Execute(ctx, "u1", domainuser.UpdateRequest{Name: strPtr("   ")})
		if !errors.Is(err, domainuser.ErrNameRequired) {
			t.Fatalf("want ErrNameRequired, got %v", err)
		}
	})

	t.Run("blank email after trim returns ErrUsernameRequired", func(t *testing.T) {
		uc := NewUpdateUseCase(&mocks.UserRepository{})
		_, err := uc.Execute(ctx, "u1", domainuser.UpdateRequest{Email: strPtr("   ")})
		if !errors.Is(err, domainuser.ErrUsernameRequired) {
			t.Fatalf("want ErrUsernameRequired, got %v", err)
		}
	})

	t.Run("email colliding with a different user returns ErrUsernameAlreadyExists", func(t *testing.T) {
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return &domainuser.User{ID: "someone-else", Email: email}, nil
			},
		}
		uc := NewUpdateUseCase(repo)
		_, err := uc.Execute(ctx, "u1", domainuser.UpdateRequest{Email: strPtr("taken@example.com")})
		if !errors.Is(err, domainuser.ErrUsernameAlreadyExists) {
			t.Fatalf("want ErrUsernameAlreadyExists, got %v", err)
		}
	})

	t.Run("email matching the same user's own record is allowed (no false positive)", func(t *testing.T) {
		var updateCalled bool
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return &domainuser.User{ID: "u1", Email: email}, nil
			},
			UpdateFunc: func(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
				updateCalled = true
				return &domainuser.User{ID: id, Email: *req.Email}, nil
			},
		}
		uc := NewUpdateUseCase(repo)
		_, err := uc.Execute(ctx, "u1", domainuser.UpdateRequest{Email: strPtr("ada@example.com")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !updateCalled {
			t.Fatal("expected repo.Update to be called")
		}
	})

	t.Run("trims name and email before passing to repository", func(t *testing.T) {
		var gotReq domainuser.UpdateRequest
		repo := &mocks.UserRepository{
			SearchByEmailFunc: func(ctx context.Context, email string) (*domainuser.User, error) {
				return nil, domainuser.ErrUserNotFound
			},
			UpdateFunc: func(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
				gotReq = req
				return &domainuser.User{ID: id, Name: *req.Name, Email: *req.Email}, nil
			},
		}
		uc := NewUpdateUseCase(repo)
		_, err := uc.Execute(ctx, "u1", domainuser.UpdateRequest{
			Name:  strPtr("  Ada  "),
			Email: strPtr("  ada@example.com  "),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *gotReq.Name != "Ada" {
			t.Errorf("want trimmed name %q, got %q", "Ada", *gotReq.Name)
		}
		if *gotReq.Email != "ada@example.com" {
			t.Errorf("want trimmed email %q, got %q", "ada@example.com", *gotReq.Email)
		}
	})

	t.Run("partial update leaves unset fields nil", func(t *testing.T) {
		var gotReq domainuser.UpdateRequest
		repo := &mocks.UserRepository{
			UpdateFunc: func(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
				gotReq = req
				return &domainuser.User{ID: id}, nil
			},
		}
		uc := NewUpdateUseCase(repo)
		_, err := uc.Execute(ctx, "u1", domainuser.UpdateRequest{Name: strPtr("Ada")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotReq.Email != nil {
			t.Errorf("want Email to remain nil, got %q", *gotReq.Email)
		}
	})

	t.Run("not found from repository is propagated", func(t *testing.T) {
		repo := &mocks.UserRepository{
			UpdateFunc: func(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
				return nil, domainuser.ErrUserNotFound
			},
		}
		uc := NewUpdateUseCase(repo)
		_, err := uc.Execute(ctx, "missing-id", domainuser.UpdateRequest{Name: strPtr("Ada")})
		if !errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})
}
