package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresBookmarkRepository struct {
	db *sql.DB
}

func NewPostgresBookmarkRepository(db *sql.DB) *PostgresBookmarkRepository {
	return &PostgresBookmarkRepository{db: db}
}

func (r *PostgresBookmarkRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresBookmarkRepository) CreateBookmark(ctx context.Context, userID, scholarshipID string) (Bookmark, error) {
	const insertQuery = `
		INSERT INTO bookmarks (
			user_id,
			scholarship_id
		)
		VALUES ($1, $2)
		ON CONFLICT (user_id, scholarship_id) DO NOTHING
		RETURNING
			user_id,
			scholarship_id,
			created_at
	`

	var bookmark Bookmark
	err := r.db.QueryRowContext(ctx, insertQuery, userID, scholarshipID).Scan(
		&bookmark.UserID,
		&bookmark.ScholarshipID,
		&bookmark.CreatedAt,
	)
	if err == nil {
		return bookmark, nil
	}

	if isForeignKeyViolation(err) {
		return Bookmark{}, ErrScholarshipNotFound
	}

	if err != sql.ErrNoRows {
		return Bookmark{}, fmt.Errorf("create bookmark: %w", err)
	}

	const selectQuery = `
		SELECT
			user_id,
			scholarship_id,
			created_at
		FROM bookmarks
		WHERE user_id = $1 AND scholarship_id = $2
	`

	if err := r.db.QueryRowContext(ctx, selectQuery, userID, scholarshipID).Scan(
		&bookmark.UserID,
		&bookmark.ScholarshipID,
		&bookmark.CreatedAt,
	); err != nil {
		return Bookmark{}, fmt.Errorf("select existing bookmark: %w", err)
	}

	return bookmark, nil
}

func (r *PostgresBookmarkRepository) ListBookmarks(ctx context.Context, userID string) ([]Bookmark, error) {
	const query = `
		SELECT
			user_id,
			scholarship_id,
			created_at
		FROM bookmarks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var bookmark Bookmark
		if err := rows.Scan(
			&bookmark.UserID,
			&bookmark.ScholarshipID,
			&bookmark.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan bookmark row: %w", err)
		}

		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookmark rows: %w", err)
	}

	return bookmarks, nil
}

func (r *PostgresBookmarkRepository) DeleteBookmark(ctx context.Context, userID, scholarshipID string) error {
	const query = `
		DELETE FROM bookmarks
		WHERE user_id = $1 AND scholarship_id = $2
	`

	if _, err := r.db.ExecContext(ctx, query, userID, scholarshipID); err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}

	return nil
}
