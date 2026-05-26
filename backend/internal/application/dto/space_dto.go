package dto

import "time"

type CreateSpaceRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=500"`
	Visibility  string `json:"visibility" binding:"omitempty,oneof=private public"`
}

type UpdateSpaceRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=500"`
	Visibility  string `json:"visibility" binding:"omitempty,oneof=private public"`
}

type SpaceResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Visibility  string    `json:"visibility"`
	OwnerID     uint64    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SpaceMemberResponse struct {
	ID        uint64    `json:"id"`
	SpaceID   uint64    `json:"space_id"`
	UserID    uint64    `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AddMemberRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member viewer"`
}
