package link

import "context"

type Repository interface {
	Create(ctx context.Context, originalURL, shortName string) (*Link, error)
	GetByID(ctx context.Context, id int64) (*Link, error)
	List(ctx context.Context) ([]Link, error)
	ListPaginated(ctx context.Context, limit, offset int32) ([]Link, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (*Link, error)
	Delete(ctx context.Context, id int64) error
}
