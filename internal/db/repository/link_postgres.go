package repository

import (
	"context"
	"errors"
	"fmt"

	"code/internal/db/sqlc"
	"code/internal/domain/link"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkPostgres struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewLinkPostgres(pool *pgxpool.Pool) *LinkPostgres {
	return &LinkPostgres{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func toLink(row sqlc.Link) *link.Link {
	return &link.Link{
		ID:          row.ID,
		OriginalURL: row.OriginalUrl,
		ShortName:   row.ShortName,
		CreatedAt:   row.CreatedAt.Time,
	}
}

func (r *LinkPostgres) Create(ctx context.Context, originalURL, shortName string) (*link.Link, error) {
	row, err := r.queries.CreateLink(ctx, sqlc.CreateLinkParams{
		OriginalUrl: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, link.ErrShortNameTaken
		}
		return nil, fmt.Errorf("create link: %w", err)
	}

	return toLink(row), nil
}

func (r *LinkPostgres) GetByID(ctx context.Context, id int64) (*link.Link, error) {
	row, err := r.queries.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, link.ErrNotFound
		}
		return nil, fmt.Errorf("get link by id: %w", err)
	}

	return toLink(row), nil
}

func (r *LinkPostgres) List(ctx context.Context) ([]link.Link, error) {
	rows, err := r.queries.ListLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}

	links := make([]link.Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, *toLink(row))
	}

	return links, nil
}

func (r *LinkPostgres) Update(ctx context.Context, id int64, originalURL, shortName string) (*link.Link, error) {
	row, err := r.queries.UpdateLink(ctx, sqlc.UpdateLinkParams{
		ID:          id,
		OriginalUrl: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, link.ErrNotFound
		}
		if isUniqueViolation(err) {
			return nil, link.ErrShortNameTaken
		}
		return nil, fmt.Errorf("update link: %w", err)
	}

	return toLink(row), nil
}

func (r *LinkPostgres) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteLink(ctx, id); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
