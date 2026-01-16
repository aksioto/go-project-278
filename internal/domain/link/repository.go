package link

import "context"

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=link LinkRepository VisitRepository

type LinkRepository interface {
	Create(ctx context.Context, originalURL, shortName string) (*Link, error)
	GetByID(ctx context.Context, id int64) (*Link, error)
	GetByShortName(ctx context.Context, shortName string) (*Link, error)
	List(ctx context.Context) ([]Link, error)
	ListPaginated(ctx context.Context, limit, offset int32) ([]Link, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (*Link, error)
	Delete(ctx context.Context, id int64) error
}

type VisitRepository interface {
	Create(ctx context.Context, visit Visit) (*Visit, error)
	ListPaginated(ctx context.Context, limit, offset int32) ([]Visit, error)
	Count(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id int64) error
}
