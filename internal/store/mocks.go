package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStorage() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Create(context.Context, *sql.Tx, *User) error {
	return nil
}
func (m *MockUserStore) GetByID(context.Context, int64) (*User, error) {
	return &User{
		ID: 200,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}, nil
}
func (m *MockUserStore) CreateAndInvite(context.Context, *User, string, time.Duration) error {
	return nil
}
func (m *MockUserStore) Activate(context.Context, string) error {
	return nil
}
func (m *MockUserStore) Delete(context.Context, int64) error {
	return nil
}
func (m *MockUserStore) GetByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}
