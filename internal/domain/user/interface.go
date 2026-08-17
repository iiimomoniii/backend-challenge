package user

import "context"

// interface
type Repository interface {
	Create(ctx context.Context, u *User) error
	SearchByID(ctx context.Context, id string) (*User, error)
	SearchByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*User, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}
