package entity

import "time"

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID           uint64
	Email        string
	Username     string
	PasswordHash string
	Avatar       string
	Role         UserRole
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}
