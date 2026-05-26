package service

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/application/dto"
)

type AdminService interface {
	GetDashboardStats(ctx context.Context) (*dto.DashboardStatsResponse, error)
	ListUsers(ctx context.Context, page, pageSize int) ([]dto.AdminUserResponse, int64, error)
	UpdateUserRole(ctx context.Context, userID uint64, role string) error
	UpdateUserStatus(ctx context.Context, userID uint64, status string) error
	ListSpaces(ctx context.Context, page, pageSize int) ([]dto.AdminSpaceResponse, int64, error)
	GetSpaceDetail(ctx context.Context, spaceID uint64) (*dto.AdminSpaceResponse, error)
	DeleteSpace(ctx context.Context, spaceID uint64) error
	ListPages(ctx context.Context, page, pageSize int) ([]dto.AdminPageResponse, int64, error)
	DeletePage(ctx context.Context, pageID uint64) error
	GetSettings(ctx context.Context) (*dto.SettingsResponse, error)
	UpdateSettings(ctx context.Context, req dto.SettingsRequest) error
	TestRAGConnection(ctx context.Context, req dto.TestRAGRequest) (*dto.TestRAGResponse, error)
}
