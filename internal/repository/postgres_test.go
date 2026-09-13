package repository

import (
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
)

func TestMapPostgresError_DuplicateCode(t *testing.T) {
	err := &pgconn.PgError{
		ConstraintName: "links_pkey",
	}

	got := mapPostgresError(err)

	if !errors.Is(got, domain.ErrCodeAlreadyExists) {
		t.Fatalf(
			"mapPostgresError() = %v, want ErrCodeAlreadyExists",
			got,
		)
	}
}

func TestMapPostgresError_DuplicateURL(t *testing.T) {
	err := &pgconn.PgError{
		ConstraintName: "links_original_url_key",
	}

	got := mapPostgresError(err)

	if !errors.Is(got, domain.ErrURLAlreadyExists) {
		t.Fatalf(
			"mapPostgresError() = %v, want ErrURLAlreadyExists",
			got,
		)
	}
}

func TestMapPostgresError_UnknownError(t *testing.T) {
	err := errors.New("database unavailable")

	got := mapPostgresError(err)

	if !errors.Is(got, err) {
		t.Fatalf(
			"mapPostgresError() = %v, want original error",
			got,
		)
	}
}
