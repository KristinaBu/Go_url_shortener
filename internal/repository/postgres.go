package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) FindByURL(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	const query = `
		SELECT short_code, original_url
		FROM links
		WHERE original_url = $1
	`

	var link domain.Link

	err := r.db.QueryRowContext(
		ctx,
		query,
		originalURL,
	).Scan(
		&link.ShortCode,
		&link.OriginalURL,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Link{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Link{}, err
	}

	return link, nil
}

func (r *PostgresRepository) FindByCode(
	ctx context.Context,
	shortCode string,
) (domain.Link, error) {
	const query = `
		SELECT short_code, original_url
		FROM links
		WHERE short_code = $1
	`

	var link domain.Link

	err := r.db.QueryRowContext(
		ctx,
		query,
		shortCode,
	).Scan(
		&link.ShortCode,
		&link.OriginalURL,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Link{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Link{}, err
	}

	return link, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	link domain.Link,
) error {
	const query = `
		INSERT INTO links (short_code, original_url)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		link.ShortCode,
		link.OriginalURL,
	)

	if err != nil {
		return mapPostgresError(err)
	}

	return nil
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.ConstraintName {
	case "links_pkey":
		return domain.ErrCodeAlreadyExists

	case "links_original_url_key":
		return domain.ErrURLAlreadyExists

	default:
		return err
	}
}
