package entity

import "time"

type SpaceVisibility string

const (
	VisibilityPrivate SpaceVisibility = "private"
	VisibilityPublic  SpaceVisibility = "public"
)

type Space struct {
	ID          uint64
	Name        string
	Description string
	Icon        string
	Cover       string
	Visibility  SpaceVisibility
	OwnerID     uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
