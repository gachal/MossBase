package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockUserRepo struct {
	users    map[uint64]*entity.User
	emailMap map[string]*entity.User
	nextID   uint64
	count    int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:    make(map[uint64]*entity.User),
		emailMap: make(map[string]*entity.User),
		nextID:   1,
		count:    0,
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *entity.User) error {
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.emailMap[user.Email] = user
	m.count++
	return nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id uint64) (*entity.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	u, ok := m.emailMap[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Update(_ context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]entity.User, int64, error) {
	return nil, m.count, nil
}

func (m *mockUserRepo) Count(_ context.Context) (int64, error) {
	return m.count, nil
}

func TestUserService_Register(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	t.Run("first user becomes admin", func(t *testing.T) {
		result, err := svc.Register(context.Background(), dto.RegisterRequest{
			Email: "admin@test.com", Username: "admin", Password: "password123",
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "admin", result.User.Role)
	})

	t.Run("second user is regular user", func(t *testing.T) {
		result, err := svc.Register(context.Background(), dto.RegisterRequest{
			Email: "user@test.com", Username: "user1", Password: "password123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "user", result.User.Role)
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		_, err := svc.Register(context.Background(), dto.RegisterRequest{
			Email: "admin@test.com", Username: "another", Password: "password123",
		})
		assert.Error(t, err)
	})
}

func TestUserService_Login(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, "test-secret", 24)

	svc.Register(context.Background(), dto.RegisterRequest{
		Email: "login@test.com", Username: "loginuser", Password: "mypass123",
	})

	t.Run("successful login", func(t *testing.T) {
		result, err := svc.Login(context.Background(), dto.LoginRequest{
			Email: "login@test.com", Password: "mypass123",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, result.Token)
		assert.Equal(t, "login@test.com", result.User.Email)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(context.Background(), dto.LoginRequest{
			Email: "login@test.com", Password: "wrong",
		})
		assert.Error(t, err)
	})

	t.Run("unknown email", func(t *testing.T) {
		_, err := svc.Login(context.Background(), dto.LoginRequest{
			Email: "unknown@test.com", Password: "mypass123",
		})
		assert.Error(t, err)
	})
}
