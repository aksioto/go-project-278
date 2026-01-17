package visit

import (
	"context"
	"log/slog"
	"time"

	"code/internal/domain/link"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=visit Service

const defaultDBTimeout = 5 * time.Second

type ListResult struct {
	Visits []link.Visit
	Total  int64
}

type Service interface {
	CreateVisit(ctx context.Context, visit link.Visit) (*link.Visit, error)
	ListVisitsPaginated(ctx context.Context, limit, offset int32) (*ListResult, error)
	CountVisits(ctx context.Context) (int64, error)
	DeleteVisit(ctx context.Context, id int64) error
}

type service struct {
	visitRepo link.VisitRepository
	logger    *slog.Logger
}

var _ Service = (*service)(nil)

func NewService(visitRepo link.VisitRepository, logger *slog.Logger) *service {
	return &service{
		visitRepo: visitRepo,
		logger:    logger,
	}
}

func (s *service) CreateVisit(ctx context.Context, visit link.Visit) (*link.Visit, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.visitRepo.Create(dbCtx, visit)
}

func (s *service) ListVisitsPaginated(ctx context.Context, limit, offset int32) (*ListResult, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()

	visits, err := s.visitRepo.ListPaginated(dbCtx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.visitRepo.Count(dbCtx)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Visits: visits,
		Total:  total,
	}, nil
}

func (s *service) CountVisits(ctx context.Context) (int64, error) {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.visitRepo.Count(dbCtx)
}

func (s *service) DeleteVisit(ctx context.Context, id int64) error {
	dbCtx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	return s.visitRepo.Delete(dbCtx, id)
}
