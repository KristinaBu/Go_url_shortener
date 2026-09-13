package repository

import (
	"context"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
)

type LinkRepository interface {
	FindByURL(ctx context.Context, originalURL string) (domain.Link, error)
	FindByCode(ctx context.Context, shortCode string) (domain.Link, error)
	Create(ctx context.Context, link domain.Link) error
}
