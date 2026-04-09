package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresApplicationRepository struct {
	db *sql.DB
}

func NewPostgresApplicationRepository(db *sql.DB) *PostgresApplicationRepository {
	return &PostgresApplicationRepository{db: db}
}

func (r *PostgresApplicationRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresApplicationRepository) CreateApplication(ctx context.Context, params CreateApplicationParams) (Application, error) {
	const query = `
		INSERT INTO applications (
			user_id,
			scholarship_id,
			status,
			notes
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			user_id,
			scholarship_id,
			status,
			notes,
			created_at,
			updated_at
	`

	var application Application
	if err := r.db.QueryRowContext(
		ctx,
		query,
		params.UserID,
		params.ScholarshipID,
		params.Status,
		params.Notes,
	).Scan(
		&application.ID,
		&application.UserID,
		&application.ScholarshipID,
		&application.Status,
		&application.Notes,
		&application.CreatedAt,
		&application.UpdatedAt,
	); err != nil {
		if isUniqueViolation(err) {
			return Application{}, ErrApplicationAlreadyExists
		}

		if isForeignKeyViolation(err) {
			return Application{}, ErrScholarshipNotFound
		}

		return Application{}, fmt.Errorf("create application: %w", err)
	}

	return application, nil
}

func (r *PostgresApplicationRepository) ListApplications(ctx context.Context, userID string) ([]Application, error) {
	const query = `
		SELECT
			id,
			user_id,
			scholarship_id,
			status,
			notes,
			created_at,
			updated_at
		FROM applications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var application Application
		if err := rows.Scan(
			&application.ID,
			&application.UserID,
			&application.ScholarshipID,
			&application.Status,
			&application.Notes,
			&application.CreatedAt,
			&application.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan application row: %w", err)
		}

		applications = append(applications, application)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate application rows: %w", err)
	}

	return applications, nil
}

func (r *PostgresApplicationRepository) UpdateApplication(ctx context.Context, params UpdateApplicationParams) (Application, error) {
	const query = `
		UPDATE applications
		SET
			status = COALESCE($3, status),
			notes = COALESCE($4, notes),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING
			id,
			user_id,
			scholarship_id,
			status,
			notes,
			created_at,
			updated_at
	`

	var application Application
	if err := r.db.QueryRowContext(
		ctx,
		query,
		params.ID,
		params.UserID,
		params.Status,
		params.Notes,
	).Scan(
		&application.ID,
		&application.UserID,
		&application.ScholarshipID,
		&application.Status,
		&application.Notes,
		&application.CreatedAt,
		&application.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Application{}, ErrApplicationNotFound
		}

		return Application{}, fmt.Errorf("update application: %w", err)
	}

	return application, nil
}
