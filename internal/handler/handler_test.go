package handler

import (
	"context"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/KristinaBu/Go_url_shortener/internal/service"
)

type fakeRepository struct {
	byURL  map[string]domain.Link
	byCode map[string]domain.Link
}

type fakeCache struct {
	items map[string]domain.Link
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		items: make(map[string]domain.Link),
	}
}

func (c *fakeCache) Get(key string) (domain.Link, bool) {
	link, ok := c.items[key]
	return link, ok
}

func (c *fakeCache) Set(key string, link domain.Link) {
	c.items[key] = link
}

func newTestHandler() *Handler {
	repo := &fakeRepository{
		byURL:  make(map[string]domain.Link),
		byCode: make(map[string]domain.Link),
	}

	generator := &fakeGenerator{
		code: "abc123_XYZ",
	}

	cache := newFakeCache()

	svc := service.NewLinkService(
		repo,
		generator,
		cache,
	)

	return New(svc)
}

type fakeGenerator struct {
	code string
}

func (g *fakeGenerator) Generate() (string, error) {
	return g.code, nil
}

func (r *fakeRepository) Create(_ context.Context, link domain.Link) error {
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

func (r *fakeRepository) FindByURL(
	_ context.Context,
	url string,
) (domain.Link, error) {
	link, exists := r.byURL[url]
	if !exists {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}

func (r *fakeRepository) FindByCode(
	_ context.Context,
	code string,
) (domain.Link, error) {
	link, exists := r.byCode[code]
	if !exists {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}
