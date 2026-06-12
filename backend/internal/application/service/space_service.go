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
)

type SpaceService interface {
	Create(ctx context.Context, ownerID uint64, req dto.CreateSpaceRequest) (*dto.SpaceResponse, error)
	Update(ctx context.Context, spaceID uint64, req dto.UpdateSpaceRequest) (*dto.SpaceResponse, error)
	Delete(ctx context.Context, spaceID uint64) error
	GetByID(ctx context.Context, spaceID uint64) (*dto.SpaceResponse, error)
	ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]dto.SpaceResponse, int64, error)
	AddMember(ctx context.Context, spaceID, requesterID uint64, req dto.AddMemberRequest) error
	RemoveMember(ctx context.Context, spaceID, requesterID, targetUserID uint64) error
	ListMembers(ctx context.Context, spaceID uint64) ([]dto.SpaceMemberResponse, error)
}

type SpaceServiceImpl struct {
	spaceRepo      repository.SpaceRepository
	spaceMemberRepo repository.SpaceMemberRepository
	userRepo       repository.UserRepository
}

func NewSpaceService(
	spaceRepo repository.SpaceRepository,
	spaceMemberRepo repository.SpaceMemberRepository,
	userRepo repository.UserRepository,
) SpaceService {
	return &SpaceServiceImpl{spaceRepo: spaceRepo, spaceMemberRepo: spaceMemberRepo, userRepo: userRepo}
}

func (s *SpaceServiceImpl) Create(ctx context.Context, ownerID uint64, req dto.CreateSpaceRequest) (*dto.SpaceResponse, error) {
	visibility := "private"
	if req.Visibility != "" {
		visibility = req.Visibility
	}

	space := &entity.Space{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Cover:       req.Cover,
		Visibility:  entity.SpaceVisibility(visibility),
		OwnerID:     ownerID,
	}
	if err := s.spaceRepo.Create(ctx, space); err != nil {
		return nil, fmt.Errorf("create space: %w", err)
	}

	member := &entity.SpaceMember{
		SpaceID: space.ID,
		UserID:  ownerID,
		Role:    entity.MemberRoleAdmin,
	}
	if err := s.spaceMemberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("add owner as member: %w", err)
	}

	resp := toSpaceResponse(space)
	return &resp, nil
}

func (s *SpaceServiceImpl) Update(ctx context.Context, spaceID uint64, req dto.UpdateSpaceRequest) (*dto.SpaceResponse, error) {
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("find space: %w", err)
	}

	if req.Name != "" {
		space.Name = req.Name
	}
	if req.Description != "" {
		space.Description = req.Description
	}
	if req.Icon != "" {
		space.Icon = req.Icon
	}
	if req.Cover != "" {
		space.Cover = req.Cover
	}
	if req.Visibility != "" {
		space.Visibility = entity.SpaceVisibility(req.Visibility)
	}

	if err := s.spaceRepo.Update(ctx, space); err != nil {
		return nil, fmt.Errorf("update space: %w", err)
	}
	resp := toSpaceResponse(space)
	return &resp, nil
}

func (s *SpaceServiceImpl) Delete(ctx context.Context, spaceID uint64) error {
	return s.spaceRepo.Delete(ctx, spaceID)
}

func (s *SpaceServiceImpl) GetByID(ctx context.Context, spaceID uint64) (*dto.SpaceResponse, error) {
	space, err := s.spaceRepo.FindByID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("find space: %w", err)
	}
	resp := toSpaceResponse(space)
	return &resp, nil
}

func (s *SpaceServiceImpl) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]dto.SpaceResponse, int64, error) {
	offset := (page - 1) * pageSize
	spaces, total, err := s.spaceRepo.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list spaces: %w", err)
	}
	result := make([]dto.SpaceResponse, len(spaces))
	for i, sp := range spaces {
		result[i] = toSpaceResponse(&sp)
	}
	return result, total, nil
}

func (s *SpaceServiceImpl) AddMember(ctx context.Context, spaceID, requesterID uint64, req dto.AddMemberRequest) error {
	member, err := s.spaceMemberRepo.FindBySpaceAndUser(ctx, spaceID, requesterID)
	if err != nil {
		return errors.New("requester is not a space member")
	}
	if !member.IsAdmin() {
		return errors.New("only space admins can add members")
	}

	existing, err := s.spaceMemberRepo.FindBySpaceAndUser(ctx, spaceID, req.UserID)
	if err == nil && existing != nil {
		return errors.New("user is already a member")
	}

	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("find user: %w", err)
	}
	_ = user

	newMember := &entity.SpaceMember{
		SpaceID: spaceID,
		UserID:  req.UserID,
		Role:    entity.MemberRole(req.Role),
	}
	return s.spaceMemberRepo.Create(ctx, newMember)
}

func (s *SpaceServiceImpl) RemoveMember(ctx context.Context, spaceID, requesterID, targetUserID uint64) error {
	member, err := s.spaceMemberRepo.FindBySpaceAndUser(ctx, spaceID, requesterID)
	if err != nil {
		return errors.New("requester is not a space member")
	}
	if !member.IsAdmin() {
		return errors.New("only space admins can remove members")
	}

	adminCount, err := s.spaceMemberRepo.CountAdminsBySpaceID(ctx, spaceID)
	if err != nil {
		return fmt.Errorf("count admins: %w", err)
	}

	target, err := s.spaceMemberRepo.FindBySpaceAndUser(ctx, spaceID, targetUserID)
	if err != nil {
		return errors.New("target user is not a member")
	}

	if target.IsAdmin() && adminCount <= 1 {
		return errors.New("cannot remove the last admin")
	}

	zap.L().Info("removing member", zap.Uint64("space_id", spaceID), zap.Uint64("user_id", targetUserID))
	return s.spaceMemberRepo.Delete(ctx, spaceID, targetUserID)
}

func (s *SpaceServiceImpl) ListMembers(ctx context.Context, spaceID uint64) ([]dto.SpaceMemberResponse, error) {
	members, err := s.spaceMemberRepo.FindBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	result := make([]dto.SpaceMemberResponse, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		result = append(result, dto.SpaceMemberResponse{
			ID:        m.ID,
			SpaceID:   m.SpaceID,
			UserID:    m.UserID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      string(m.Role),
			CreatedAt: m.CreatedAt,
		})
	}
	return result, nil
}

func toSpaceResponse(sp *entity.Space) dto.SpaceResponse {
	return dto.SpaceResponse{
		ID:          sp.ID,
		Name:        sp.Name,
		Description: sp.Description,
		Icon:        sp.Icon,
		Cover:       sp.Cover,
		Visibility:  string(sp.Visibility),
		OwnerID:     sp.OwnerID,
		CreatedAt:   sp.CreatedAt,
		UpdatedAt:   sp.UpdatedAt,
	}
}
