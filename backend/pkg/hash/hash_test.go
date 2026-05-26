package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "my-secure-password-123"
	hashed, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, password, hashed)

	assert.True(t, CheckPassword(password, hashed))
	assert.False(t, CheckPassword("wrong-password", hashed))
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	hash1, _ := HashPassword("same-password")
	hash2, _ := HashPassword("same-password")
	assert.NotEqual(t, hash1, hash2, "bcrypt should produce different hashes each time")
}
