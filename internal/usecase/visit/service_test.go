package visit

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
	svc       *service
	visitRepo *link.MockVisitRepository
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	ctrl := testutil.NewController(t)
	visitRepo := link.NewMockVisitRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &testEnv{
		svc:       NewService(visitRepo, logger),
		visitRepo: visitRepo,
	}
}

func makeTestVisit(id, linkID int64) *link.Visit {
	return &link.Visit{
		ID:        id,
		LinkID:    linkID,
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
		Status:    302,
		CreatedAt: testTimestamp,
	}
}

func TestService_CreateVisit(t *testing.T) {
	tests := []struct {
		name      string
		visit     link.Visit
		mockSetup func(*link.MockVisitRepository)
		wantErr   bool
		wantID    int64
	}{
		{
			name:  "success",
			visit: link.Visit{LinkID: 1, IP: "127.0.0.1", UserAgent: "test-agent", Status: 302},
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().Create(gomock.Any(), link.Visit{LinkID: 1, IP: "127.0.0.1", UserAgent: "test-agent", Status: 302}).Return(makeTestVisit(1, 1), nil)
			},
			wantID: 1,
		},
		{
			name:  "error",
			visit: link.Visit{LinkID: 1, IP: "127.0.0.1", Status: 302},
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.visitRepo)

			result, err := env.svc.CreateVisit(context.Background(), tt.visit)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, result.ID)
		})
	}
}

func TestService_ListVisitsPaginated(t *testing.T) {
	tests := []struct {
		name      string
		limit     int32
		offset    int32
		mockSetup func(*link.MockVisitRepository)
		wantErr   bool
		wantCount int
		wantTotal int64
	}{
		{
			name:   "success",
			limit:  10,
			offset: 0,
			mockSetup: func(r *link.MockVisitRepository) {
				visits := []link.Visit{*makeTestVisit(1, 1), *makeTestVisit(2, 1)}
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(0)).Return(visits, nil)
				r.EXPECT().Count(gomock.Any()).Return(int64(100), nil)
			},
			wantCount: 2,
			wantTotal: 100,
		},
		{
			name:   "list error",
			limit:  10,
			offset: 0,
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(0)).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "count error",
			limit:  10,
			offset: 0,
			mockSetup: func(r *link.MockVisitRepository) {
				visits := []link.Visit{*makeTestVisit(1, 1)}
				r.EXPECT().ListPaginated(gomock.Any(), int32(10), int32(0)).Return(visits, nil)
				r.EXPECT().Count(gomock.Any()).Return(int64(0), errors.New("count error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.visitRepo)

			result, err := env.svc.ListVisitsPaginated(context.Background(), tt.limit, tt.offset)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result.Visits, tt.wantCount)
			assert.Equal(t, tt.wantTotal, result.Total)
		})
	}
}

func TestService_CountVisits(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*link.MockVisitRepository)
		wantErr   bool
		wantCount int64
	}{
		{
			name: "success",
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().Count(gomock.Any()).Return(int64(100), nil)
			},
			wantCount: 100,
		},
		{
			name: "error",
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().Count(gomock.Any()).Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.visitRepo)

			count, err := env.svc.CountVisits(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestService_DeleteVisit(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*link.MockVisitRepository)
		wantErr   bool
	}{
		{
			name: "success",
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().Delete(gomock.Any(), int64(10)).Return(nil)
			},
		},
		{
			name: "not found",
			mockSetup: func(r *link.MockVisitRepository) {
				r.EXPECT().Delete(gomock.Any(), int64(10)).Return(link.ErrVisitNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			tt.mockSetup(env.visitRepo)

			err := env.svc.DeleteVisit(context.Background(), 10)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
