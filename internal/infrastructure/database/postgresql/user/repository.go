package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// Repository is the PostgreSQL adapter for domainuser.Repository.
type Repository struct {
	db *sql.DB
}

var _ domainuser.Repository = (*Repository)(nil)

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// EnsureSchema creates the users table if it does not already exist.
func (r *Repository) EnsureSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)
	`)
	return err
}

func scanUser(row interface{ Scan(...any) error }) (*domainuser.User, error) {
	u := &domainuser.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainuser.ErrUserNotFound
	}
	return u, err
}

func (r *Repository) Create(ctx context.Context, u *domainuser.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password, created_at) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Name, u.Email, u.Password, u.CreatedAt,
	)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return domainuser.ErrUsernameAlreadyExists
	}
	return err
}

const userColumns = `id::text, name, email, password, created_at`

func (r *Repository) SearchByID(ctx context.Context, id string) (*domainuser.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (r *Repository) SearchByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
}

func (r *Repository) List(ctx context.Context) ([]*domainuser.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userColumns+` FROM users ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*domainuser.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
	if req.Name == nil && req.Email == nil {
		return r.SearchByID(ctx, id)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET name = COALESCE($2, name), email = COALESCE($3, email)
		WHERE id = $1
	`, id, req.Name, req.Email)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domainuser.ErrUsernameAlreadyExists
		}
		return nil, err
	}
	return r.SearchByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domainuser.ErrUserNotFound
	}
	return nil
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
