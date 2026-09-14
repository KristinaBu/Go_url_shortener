package handler

import (
	"context"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/KristinaBu/Go_url_shortener/internal/service"
)

type testRepository struct {
	byURL  map[string]domain.Link
	byCode map[string]domain.Link
}

type testCache struct {
	items map[string]domain.Link
}

func newTestCache() *testCache {
	return &testCache{
		items: make(map[string]domain.Link),
	}
}

func (c *testCache) Get(key string) (domain.Link, bool) {
	link, ok := c.items[key]
	return link, ok
}

func (c *testCache) Set(key string, link domain.Link) {
	c.items[key] = link
}

func newTestHandler() *Handler {
	repo := &testRepository{
		byURL:  make(map[string]domain.Link),
		byCode: make(map[string]domain.Link),
	}

	generator := &testGenerator{
		code: "abc123_XYZ",
	}

	cache := newTestCache()

	svc := service.NewLinkService(
		repo,
		generator,
		cache,
	)

	return New(svc)
}

type testGenerator struct {
	code string
}

func (g *testGenerator) Generate() (string, error) {
	return g.code, nil
}

func (r *testRepository) Create(_ context.Context, link domain.Link) error {
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

func (r *testRepository) FindByURL(
	_ context.Context,
	url string,
) (domain.Link, error) {
	link, exists := r.byURL[url]
	if !exists {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}

func (r *testRepository) FindByCode(
	_ context.Context,
	code string,
) (domain.Link, error) {
	link, exists := r.byCode[code]
	if !exists {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}
