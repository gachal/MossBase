package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_IsAdmin(t *testing.T) {
	assert.True(t, User{Role: RoleAdmin}.IsAdmin())
	assert.False(t, User{Role: RoleUser}.IsAdmin())
}

func TestUser_IsActive(t *testing.T) {
	assert.True(t, User{Status: UserStatusActive}.IsActive())
	assert.False(t, User{Status: UserStatusDisabled}.IsActive())
}

func TestSpaceMember_Permissions(t *testing.T) {
	assert.True(t, SpaceMember{Role: MemberRoleAdmin}.IsAdmin())
	assert.False(t, SpaceMember{Role: MemberRoleMember}.IsAdmin())

	assert.True(t, SpaceMember{Role: MemberRoleAdmin}.CanEdit())
	assert.True(t, SpaceMember{Role: MemberRoleMember}.CanEdit())
	assert.False(t, SpaceMember{Role: MemberRoleViewer}.CanEdit())
}

func TestPage_IsRoot(t *testing.T) {
	assert.True(t, Page{}.IsRoot())
	assert.False(t, Page{ParentID: uint64Ptr(1)}.IsRoot())
}

func uint64Ptr(v uint64) *uint64 { return &v }
