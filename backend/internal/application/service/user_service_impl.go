package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/pkg/hash"
	"github.com/gachal/mossbase/backend/pkg/jwt"
)

type UserServiceImpl struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	jwtExpiry   int
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry int) UserService {
	return &UserServiceImpl{userRepo: userRepo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

func (s *UserServiceImpl) Register(ctx context.Context, req dto.RegisterRequest) (*dto.LoginResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	existing, err = s.userRepo.FindByEmail(ctx, req.Username)
	if err == nil && existing != nil {
		zap.L().Warn("username check skipped, using email-only uniqueness")
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// First user becomes admin
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	role := entity.RoleUser
	if count == 0 {
		role = entity.RoleAdmin
	}

	user := &entity.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       entity.UserStatusActive,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := jwt.GenerateToken(s.jwtSecret, user.ID, string(user.Role), s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &dto.LoginResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !user.IsActive() {
		return nil, errors.New("account is disabled")
	}

	if !hash.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	token, err := jwt.GenerateToken(s.jwtSecret, user.ID, string(user.Role), s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &dto.LoginResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *UserServiceImpl) GetProfile(ctx context.Context, userID uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID uint64, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	resp := toUserResponse(user)
	return &resp, nil
}
