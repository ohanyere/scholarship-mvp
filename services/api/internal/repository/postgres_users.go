package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	const query = `
		INSERT INTO users (
			email,
			password_hash,
			full_name,
			role
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			email,
			password_hash,
			full_name,
			role,
			created_at,
			updated_at
	`

	var user User
	if err := r.db.QueryRowContext(
		ctx,
		query,
		params.Email,
		params.PasswordHash,
		params.FullName,
		params.Role,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrUserEmailTaken
		}

		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			full_name,
			role,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			full_name,
			role,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func isForeignKeyViolation(err error) bool {
	return strings.Contains(err.Error(), "violates foreign key constraint")
}
