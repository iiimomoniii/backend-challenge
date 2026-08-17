package mocks

import (
	"context"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

type UserRepository struct {
	CreateFunc        func(ctx context.Context, u *domainuser.User) error
	SearchByIDFunc    func(ctx context.Context, id string) (*domainuser.User, error)
	SearchByEmailFunc func(ctx context.Context, email string) (*domainuser.User, error)
	ListFunc          func(ctx context.Context) ([]*domainuser.User, error)
	UpdateFunc        func(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error)
	DeleteFunc        func(ctx context.Context, id string) error
	CountFunc         func(ctx context.Context) (int64, error)
}

var _ domainuser.Repository = (*UserRepository)(nil)

func (m *UserRepository) Create(ctx context.Context, u *domainuser.User) error {
	if m.CreateFunc == nil {
		panic("mocks.UserRepository: CreateFunc not set")
	}
	return m.CreateFunc(ctx, u)
}

func (m *UserRepository) SearchByID(ctx context.Context, id string) (*domainuser.User, error) {
	if m.SearchByIDFunc == nil {
		panic("mocks.UserRepository: SearchByIDFunc not set")
	}
	return m.SearchByIDFunc(ctx, id)
}

func (m *UserRepository) SearchByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	if m.SearchByEmailFunc == nil {
		panic("mocks.UserRepository: SearchByEmailFunc not set")
	}
	return m.SearchByEmailFunc(ctx, email)
}

func (m *UserRepository) List(ctx context.Context) ([]*domainuser.User, error) {
	if m.ListFunc == nil {
		panic("mocks.UserRepository: ListFunc not set")
	}
	return m.ListFunc(ctx)
}

func (m *UserRepository) Update(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
	if m.UpdateFunc == nil {
		panic("mocks.UserRepository: UpdateFunc not set")
	}
	return m.UpdateFunc(ctx, id, req)
}

func (m *UserRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc == nil {
		panic("mocks.UserRepository: DeleteFunc not set")
	}
	return m.DeleteFunc(ctx, id)
}

func (m *UserRepository) Count(ctx context.Context) (int64, error) {
	if m.CountFunc == nil {
		panic("mocks.UserRepository: CountFunc not set")
	}
	return m.CountFunc(ctx)
}
