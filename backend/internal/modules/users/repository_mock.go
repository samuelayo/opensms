package users

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	mock.Mock
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// User operations
func (m *MockRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*User), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User, passwordHash string) error {
	args := m.Called(ctx, user, passwordHash)
	return args.Error(0)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Student operations
func (m *MockRepository) ListStudents(ctx context.Context, limit, offset int) ([]*Student, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Student), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetStudentByID(ctx context.Context, id string) (*Student, *User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*Student), nil, args.Error(2)
	}
	return args.Get(0).(*Student), args.Get(1).(*User), args.Error(2)
}

func (m *MockRepository) GetStudentByUserID(ctx context.Context, userID string) (*Student, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Student), args.Error(1)
}

func (m *MockRepository) CreateStudent(ctx context.Context, student *Student) error {
	args := m.Called(ctx, student)
	return args.Error(0)
}

func (m *MockRepository) UpdateStudent(ctx context.Context, student *Student) error {
	args := m.Called(ctx, student)
	return args.Error(0)
}

// Teacher operations
func (m *MockRepository) ListTeachers(ctx context.Context, limit, offset int) ([]*Teacher, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Teacher), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetTeacherByID(ctx context.Context, id string) (*Teacher, *User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*Teacher), nil, args.Error(2)
	}
	return args.Get(0).(*Teacher), args.Get(1).(*User), args.Error(2)
}

func (m *MockRepository) GetTeacherByUserID(ctx context.Context, userID string) (*Teacher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Teacher), args.Error(1)
}

func (m *MockRepository) CreateTeacher(ctx context.Context, teacher *Teacher) error {
	args := m.Called(ctx, teacher)
	return args.Error(0)
}

func (m *MockRepository) UpdateTeacher(ctx context.Context, teacher *Teacher) error {
	args := m.Called(ctx, teacher)
	return args.Error(0)
}
