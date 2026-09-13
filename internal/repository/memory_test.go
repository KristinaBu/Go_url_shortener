package repository

import (
	"context"
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"sync"
	"testing"
)

func TestMemoryRepository_CreateAndFind(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	link := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://example.com",
	}

	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	gotByURL, err := repo.FindByURL(ctx, link.OriginalURL)
	if err != nil {
		t.Fatalf("FindByURL() error = %v", err)
	}

	if gotByURL != link {
		t.Fatalf("FindByURL() = %+v, want %+v", gotByURL, link)
	}

	gotByCode, err := repo.FindByCode(ctx, link.ShortCode)
	if err != nil {
		t.Fatalf("FindByCode() error = %v", err)
	}

	if gotByCode != link {
		t.Fatalf("FindByCode() = %+v, want %+v", gotByCode, link)
	}
}

func TestMemoryRepository_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_, err := repo.FindByURL(ctx, "https://example.com")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("FindByURL() error = %v, want ErrNotFound", err)
	}

	_, err = repo.FindByCode(ctx, "abc123_XYZ")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("FindByCode() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryRepository_DuplicateURL(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	first := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://example.com",
	}

	second := domain.Link{
		ShortCode:   "different1",
		OriginalURL: "https://example.com",
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	if err := repo.Create(ctx, second); !errors.Is(err, domain.ErrURLAlreadyExists) {
		t.Fatalf(
			"second Create() error = %v, want ErrURLAlreadyExists",
			err,
		)
	}
}

func TestMemoryRepository_DuplicateCode(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	first := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://example.com",
	}

	second := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://google.com",
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	if err := repo.Create(ctx, second); !errors.Is(err, domain.ErrCodeAlreadyExists) {
		t.Fatalf(
			"second Create() error = %v, want ErrCodeAlreadyExists",
			err,
		)
	}
}

func TestMemoryRepository_ConcurrentCreate(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	const requests = 100

	var wg sync.WaitGroup
	wg.Add(requests)

	successes := make(chan struct{}, requests)

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()

			link := domain.Link{
				ShortCode:   "abc123_XYZ",
				OriginalURL: "https://example.com",
			}

			if err := repo.Create(ctx, link); err == nil {
				successes <- struct{}{}
			}
		}()
	}

	wg.Wait()
	close(successes)

	if got := len(successes); got != 1 {
		t.Fatalf("successful creates = %d, want 1", got)
	}
}
