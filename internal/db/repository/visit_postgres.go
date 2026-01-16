package repository

import (
	"context"
	"fmt"

	"code/internal/db/sqlc"
	"code/internal/domain/link"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VisitPostgres struct {
	queries *sqlc.Queries
}

var _ link.VisitRepository = (*VisitPostgres)(nil)

func NewVisitPostgres(pool *pgxpool.Pool) *VisitPostgres {
	return &VisitPostgres{
		queries: sqlc.New(pool),
	}
}

func toVisit(row sqlc.LinkVisit) *link.Visit {
	referer := ""
	if row.Referer.Valid {
		referer = row.Referer.String
	}

	return &link.Visit{
		ID:        row.ID,
		LinkID:    row.LinkID,
		IP:        row.Ip,
		UserAgent: row.UserAgent,
		Referer:   referer,
		Status:    int(row.Status),
		CreatedAt: row.CreatedAt.Time,
	}
}

func (r *VisitPostgres) Create(ctx context.Context, visit link.Visit) (*link.Visit, error) {
	row, err := r.queries.CreateLinkVisit(ctx, sqlc.CreateLinkVisitParams{
		LinkID:    visit.LinkID,
		Ip:        visit.IP,
		UserAgent: visit.UserAgent,
		Column4:   visit.Referer,
		Status:    int32(visit.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("create visit: %w", err)
	}

	return toVisit(row), nil
}

func (r *VisitPostgres) ListPaginated(ctx context.Context, limit, offset int32) ([]link.Visit, error) {
	rows, err := r.queries.ListLinkVisitsPaginated(ctx, sqlc.ListLinkVisitsPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list visits paginated: %w", err)
	}

	visits := make([]link.Visit, 0, len(rows))
	for _, row := range rows {
		visits = append(visits, *toVisit(row))
	}

	return visits, nil
}

func (r *VisitPostgres) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountLinkVisits(ctx)
	if err != nil {
		return 0, fmt.Errorf("count visits: %w", err)
	}
	return count, nil
}

func (r *VisitPostgres) Delete(ctx context.Context, id int64) error {
	rows, err := r.queries.DeleteLinkVisit(ctx, id)
	if err != nil {
		return fmt.Errorf("delete visit: %w", err)
	}
	if rows == 0 {
		return link.ErrVisitNotFound
	}
	return nil
}
