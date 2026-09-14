package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/KristinaBu/Go_url_shortener/internal/repository"
	"github.com/KristinaBu/Go_url_shortener/pkg/cache"
	"github.com/KristinaBu/Go_url_shortener/pkg/generator"
)

const maxGenerationAttempts = 10

type LinkService struct {
	repository repository.LinkRepository
	generator  CodeGenerator
	cache      cache.Cache[string, domain.Link]
}

type CodeGenerator interface {
	Generate() (string, error)
}

func NewLinkService(
	repository repository.LinkRepository,
	cache cache.Cache[string, domain.Link],
) *LinkService {
	return &LinkService{
		repository: repository,
		generator:  generator.New(),
		cache:      cache,
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

	for range maxGenerationAttempts {
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
			s.cache.Set(link.ShortCode, link)
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

			s.cache.Set(existing.ShortCode, existing)
			return existing, nil

		default:
			return domain.Link{}, err
		}
	}

	return domain.Link{}, fmt.Errorf(
		"failed to generate unique short code after %d attempts",
		maxGenerationAttempts,
	)
}

func (s *LinkService) Get(
	ctx context.Context,
	shortCode string,
) (domain.Link, error) {
	if link, ok := s.cache.Get(shortCode); ok {
		return link, nil
	}

	link, err := s.repository.FindByCode(ctx, shortCode)
	if err != nil {
		return domain.Link{}, err
	}

	s.cache.Set(shortCode, link)

	return link, nil
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
