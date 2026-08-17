package user

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/mocks"
)

func TestDeleteUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("empty id returns ErrUserNotFound without calling repo", func(t *testing.T) {
		uc := NewDeleteUseCase(&mocks.UserRepository{})
		err := uc.Execute(ctx, "")
		if !errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})

	t.Run("not found is propagated from repository", func(t *testing.T) {
		repo := &mocks.UserRepository{
			DeleteFunc: func(ctx context.Context, id string) error {
				return domainuser.ErrUserNotFound
			},
		}
		uc := NewDeleteUseCase(repo)
		err := uc.Execute(ctx, "missing-id")
		if !errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		wantErr := errors.New("connection refused")
		repo := &mocks.UserRepository{
			DeleteFunc: func(ctx context.Context, id string) error {
				return wantErr
			},
		}
		uc := NewDeleteUseCase(repo)
		err := uc.Execute(ctx, "u1")
		if !errors.Is(err, wantErr) {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("happy path deletes the user", func(t *testing.T) {
		var gotID string
		repo := &mocks.UserRepository{
			DeleteFunc: func(ctx context.Context, id string) error {
				gotID = id
				return nil
			},
		}
		uc := NewDeleteUseCase(repo)
		err := uc.Execute(ctx, "u1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotID != "u1" {
			t.Errorf("want repo.Delete called with %q, got %q", "u1", gotID)
		}
	})
}
