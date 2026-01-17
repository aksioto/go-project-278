package link

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"code/internal/domain/link"
	"code/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

var testTimestamp = time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

type testEnv struct {
	svc      *service
	linkRepo *link.MockLinkRepository
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	ctrl := testutil.NewController(t)
	linkRepo := link.NewMockLinkRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &testEnv{
		svc:      NewService(linkRepo, logger),
		linkRepo: linkRepo,
	}
}

func makeTestLink(id int64, url, shortName string) *link.Link {
	return &link.Link{
		ID:          id,
		OriginalURL: url,
		ShortName:   shortName,
		CreatedAt:   testTimestamp,
	}
}

func TestService_CreateLink(t *testing.T) {
	tests := []struct {
		name        string
		originalURL string
		shortName   string
		mockSetup   func(*link.MockLinkRepository)
		wantErr     error
		wantID      int64
	}{
		{
			name:        "with provided short name",
			originalURL: "https://example.com",
			shortName:   "test123",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().
					Create(gomock.Any(), "https://example.com", "test123").
					Return(makeTestLink(1, "https://example.com", "test123"), nil)
			},
			wantID: 1,
		},
		{
			name:        "with generated short name",
			originalURL: "https://example.com",
			shortName:   "",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().
					Create(gomock.Any(), "https://example.com", gomock.Any()).
					Return(makeTestLink(1, "https://example.com", "abc12345"), nil)
			},
			wantID: 1,
		},
		{
			name:        "retry on collision",
			originalURL: "https://example.com",
			shortName:   "",
			mockSetup: func(r *link.MockLinkRepository) {
				gomock.InOrder(
					r.EXPECT().
						Create(gomock.Any(), "https://example.com", gomock.Any()).
						Return(nil, link.ErrShortNameTaken),
					r.EXPECT().
						Create(gomock.Any(), "https://example.com", gomock.Any()).
						Return(makeTestLink(1, "https://example.com", "newname"), nil),
				)
			},
			wantID: 1,
		},
		{
			name:        "max retries exceeded",
			originalURL: "https://example.com",
			shortName:   "",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().
					Create(gomock.Any(), "https://example.com", gomock.Any()).
					Return(nil, link.ErrShortNameTaken).
					Times(maxShortNameRetries)
			},
			wantErr: errors.New("failed to generate unique short name"),
		},
		{
			name:        "short name taken with explicit name",
			originalURL: "https://example.com",
			shortName:   "taken",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().
					Create(gomock.Any(), "https://example.com", "taken").
					Return(nil, link.ErrShortNameTaken)
			},
			wantErr: link.ErrShortNameTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			result, err := env.svc.CreateLink(context.Background(), tt.originalURL, tt.shortName)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, link.ErrShortNameTaken) {
					require.ErrorIs(t, err, link.ErrShortNameTaken)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, result.ID)
		})
	}
}

func TestService_GetLink(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(*link.MockLinkRepository)
		wantErr   error
		wantID    int64
	}{
		{
			name: "success",
			id:   1,
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().GetByID(gomock.Any(), int64(1)).Return(makeTestLink(1, "https://example.com", "test123"), nil)
			},
			wantID: 1,
		},
		{
			name: "not found",
			id:   999,
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().GetByID(gomock.Any(), int64(999)).Return(nil, link.ErrNotFound)
			},
			wantErr: link.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			result, err := env.svc.GetLink(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, result.ID)
		})
	}
}

func TestService_GetLinkByShortName(t *testing.T) {
	tests := []struct {
		name          string
		shortName     string
		mockSetup     func(*link.MockLinkRepository)
		wantErr       error
		wantShortName string
	}{
		{
			name:      "success",
			shortName: "test123",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().GetByShortName(gomock.Any(), "test123").Return(makeTestLink(1, "https://example.com", "test123"), nil)
			},
			wantShortName: "test123",
		},
		{
			name:      "not found",
			shortName: "notexist",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().GetByShortName(gomock.Any(), "notexist").Return(nil, link.ErrNotFound)
			},
			wantErr: link.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			result, err := env.svc.GetLinkByShortName(context.Background(), tt.shortName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantShortName, result.ShortName)
		})
	}
}

func TestService_ListLinks(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*link.MockLinkRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "success",
			mockSetup: func(r *link.MockLinkRepository) {
				links := []link.Link{*makeTestLink(1, "https://example.com/1", "one"), *makeTestLink(2, "https://example.com/2", "two")}
				r.EXPECT().List(gomock.Any()).Return(links, nil)
			},
			wantCount: 2,
		},
		{
			name: "empty",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().List(gomock.Any()).Return([]link.Link{}, nil)
			},
			wantCount: 0,
		},
		{
			name: "error",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().List(gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			result, err := env.svc.ListLinks(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, tt.wantCount)
		})
	}
}

func TestService_ListLinksPaginated(t *testing.T) {
	tests := []struct {
		name      string
		limit     int32
		offset    int32
		mockSetup func(*link.MockLinkRepository)
		wantErr   bool
		wantCount int
		wantTotal int64
	}{
		{
			name:   "success",
			limit:  10,
			offset: 0,
			mockSetup: func(r *link.MockLinkRepository) {
				links := []link.Link{*makeTestLink(1, "https://example.com/1", "one"), *makeTestLink(2, "https://example.com/2", "two")}
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(0)).Return(links, nil)
				r.EXPECT().Count(gomock.Any()).Return(int64(42), nil)
			},
			wantCount: 2,
			wantTotal: 42,
		},
		{
			name:   "empty result",
			limit:  10,
			offset: 100,
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(100)).Return([]link.Link{}, nil)
				r.EXPECT().Count(gomock.Any()).Return(int64(5), nil)
			},
			wantCount: 0,
			wantTotal: 5,
		},
		{
			name:   "list error",
			limit:  10,
			offset: 0,
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(0)).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "count error",
			limit:  10,
			offset: 0,
			mockSetup: func(r *link.MockLinkRepository) {
				links := []link.Link{*makeTestLink(1, "https://example.com/1", "one")}
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(0)).Return(links, nil)
				r.EXPECT().Count(gomock.Any()).Return(int64(0), errors.New("count error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			result, err := env.svc.ListLinksPaginated(context.Background(), tt.limit, tt.offset)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result.Links, tt.wantCount)
			assert.Equal(t, tt.wantTotal, result.Total)
		})
	}
}

func TestService_UpdateLink(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		originalURL string
		shortName   string
		mockSetup   func(*link.MockLinkRepository)
		wantErr     error
		wantURL     string
	}{
		{
			name:        "success",
			id:          1,
			originalURL: "https://updated.com",
			shortName:   "upd",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Update(gomock.Any(), int64(1), "https://updated.com", "upd").Return(makeTestLink(1, "https://updated.com", "upd"), nil)
			},
			wantURL: "https://updated.com",
		},
		{
			name:        "not found",
			id:          999,
			originalURL: "https://updated.com",
			shortName:   "upd",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Update(gomock.Any(), int64(999), "https://updated.com", "upd").Return(nil, link.ErrNotFound)
			},
			wantErr: link.ErrNotFound,
		},
		{
			name:        "short name taken",
			id:          1,
			originalURL: "https://updated.com",
			shortName:   "taken",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Update(gomock.Any(), int64(1), "https://updated.com", "taken").Return(nil, link.ErrShortNameTaken)
			},
			wantErr: link.ErrShortNameTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			result, err := env.svc.UpdateLink(context.Background(), tt.id, tt.originalURL, tt.shortName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, result.OriginalURL)
		})
	}
}

func TestService_DeleteLink(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(*link.MockLinkRepository)
		wantErr   error
	}{
		{
			name: "success",
			id:   1,
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)
			},
		},
		{
			name: "not found",
			id:   999,
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Delete(gomock.Any(), int64(999)).Return(link.ErrNotFound)
			},
			wantErr: link.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			err := env.svc.DeleteLink(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_CountLinks(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*link.MockLinkRepository)
		wantErr   bool
		wantCount int64
	}{
		{
			name: "success",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Count(gomock.Any()).Return(int64(42), nil)
			},
			wantCount: 42,
		},
		{
			name: "error",
			mockSetup: func(r *link.MockLinkRepository) {
				r.EXPECT().Count(gomock.Any()).Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.linkRepo)

			count, err := env.svc.CountLinks(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
