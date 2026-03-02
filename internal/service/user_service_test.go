package service

import (
	"context"
	"testing"
	"time"

	"golang-grpc-enterprise-demo/internal/model"
)

// ---------- mock repository ----------

type mockUserRepo struct {
	users map[int64]*model.User
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		users: map[int64]*model.User{
			1: {ID: 1, Name: "Alice", Email: "alice@example.com", PasswordHash: "$2a$10$dummy", CreatedAt: time.Now()},
			2: {ID: 2, Name: "Bob", Email: "bob@example.com", PasswordHash: "$2a$10$dummy", CreatedAt: time.Now()},
		},
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	user.ID = int64(len(m.users) + 1)
	user.CreatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]*model.User, int64, error) {
	all := make([]*model.User, 0, len(m.users))
	for _, u := range m.users {
		all = append(all, u)
	}
	total := int64(len(all))
	if offset >= len(all) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}

// ---------- tests ----------

func TestGetUser_Found(t *testing.T) {
	svc := NewUserService(newMockRepo())

	user, err := svc.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Name != "Alice" {
		t.Errorf("expected Alice, got %s", user.Name)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc := NewUserService(newMockRepo())

	_, err := svc.GetByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestListUsers_Pagination(t *testing.T) {
	svc := NewUserService(newMockRepo())

	users, total, err := svc.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user per page, got %d", len(users))
	}
}
