package link

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"code/internal/domain/link"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=link Service

const (
	defaultDBTimeout    = 5 * time.Second
	maxShortNameRetries = 3
	shortNameLength     = 8
)

type ListResult struct {
	Links []link.Link
	Total int64
}

type Service interface {
	CreateLink(ctx context.Context, originalURL, shortName string) (*link.Link, error)
	GetLink(ctx context.Context, id int64) (*link.Link, error)
	GetLinkByShortName(ctx context.Context, shortName string) (*link.Link, error)
	ListLinks(ctx context.Context) ([]link.Link, error)
	ListLinksPaginated(ctx context.Context, limit, offset int32) (*ListResult, error)
	CountLinks(ctx context.Context) (int64, error)
	UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (*link.Link, error)
	DeleteLink(ctx context.Context, id int64) error
}

type service struct {
	linkRepo link.LinkRepository
	logger   *slog.Logger
}

var _ Service = (*service)(nil)

func NewService(linkRepo link.LinkRepository, logger *slog.Logger) *service {
	return &service{
		linkRepo: linkRepo,
		logger:   logger,
	}
}

func (s *service) CreateLink(ctx context.Context, originalURL, shortName string) (*link.Link, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()

	if shortName != "" {
		return s.linkRepo.Create(dbCtx, originalURL, shortName)
	}

	for i := 0; i < maxShortNameRetries; i++ {
		generated, err := s.generateShortName()
		if err != nil {
			return nil, fmt.Errorf("generate short name: %w", err)
		}

		l, err := s.linkRepo.Create(dbCtx, originalURL, generated)
		if err == nil {
			return l, nil
		}

		if !errors.Is(err, link.ErrShortNameTaken) {
			return nil, err
		}

		s.logger.Warn("short name collision, retrying",
			slog.String("short_name", generated),
			slog.Int("attempt", i+1),
		)
	}

	return nil, fmt.Errorf("failed to generate unique short name after %d attempts", maxShortNameRetries)
}

func (s *service) GetLink(ctx context.Context, id int64) (*link.Link, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.linkRepo.GetByID(dbCtx, id)
}

func (s *service) GetLinkByShortName(ctx context.Context, shortName string) (*link.Link, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.linkRepo.GetByShortName(dbCtx, shortName)
}

func (s *service) ListLinks(ctx context.Context) ([]link.Link, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.linkRepo.List(dbCtx)
}

func (s *service) ListLinksPaginated(ctx context.Context, limit, offset int32) (*ListResult, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()

	links, err := s.linkRepo.ListPaginated(dbCtx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.linkRepo.Count(dbCtx)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Links: links,
		Total: total,
	}, nil
}

func (s *service) UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (*link.Link, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.linkRepo.Update(dbCtx, id, originalURL, shortName)
}

func (s *service) DeleteLink(ctx context.Context, id int64) error {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.linkRepo.Delete(dbCtx, id)
}

func (s *service) CountLinks(ctx context.Context) (int64, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.linkRepo.Count(dbCtx)
}

func (s *service) generateShortName() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	encoded = strings.TrimRight(encoded, "=")
	if len(encoded) > 8 {
		encoded = encoded[:8]
	}
	return encoded, nil
}
