package academic

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

	handler := &Handler{
		deps: HandlerDeps{
			DB:       mockDB,
			Cache:    nil,
			EventBus: mockEventBus,
			RBAC:     rbac,
			Logger:   zap.NewNop(),
		},
		repo: mockRepo,
	}

	return handler, mockRepo
}

func TestHandler_ListAcademicYears_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	years := []*AcademicYear{
		{
			ID:        "year-1",
			TenantID:  "tenant-123",
			Name:      "2024-2025",
			StartDate: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC),
			IsCurrent: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockRepo.On("ListAcademicYears", mock.Anything, 20, 0).Return(years, int64(1), nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.ListAcademicYears(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_CreateAcademicYear_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateAcademicYearRequest{
		TenantID:  "tenant-123",
		Name:      "2024-2025",
		StartDate: "2024-09-01",
		EndDate:   "2025-06-30",
		IsCurrent: true,
	}

	mockRepo.On("CreateAcademicYear", mock.Anything, mock.AnythingOfType("*academic.AcademicYear")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "academic_year.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.CreateAcademicYear(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_CreateClass_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateClassRequest{
		TenantID:       "tenant-123",
		SchoolID:       "school-123",
		GradeLevelID:   "grade-123",
		AcademicYearID: "year-123",
		Name:           "Class 10A",
		Section:        "A",
		MaxStudents:    30,
	}

	mockRepo.On("CreateClass", mock.Anything, mock.AnythingOfType("*academic.Class")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "class.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.CreateClass(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_GetClass_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	class := &Class{
		ID:              "class-123",
		TenantID:        "tenant-123",
		SchoolID:        "school-123",
		GradeLevelID:    "grade-123",
		AcademicYearID:  "year-123",
		Name:            "Class 10A",
		MaxStudents:     30,
		CurrentStudents: 25,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockRepo.On("GetClassByID", mock.Anything, "class-123").Return(class, nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Params("id", "class-123")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.GetClass(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_CreateGrade_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateGradeRequest{
		TenantID:       "tenant-123",
		StudentID:      "student-123",
		ClassID:        "class-123",
		SubjectID:      "subject-123",
		AssessmentType: "exam",
		Score:          85.5,
		MaxScore:       100.0,
		Grade:          "A",
		GradePoint:     4.0,
	}

	mockRepo.On("CreateGrade", mock.Anything, mock.AnythingOfType("*academic.Grade")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "grade.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")
	ctx.Locals("user_id", "teacher-123")

	err := handler.CreateGrade(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_CreateGrade_ScoreExceedsMax(t *testing.T) {
	handler, _ := setupTestHandler()
	app := fiber.New()

	reqBody := CreateGradeRequest{
		TenantID:       "tenant-123",
		StudentID:      "student-123",
		ClassID:        "class-123",
		SubjectID:      "subject-123",
		AssessmentType: "exam",
		Score:          110.0, // Exceeds max score
		MaxScore:       100.0,
	}

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")
	ctx.Locals("user_id", "teacher-123")

	err := handler.CreateGrade(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
}

func TestHandler_UpdateGrade_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	existingGrade := &Grade{
		ID:             "grade-123",
		TenantID:       "tenant-123",
		StudentID:      "student-123",
		ClassID:        "class-123",
		SubjectID:      "subject-123",
		AssessmentType: "exam",
		Score:          80.0,
		MaxScore:       100.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	reqBody := UpdateGradeRequest{
		Score:    90.0,
		MaxScore: 100.0,
		Grade:    "A",
	}

	mockRepo.On("GetGradeByID", mock.Anything, "grade-123").Return(existingGrade, nil)
	mockRepo.On("UpdateGrade", mock.Anything, mock.AnythingOfType("*academic.Grade")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "grade.updated", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("PUT")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Params("id", "grade-123")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.UpdateGrade(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_MarkAttendance_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := MarkAttendanceRequest{
		TenantID:  "tenant-123",
		StudentID: "student-123",
		ClassID:   "class-123",
		Date:      "2024-01-15",
		Status:    "present",
	}

	mockRepo.On("MarkAttendance", mock.Anything, mock.AnythingOfType("*academic.Attendance")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "attendance.marked", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")
	ctx.Locals("user_id", "teacher-123")

	err := handler.MarkAttendance(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_BulkMarkAttendance_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := BulkAttendanceRequest{
		TenantID: "tenant-123",
		ClassID:  "class-123",
		Date:     "2024-01-15",
		Records: []BulkAttendanceRecord{
			{StudentID: "student-1", Status: "present"},
			{StudentID: "student-2", Status: "absent"},
			{StudentID: "student-3", Status: "late"},
		},
	}

	mockRepo.On("BulkMarkAttendance", mock.Anything, mock.AnythingOfType("[]*academic.Attendance")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "attendance.bulk_marked", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")
	ctx.Locals("user_id", "teacher-123")

	err := handler.BulkMarkAttendance(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_GetAttendanceStats_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	now := time.Now()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	attendance := []*Attendance{
		{ID: "1", Status: "present", Date: startDate.AddDate(0, 0, 1)},
		{ID: "2", Status: "present", Date: startDate.AddDate(0, 0, 2)},
		{ID: "3", Status: "absent", Date: startDate.AddDate(0, 0, 3)},
		{ID: "4", Status: "late", Date: startDate.AddDate(0, 0, 4)},
		{ID: "5", Status: "present", Date: startDate.AddDate(0, 0, 5)},
	}

	mockRepo.On("ListAttendanceByStudent", mock.Anything, "student-123", "class-123", mock.Anything, mock.Anything).Return(attendance, nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Params("student_id", "student-123")
	ctx.Request().URI().QueryArgs().Add("class_id", "class-123")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.GetAttendanceStats(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_CreateAssignment_Success(t *testing.T) {
	handler, mockRepo := setupTestHandler()
	app := fiber.New()

	reqBody := CreateAssignmentRequest{
		TenantID:    "tenant-123",
		ClassID:     "class-123",
		SubjectID:   "subject-123",
		TeacherID:   "teacher-123",
		Title:       "Math Assignment 1",
		Description: "Complete exercises 1-10",
		MaxScore:    100.0,
		DueDate:     time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
	}

	mockRepo.On("CreateAssignment", mock.Anything, mock.AnythingOfType("*academic.Assignment")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, "tenant-123").Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "assignment.created", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Locals("tenant_id", "tenant-123")

	err := handler.CreateAssignment(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestToAcademicYearDTO(t *testing.T) {
	startDate := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	year := &AcademicYear{
		ID:        "year-123",
		TenantID:  "tenant-123",
		Name:      "2024-2025",
		StartDate: startDate,
		EndDate:   endDate,
		IsCurrent: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dto := ToAcademicYearDTO(year)

	assert.Equal(t, year.ID, dto.ID)
	assert.Equal(t, year.TenantID, dto.TenantID)
	assert.Equal(t, year.Name, dto.Name)
	assert.Equal(t, "2024-09-01", dto.StartDate)
	assert.Equal(t, "2025-06-30", dto.EndDate)
	assert.Equal(t, year.IsCurrent, dto.IsCurrent)
	assert.NotEmpty(t, dto.CreatedAt)
	assert.NotEmpty(t, dto.UpdatedAt)
}

func TestToGradeDTO(t *testing.T) {
	now := time.Now()
	grade := "A"
	gradePoint := 4.0
	remarks := "Excellent work"
	graderID := "teacher-123"
	gradedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	gradeObj := &Grade{
		ID:             "grade-123",
		TenantID:       "tenant-123",
		StudentID:      "student-123",
		ClassID:        "class-123",
		SubjectID:      "subject-123",
		AssessmentType: "exam",
		Score:          85.5,
		MaxScore:       100.0,
		Grade:          &grade,
		GradePoint:     &gradePoint,
		Remarks:        &remarks,
		GradedBy:       &graderID,
		GradedDate:     &gradedDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	dto := ToGradeDTO(gradeObj)

	assert.Equal(t, gradeObj.ID, dto.ID)
	assert.Equal(t, gradeObj.TenantID, dto.TenantID)
	assert.Equal(t, gradeObj.StudentID, dto.StudentID)
	assert.Equal(t, gradeObj.ClassID, dto.ClassID)
	assert.Equal(t, gradeObj.SubjectID, dto.SubjectID)
	assert.Equal(t, gradeObj.AssessmentType, dto.AssessmentType)
	assert.Equal(t, gradeObj.Score, dto.Score)
	assert.Equal(t, gradeObj.MaxScore, dto.MaxScore)
	assert.Equal(t, 85.5, dto.Percentage)
	assert.Equal(t, grade, dto.Grade)
	assert.Equal(t, gradePoint, dto.GradePoint)
	assert.Equal(t, remarks, dto.Remarks)
	assert.Equal(t, graderID, dto.GradedBy)
	assert.Equal(t, "2024-01-15", dto.GradedDate)
}

func TestToAttendanceDTO(t *testing.T) {
	now := time.Now()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	remarks := "Late due to traffic"
	markerID := "teacher-123"
	markedAt := now

	attendance := &Attendance{
		ID:        "att-123",
		TenantID:  "tenant-123",
		StudentID: "student-123",
		ClassID:   "class-123",
		Date:      date,
		Status:    "late",
		Remarks:   &remarks,
		MarkedBy:  &markerID,
		MarkedAt:  &markedAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dto := ToAttendanceDTO(attendance)

	assert.Equal(t, attendance.ID, dto.ID)
	assert.Equal(t, attendance.TenantID, dto.TenantID)
	assert.Equal(t, attendance.StudentID, dto.StudentID)
	assert.Equal(t, attendance.ClassID, dto.ClassID)
	assert.Equal(t, "2024-01-15", dto.Date)
	assert.Equal(t, attendance.Status, dto.Status)
	assert.Equal(t, remarks, dto.Remarks)
	assert.Equal(t, markerID, dto.MarkedBy)
	assert.NotEmpty(t, dto.MarkedAt)
}
