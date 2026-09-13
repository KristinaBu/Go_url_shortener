package repository

import (
	"context"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"sync"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	byURL  map[string]domain.Link
	byCode map[string]domain.Link
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		byURL:  make(map[string]domain.Link),
		byCode: make(map[string]domain.Link),
	}
}

func (r *MemoryRepository) FindByURL(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.byURL[originalURL]
	if !ok {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}

func (r *MemoryRepository) FindByCode(
	ctx context.Context,
	shortCode string,
) (domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.byCode[shortCode]
	if !ok {
		return domain.Link{}, domain.ErrNotFound
	}

	return link, nil
}

func (r *MemoryRepository) Create(
	ctx context.Context,
	link domain.Link,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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
