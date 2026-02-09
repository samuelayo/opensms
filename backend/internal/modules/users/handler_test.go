package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/shared/security"
)

// MockDB for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) SetTenant(ctx context.Context, tenantID string) {
	m.Called(ctx, tenantID)
}

func (m *MockDB) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockEventBus for testing
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(subject string, data interface{}) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockEventBus) Health() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEventBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestHandler() (*Handler, *MockRepository) {
	mockRepo := NewMockRepository()
	mockDB := new(MockDB)
	mockEventBus := new(MockEventBus)
	rbac := security.NewRBAC()
	crypto, _ := security.NewCrypto("12345678901234567890123456789012")

	handler := &Handler{
		deps: HandlerDeps{
			DB:       mockDB,
			Cache:    nil,
			EventBus: mockEventBus,
			RBAC:     rbac,
			Logger:   zap.NewNop(),
			Crypto:   crypto,
		},
		repo: mockRepo,
	}

	return handler, mockRepo
}

func TestHandler_ListUsers_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	users := []*User{
		{
			ID:        "user-1",
			TenantID:  "tenant-123",
			Email:     "user1@example.com",
			Role:      "student",
			Status:    "active",
			FirstName: "John",
			LastName:  "Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "user-2",
			TenantID:  "tenant-123",
			Email:     "user2@example.com",
			Role:      "teacher",
			Status:    "active",
			FirstName: "Jane",
			LastName:  "Smith",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockRepo.On("ListUsers", mock.Anything, 20, 0).Return(users, int64(2), nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.ListUsers(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestHandler_CreateUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateUserRequest{
		Email:     "newuser@example.com",
		Password:  "SecurePass123!@#",
		FirstName: "New",
		LastName:  "User",
		Role:      "student",
		TenantID:  "tenant-123",
	}

	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*users.User"), mock.Anything).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "user.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.CreateUser(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_GetUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	user := &User{
		ID:        "user-123",
		TenantID:  "tenant-123",
		Email:     "test@example.com",
		Role:      "student",
		Status:    "active",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetUserByID", mock.Anything, "user-123").Return(user, nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Params("id", "user-123")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.GetUser(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_GetUser_NotFound(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	mockRepo.On("GetUserByID", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Params("id", "nonexistent")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.GetUser(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_UpdateUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	existingUser := &User{
		ID:        "user-123",
		TenantID:  "tenant-123",
		Email:     "test@example.com",
		Role:      "student",
		Status:    "active",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	reqBody := UpdateUserRequest{
		FirstName: "UpdatedFirst",
		LastName:  "UpdatedLast",
	}

	mockRepo.On("GetUserByID", mock.Anything, "user-123").Return(existingUser, nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*users.User")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "user.updated", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("PUT")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Params("id", "user-123")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.UpdateUser(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_DeleteUser_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	mockRepo.On("DeleteUser", mock.Anything, "user-123").Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "user.deleted", mock.Anything).Return(nil)

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("DELETE")
	ctx.Params("id", "user-123")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.DeleteUser(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_CreateStudent_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateStudentRequest{
		Email:           "student@example.com",
		Password:        "SecurePass123!@#",
		FirstName:       "Student",
		LastName:        "Test",
		TenantID:        "tenant-123",
		SchoolID:        "school-123",
		AdmissionNumber: "ADM001",
		AdmissionDate:   "2024-01-01",
		DateOfBirth:     "2000-01-01",
		Gender:          "male",
	}

	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*users.User"), mock.Anything).Return(nil)
	mockRepo.On("CreateStudent", mock.Anything, mock.AnythingOfType("*users.Student")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "student.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.CreateStudent(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_CreateTeacher_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateTeacherRequest{
		Email:         "teacher@example.com",
		Password:      "SecurePass123!@#",
		FirstName:     "Teacher",
		LastName:      "Test",
		TenantID:      "tenant-123",
		SchoolID:      "school-123",
		EmployeeID:    "EMP001",
		DateOfBirth:   "1980-01-01",
		Gender:        "female",
		DateOfJoining: "2024-01-01",
	}

	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*users.User"), mock.Anything).Return(nil)
	mockRepo.On("CreateTeacher", mock.Anything, mock.AnythingOfType("*users.Teacher")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "teacher.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.CreateTeacher(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestToUserDTO(t *testing.T) {
	now := time.Now()
	middleName := "Middle"
	phone := "+1234567890"
	avatarURL := "https://example.com/avatar.jpg"

	user := &User{
		ID:               "user-123",
		TenantID:         "tenant-123",
		Email:            "test@example.com",
		Role:             "student",
		Status:           "active",
		FirstName:        "John",
		LastName:         "Doe",
		MiddleName:       &middleName,
		Phone:            &phone,
		AvatarURL:        &avatarURL,
		EmailVerified:    true,
		TwoFactorEnabled: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	dto := ToUserDTO(user)

	assert.Equal(t, user.ID, dto.ID)
	assert.Equal(t, user.TenantID, dto.TenantID)
	assert.Equal(t, user.Email, dto.Email)
	assert.Equal(t, user.Role, dto.Role)
	assert.Equal(t, user.Status, dto.Status)
	assert.Equal(t, user.FirstName, dto.FirstName)
	assert.Equal(t, user.LastName, dto.LastName)
	assert.Equal(t, middleName, dto.MiddleName)
	assert.Equal(t, phone, dto.Phone)
	assert.Equal(t, avatarURL, dto.AvatarURL)
	assert.Equal(t, user.EmailVerified, dto.EmailVerified)
	assert.Equal(t, user.TwoFactorEnabled, dto.TwoFactorEnabled)
}

func TestToStudentDTO(t *testing.T) {
	now := time.Now()
	admissionDate, _ := time.Parse("2006-01-02", "2024-01-01")
	dateOfBirth, _ := time.Parse("2006-01-02", "2005-05-15")

	user := &User{
		ID:        "user-123",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	student := &Student{
		ID:              "student-123",
		TenantID:        "tenant-123",
		UserID:          "user-123",
		SchoolID:        "school-123",
		AdmissionNumber: "ADM001",
		AdmissionDate:   admissionDate,
		DateOfBirth:     dateOfBirth,
		Gender:          "male",
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	dto := ToStudentDTO(student, user)

	assert.Equal(t, student.ID, dto.ID)
	assert.Equal(t, student.TenantID, dto.TenantID)
	assert.Equal(t, student.UserID, dto.UserID)
	assert.Equal(t, student.SchoolID, dto.SchoolID)
	assert.Equal(t, student.AdmissionNumber, dto.AdmissionNumber)
	assert.Equal(t, "2024-01-01", dto.AdmissionDate)
	assert.Equal(t, "2005-05-15", dto.DateOfBirth)
	assert.Equal(t, student.Gender, dto.Gender)
	assert.Equal(t, student.IsActive, dto.IsActive)
	assert.NotNil(t, dto.User)
	assert.Equal(t, user.FirstName, dto.User.FirstName)
}

func TestToTeacherDTO(t *testing.T) {
	now := time.Now()
	dateOfBirth, _ := time.Parse("2006-01-02", "1980-01-01")
	dateOfJoining, _ := time.Parse("2006-01-02", "2024-01-01")
	salary := 50000.00

	user := &User{
		ID:        "user-123",
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
	}

	teacher := &Teacher{
		ID:              "teacher-123",
		TenantID:        "tenant-123",
		UserID:          "user-123",
		SchoolID:        "school-123",
		EmployeeID:      "EMP001",
		DateOfBirth:     dateOfBirth,
		Gender:          "female",
		DateOfJoining:   dateOfJoining,
		ExperienceYears: 5,
		Salary:          &salary,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	dto := ToTeacherDTO(teacher, user)

	assert.Equal(t, teacher.ID, dto.ID)
	assert.Equal(t, teacher.TenantID, dto.TenantID)
	assert.Equal(t, teacher.UserID, dto.UserID)
	assert.Equal(t, teacher.SchoolID, dto.SchoolID)
	assert.Equal(t, teacher.EmployeeID, dto.EmployeeID)
	assert.Equal(t, "1980-01-01", dto.DateOfBirth)
	assert.Equal(t, "2024-01-01", dto.DateOfJoining)
	assert.Equal(t, teacher.Gender, dto.Gender)
	assert.Equal(t, teacher.ExperienceYears, dto.ExperienceYears)
	assert.Equal(t, salary, dto.Salary)
	assert.Equal(t, teacher.IsActive, dto.IsActive)
	assert.NotNil(t, dto.User)
	assert.Equal(t, user.FirstName, dto.User.FirstName)
}
