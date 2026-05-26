package entity

import "time"

type MemberRole string

const (
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
	MemberRoleViewer MemberRole = "viewer"
)

type SpaceMember struct {
	ID        uint64
	SpaceID   uint64
	UserID    uint64
	Role      MemberRole
	CreatedAt time.Time
}

func (m SpaceMember) IsAdmin() bool {
	return m.Role == MemberRoleAdmin
}

func (m SpaceMember) CanEdit() bool {
	return m.Role == MemberRoleAdmin || m.Role == MemberRoleMember
}
