package service

import (
	"context"
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/cache"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"testing"
)

type fakeRepository struct {
	byURL  map[string]domain.Link
	byCode map[string]domain.Link
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		byURL:  make(map[string]domain.Link),
		byCode: make(map[string]domain.Link),
	}
}

func (r *fakeRepository) FindByURL(
	_ context.Context,
	originalURL string,
) (domain.Link, error) {
	link, ok := r.byURL[originalURL]
	if !ok {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}

func (r *fakeRepository) FindByCode(
	_ context.Context,
	shortCode string,
) (domain.Link, error) {
	link, ok := r.byCode[shortCode]
	if !ok {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}

func (r *fakeRepository) Create(
	_ context.Context,
	link domain.Link,
) error {
	if _, exists := r.byURL[link.OriginalURL]; exists {
		return domain.ErrURLAlreadyExists
	}

	if _, exists := r.byCode[link.ShortCode]; exists {
		return domain.ErrCodeAlreadyExists
	}

	r.byURL[link.OriginalURL] = link
	r.byCode[link.ShortCode] = link

	return nil
}

type fakeGenerator struct {
	codes []string
	index int
}

func (g *fakeGenerator) Generate() (string, error) {
	if g.index >= len(g.codes) {
		return "", errors.New("no more codes")
	}

	code := g.codes[g.index]
	g.index++

	return code, nil
}

func newFakeCache() cache.LinkCache {
	c, _ := cache.NewLRU(100)
	return c
}

func TestLinkService_Create(t *testing.T) {
	repo := newFakeRepository()

	generator := &fakeGenerator{
		codes: []string{"abc123_XYZ"},
	}

	service := NewLinkService(
		repo,
		generator,
		newFakeCache(),
	)

	link, err := service.Create(
		context.Background(),
		"https://example.com",
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if link.ShortCode != "abc123_XYZ" {
		t.Fatalf("ShortCode = %q, want %q", link.ShortCode, "abc123_XYZ")
	}

	if link.OriginalURL != "https://example.com" {
		t.Fatalf(
			"OriginalURL = %q, want %q",
			link.OriginalURL,
			"https://example.com",
		)
	}
}

func TestLinkService_CreateSameURLReturnsExistingLink(t *testing.T) {
	repo := newFakeRepository()

	existing := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://example.com",
	}

	repo.byURL[existing.OriginalURL] = existing
	repo.byCode[existing.ShortCode] = existing

	generator := &fakeGenerator{
		codes: []string{"should_not_be_used"},
	}

	service := NewLinkService(
		repo,
		generator,
		newFakeCache(),
	)

	link, err := service.Create(
		context.Background(),
		existing.OriginalURL,
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if link != existing {
		t.Fatalf("Create() = %+v, want %+v", link, existing)
	}

	if generator.index != 0 {
		t.Fatalf("generator calls = %d, want 0", generator.index)
	}
}

func TestLinkService_CreateRetriesOnCodeCollision(t *testing.T) {
	repo := newFakeRepository()

	existing := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://other.example",
	}

	repo.byCode[existing.ShortCode] = existing

	generator := &fakeGenerator{
		codes: []string{
			"abc123_XYZ",
			"newCode_12",
		},
	}

	service := NewLinkService(
		repo,
		generator,
		newFakeCache(),
	)

	link, err := service.Create(
		context.Background(),
		"https://example.com",
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if link.ShortCode != "newCode_12" {
		t.Fatalf(
			"ShortCode = %q, want %q",
			link.ShortCode,
			"newCode_12",
		)
	}

	if generator.index != 2 {
		t.Fatalf("generator calls = %d, want 2", generator.index)
	}
}

func TestLinkService_Get(t *testing.T) {
	repo := newFakeRepository()

	link := domain.Link{
		ShortCode:   "abc123_XYZ",
		OriginalURL: "https://example.com",
	}

	repo.byCode[link.ShortCode] = link

	service := NewLinkService(
		repo,
		&fakeGenerator{},
		newFakeCache(),
	)

	got, err := service.Get(
		context.Background(),
		link.ShortCode,
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != link {
		t.Fatalf("Get() = %+v, want %+v", got, link)
	}
}
