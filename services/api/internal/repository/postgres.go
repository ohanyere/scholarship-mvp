package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresScholarshipRepository struct {
	db *sql.DB
}

func NewPostgresScholarshipRepository(db *sql.DB) *PostgresScholarshipRepository {
	return &PostgresScholarshipRepository{db: db}
}

func (r *PostgresScholarshipRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresScholarshipRepository) CreateScholarship(ctx context.Context, params CreateScholarshipParams) (Scholarship, error) {
	// Insert a new scholarship row and return the stored record so the API can
	// respond with database-assigned fields like id and timestamps.
	const query = `
		INSERT INTO scholarships (
			title,
			provider,
			description,
			amount,
			currency,
			deadline_at,
			eligibility,
			application_url,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id,
			title,
			provider,
			description,
			amount,
			currency,
			deadline_at,
			eligibility,
			application_url,
			status,
			created_at,
			updated_at
	`

	var scholarship Scholarship
	if err := r.db.QueryRowContext(
		ctx,
		query,
		params.Title,
		params.Provider,
		params.Description,
		params.Amount,
		params.Currency,
		params.DeadlineAt,
		params.Eligibility,
		params.ApplicationURL,
		params.Status,
	).Scan(
		&scholarship.ID,
		&scholarship.Title,
		&scholarship.Provider,
		&scholarship.Description,
		&scholarship.Amount,
		&scholarship.Currency,
		&scholarship.DeadlineAt,
		&scholarship.Eligibility,
		&scholarship.ApplicationURL,
		&scholarship.Status,
		&scholarship.CreatedAt,
		&scholarship.UpdatedAt,
	); err != nil {
		return Scholarship{}, fmt.Errorf("create scholarship: %w", err)
	}

	return scholarship, nil
}

func (r *PostgresScholarshipRepository) ListPublishedScholarships(ctx context.Context) ([]Scholarship, error) {
	// Public list queries only return published scholarships, ordered so the
	// nearest deadlines appear first.
	const query = `
		SELECT
			id,
			title,
			provider,
			description,
			amount,
			currency,
			deadline_at,
			eligibility,
			application_url,
			status,
			created_at,
			updated_at
		FROM scholarships
		WHERE status = 'published'
		ORDER BY deadline_at ASC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list scholarships: %w", err)
	}
	defer rows.Close()

	var scholarships []Scholarship
	for rows.Next() {
		var scholarship Scholarship
		if err := rows.Scan(
			&scholarship.ID,
			&scholarship.Title,
			&scholarship.Provider,
			&scholarship.Description,
			&scholarship.Amount,
			&scholarship.Currency,
			&scholarship.DeadlineAt,
			&scholarship.Eligibility,
			&scholarship.ApplicationURL,
			&scholarship.Status,
			&scholarship.CreatedAt,
			&scholarship.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan scholarship row: %w", err)
		}

		scholarships = append(scholarships, scholarship)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scholarship rows: %w", err)
	}

	return scholarships, nil
}

func (r *PostgresScholarshipRepository) GetPublishedScholarshipByID(ctx context.Context, id string) (Scholarship, error) {
	// Public detail queries only return a scholarship when the id exists and
	// the record is published.
	const query = `
		SELECT
			id,
			title,
			provider,
			description,
			amount,
			currency,
			deadline_at,
			eligibility,
			application_url,
			status,
			created_at,
			updated_at
		FROM scholarships
		WHERE id = $1 AND status = 'published'
	`

	var scholarship Scholarship
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&scholarship.ID,
		&scholarship.Title,
		&scholarship.Provider,
		&scholarship.Description,
		&scholarship.Amount,
		&scholarship.Currency,
		&scholarship.DeadlineAt,
		&scholarship.Eligibility,
		&scholarship.ApplicationURL,
		&scholarship.Status,
		&scholarship.CreatedAt,
		&scholarship.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Scholarship{}, ErrScholarshipNotFound
		}

		return Scholarship{}, fmt.Errorf("get scholarship by id: %w", err)
	}

	return scholarship, nil
}
