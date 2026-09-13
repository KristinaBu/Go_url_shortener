package cache

import "github.com/KristinaBu/Go_url_shortener/internal/domain"

type LinkCache interface {
	Get(shortCode string) (domain.Link, bool)
	Set(shortCode string, link domain.Link)
}
