package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/ziliscite/video-to-mp3/auth/internal/domain"
)

var (
	ErrEditConflict   = errors.New("conflict")
	ErrRecordNotFound = errors.New("not found")
	ErrDuplicateEntry = errors.New("duplicate")
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Insert(ctx context.Context, user *domain.User) error
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (u userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, username, email, password_hash, created_at, updated_at, is_admin
        FROM users
        WHERE email = $1;
	`

	var user domain.User
	var hash []byte
	err := u.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email,
		&hash, &user.CreatedAt, &user.UpdatedAt,
		&user.IsAdmin,
	)

	user.Password.InsertHash(hash)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, fmt.Errorf("something's wrong: %w", err)
		}
	}

	return &user, nil
}

func (u userRepo) Insert(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (username, email, password_hash) 
        VALUES ($1, $2, $3)
        RETURNING id
	`

	args := []any{user.Username, user.Email, user.Hash()}

	if err := u.db.QueryRow(ctx, query, args...).Scan(&user.ID); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			return ErrDuplicateEntry
		default:
			return fmt.Errorf("something's wrong: %w", err)
		}
	}

	return nil
}
