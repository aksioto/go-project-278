package link

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"code/internal/domain/link"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=link Service

type ListResult struct {
	Links []link.Link
	Total int64
}

type Service interface {
	CreateLink(ctx context.Context, originalURL, shortName string) (*link.Link, error)
	GetLink(ctx context.Context, id int64) (*link.Link, error)
	ListLinks(ctx context.Context) ([]link.Link, error)
	ListLinksPaginated(ctx context.Context, limit, offset int32) (*ListResult, error)
	CountLinks(ctx context.Context) (int64, error)
	UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (*link.Link, error)
	DeleteLink(ctx context.Context, id int64) error
}

type service struct {
	repo link.Repository
}

var _ Service = (*service)(nil)

func NewService(repo link.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateLink(ctx context.Context, originalURL, shortName string) (*link.Link, error) {
	if shortName == "" {
		generated, err := s.generateShortName()
		if err != nil {
			return nil, fmt.Errorf("generate short name: %w", err)
		}
		shortName = generated
	}

	l, err := s.repo.Create(ctx, originalURL, shortName)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (s *service) GetLink(ctx context.Context, id int64) (*link.Link, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListLinks(ctx context.Context) ([]link.Link, error) {
	return s.repo.List(ctx)
}

func (s *service) ListLinksPaginated(ctx context.Context, limit, offset int32) (*ListResult, error) {
	links, err := s.repo.ListPaginated(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Links: links,
		Total: total,
	}, nil
}

func (s *service) UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (*link.Link, error) {
	return s.repo.Update(ctx, id, originalURL, shortName)
}

func (s *service) DeleteLink(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) CountLinks(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
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
