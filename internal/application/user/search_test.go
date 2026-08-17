package user

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/internal/mocks"
)

func TestSearchUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("empty id returns ErrUserNotFound without calling repo", func(t *testing.T) {
		uc := NewSearchUseCase(&mocks.UserRepository{}) // no SearchByIDFunc set — would panic if called
		_, err := uc.Execute(ctx, "")
		if !errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})

	t.Run("not found is propagated from repository", func(t *testing.T) {
		repo := &mocks.UserRepository{
			SearchByIDFunc: func(ctx context.Context, id string) (*domainuser.User, error) {
				return nil, domainuser.ErrUserNotFound
			},
		}
		uc := NewSearchUseCase(repo)
		_, err := uc.Execute(ctx, "missing-id")
		if !errors.Is(err, domainuser.ErrUserNotFound) {
			t.Fatalf("want ErrUserNotFound, got %v", err)
		}
	})

	t.Run("happy path returns the user", func(t *testing.T) {
		want := &domainuser.User{ID: "u1", Name: "Ada"}
		repo := &mocks.UserRepository{
			SearchByIDFunc: func(ctx context.Context, id string) (*domainuser.User, error) {
				if id != "u1" {
					t.Errorf("want SearchByID called with %q, got %q", "u1", id)
				}
				return want, nil
			},
		}
		uc := NewSearchUseCase(repo)
		got, err := uc.Execute(ctx, "u1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("want %+v, got %+v", want, got)
		}
	})
}

func TestSearchAllUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("repository error is propagated", func(t *testing.T) {
		wantErr := errors.New("connection refused")
		repo := &mocks.UserRepository{
			ListFunc: func(ctx context.Context) ([]*domainuser.User, error) {
				return nil, wantErr
			},
		}
		uc := NewSearchAllUseCase(repo)
		_, err := uc.Execute(ctx)
		if !errors.Is(err, wantErr) {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("nil result from repository is normalized to an empty, non-nil slice", func(t *testing.T) {
		repo := &mocks.UserRepository{
			ListFunc: func(ctx context.Context) ([]*domainuser.User, error) {
				return nil, nil
			},
		}
		uc := NewSearchAllUseCase(repo)
		got, err := uc.Execute(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("want non-nil slice, got nil")
		}
		if len(got) != 0 {
			t.Errorf("want empty slice, got %d items", len(got))
		}
	})

	t.Run("happy path returns all users", func(t *testing.T) {
		want := []*domainuser.User{
			{ID: "u1", Name: "Ada"},
			{ID: "u2", Name: "Grace"},
		}
		repo := &mocks.UserRepository{
			ListFunc: func(ctx context.Context) ([]*domainuser.User, error) {
				return want, nil
			},
		}
		uc := NewSearchAllUseCase(repo)
		got, err := uc.Execute(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(want) {
			t.Fatalf("want %d users, got %d", len(want), len(got))
		}
	})
}
