package service

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.LoginResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetProfile(ctx context.Context, userID uint64) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint64, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

func toUserResponse(u *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Avatar:    u.Avatar,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
