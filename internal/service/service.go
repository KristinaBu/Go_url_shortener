package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"net/url"
)

type LinkRepository interface {
	FindByURL(ctx context.Context, originalURL string) (domain.Link, error)
	FindByCode(ctx context.Context, shortCode string) (domain.Link, error)
	Create(ctx context.Context, link domain.Link) error
}

type CodeGenerator interface {
	Generate() (string, error)
}

type LinkService struct {
	repository LinkRepository
	generator  CodeGenerator
}

func NewLinkService(
	repository LinkRepository,
	generator CodeGenerator,
) *LinkService {
	return &LinkService{
		repository: repository,
		generator:  generator,
	}
}

func (s *LinkService) Create(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	if err := validateURL(originalURL); err != nil {
		return domain.Link{}, err
	}

	existing, err := s.repository.FindByURL(ctx, originalURL)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Link{}, err
	}

	const (
		MaxURLLength          = 2048
		MaxGenerationAttempts = 10
	)

	for attempt := 0; attempt < MaxGenerationAttempts; attempt++ {
		shortCode, err := s.generator.Generate()
		if err != nil {
			return domain.Link{}, err
		}

		link := domain.Link{
			ShortCode:   shortCode,
			OriginalURL: originalURL,
		}

		err = s.repository.Create(ctx, link)
		if err == nil {
			return link, nil
		}

		switch {
		case errors.Is(err, domain.ErrCodeAlreadyExists):
			continue

		case errors.Is(err, domain.ErrURLAlreadyExists):
			existing, findErr := s.repository.FindByURL(ctx, originalURL)
			if findErr != nil {
				return domain.Link{}, findErr
			}

			return existing, nil

		default:
			return domain.Link{}, err
		}
	}

	return domain.Link{}, fmt.Errorf(
		"failed to generate unique short code after %d attempts",
		MaxGenerationAttempts,
	)
}

func (s *LinkService) Get(
	ctx context.Context,
	shortCode string,
) (domain.Link, error) {
	return s.repository.FindByCode(ctx, shortCode)
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return domain.ErrInvalidURL
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return domain.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return domain.ErrInvalidURL
	}

	if parsed.Host == "" {
		return domain.ErrInvalidURL
	}

	return nil
}
