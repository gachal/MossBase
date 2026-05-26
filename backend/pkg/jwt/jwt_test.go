package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-key"

	token, err := GenerateToken(secret, 1, "admin", 24)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseToken(secret, token)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), claims.UserID)
	assert.Equal(t, "admin", claims.UserRole)
}

func TestParseToken_Invalid(t *testing.T) {
	_, err := ParseToken("secret", "invalid-token")
	assert.Error(t, err)
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken("secret-1", 1, "user", 24)
	_, err := ParseToken("secret-2", token)
	assert.Error(t, err)
}

func TestParseToken_Expired(t *testing.T) {
	token, _ := GenerateToken("secret", 1, "user", -1)
	_, err := ParseToken("secret", token)
	assert.Error(t, err)
}

func TestTokenExpiry(t *testing.T) {
	token, _ := GenerateToken("secret", 1, "user", 1)
	claims, err := ParseToken("secret", token)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 2*time.Second)
}
