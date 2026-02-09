package academic

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// HandlerDeps holds dependencies for the academic handler
type HandlerDeps struct {
	DB       *database.DB
	Cache    *cache.RedisClient
	EventBus *eventbus.NATSConnection
	RBAC     *security.RBAC
	Logger   *zap.Logger
}

// Handler handles academic management requests
type Handler struct {
	deps     HandlerDeps
	repo     Repository
	validate *validator.Validate
}

// NewHandler creates a new academic handler
func NewHandler(deps HandlerDeps) *Handler {
	return &Handler{
		deps:     deps,
		repo:     NewPostgresRepository(deps.DB.Pool),
		validate: validator.New(),
	}
}

// Academic Years

// ListAcademicYears retrieves all academic years with pagination
func (h *Handler) ListAcademicYears(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get tenant from context
	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	// Parse pagination
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	// Get academic years
	years, total, err := h.repo.ListAcademicYears(ctx, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list academic years", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve academic years",
		})
	}

	// Convert to DTOs
	dtos := make([]*AcademicYearDTO, len(years))
	for i, year := range years {
		dtos[i] = ToAcademicYearDTO(year)
	}

	return c.JSON(&PaginatedResponse{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

// CreateAcademicYear creates a new academic year
func (h *Handler) CreateAcademicYear(c *fiber.Ctx) error {
	ctx := context.Background()

	var req CreateAcademicYearRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start_date format, use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end_date format, use YYYY-MM-DD",
		})
	}

	if endDate.Before(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "End date must be after start date",
		})
	}

	year := &AcademicYear{
		ID:        uuid.New().String(),
		TenantID:  req.TenantID,
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		IsCurrent: req.IsCurrent,
	}

	if err := h.repo.CreateAcademicYear(ctx, year); err != nil {
		h.deps.Logger.Error("Failed to create academic year", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create academic year",
		})
	}

	// Publish event
	eventData := map[string]interface{}{
		"academic_year_id": year.ID,
		"tenant_id":        year.TenantID,
		"name":             year.Name,
		"is_current":       year.IsCurrent,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("academic_year.created", eventJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(ToAcademicYearDTO(year))
}

// Classes

// ListClasses retrieves all classes with pagination
func (h *Handler) ListClasses(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	classes, total, err := h.repo.ListClasses(ctx, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list classes", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve classes",
		})
	}

	dtos := make([]*ClassDTO, len(classes))
	for i, class := range classes {
		dtos[i] = ToClassDTO(class)
	}

	return c.JSON(&PaginatedResponse{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

// GetClass retrieves a single class by ID
func (h *Handler) GetClass(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	classID := c.Params("id")
	if classID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Class ID is required",
		})
	}

	class, err := h.repo.GetClassByID(ctx, classID)
	if err != nil {
		if err.Error() == "class not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Class not found",
			})
		}
		h.deps.Logger.Error("Failed to get class", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve class",
		})
	}

	return c.JSON(ToClassDTO(class))
}

// CreateClass creates a new class
func (h *Handler) CreateClass(c *fiber.Ctx) error {
	ctx := context.Background()

	var req CreateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	class := &Class{
		ID:              uuid.New().String(),
		TenantID:        req.TenantID,
		SchoolID:        req.SchoolID,
		GradeLevelID:    req.GradeLevelID,
		AcademicYearID:  req.AcademicYearID,
		Name:            req.Name,
		MaxStudents:     req.MaxStudents,
		CurrentStudents: 0,
	}

	if req.Section != "" {
		class.Section = &req.Section
	}
	if req.ClassTeacherID != "" {
		class.ClassTeacherID = &req.ClassTeacherID
	}
	if req.ClassroomNumber != "" {
		class.ClassroomNumber = &req.ClassroomNumber
	}

	if err := h.repo.CreateClass(ctx, class); err != nil {
		h.deps.Logger.Error("Failed to create class", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create class",
		})
	}

	eventData := map[string]interface{}{
		"class_id":  class.ID,
		"tenant_id": class.TenantID,
		"school_id": class.SchoolID,
		"name":      class.Name,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("class.created", eventJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(ToClassDTO(class))
}

// UpdateClass updates an existing class
func (h *Handler) UpdateClass(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	classID := c.Params("id")
	if classID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Class ID is required",
		})
	}

	var req UpdateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	class, err := h.repo.GetClassByID(ctx, classID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Class not found",
		})
	}

	if req.Name != "" {
		class.Name = req.Name
	}
	if req.Section != "" {
		class.Section = &req.Section
	}
	if req.ClassTeacherID != "" {
		class.ClassTeacherID = &req.ClassTeacherID
	}
	if req.MaxStudents > 0 {
		class.MaxStudents = req.MaxStudents
	}
	if req.ClassroomNumber != "" {
		class.ClassroomNumber = &req.ClassroomNumber
	}

	if err := h.repo.UpdateClass(ctx, class); err != nil {
		h.deps.Logger.Error("Failed to update class", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update class",
		})
	}

	eventData := map[string]interface{}{
		"class_id":  class.ID,
		"tenant_id": class.TenantID,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("class.updated", eventJSON)
	}

	return c.JSON(ToClassDTO(class))
}

// Subjects

// ListSubjects retrieves all subjects with pagination
func (h *Handler) ListSubjects(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	subjects, total, err := h.repo.ListSubjects(ctx, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list subjects", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve subjects",
		})
	}

	dtos := make([]*SubjectDTO, len(subjects))
	for i, subject := range subjects {
		dtos[i] = ToSubjectDTO(subject)
	}

	return c.JSON(&PaginatedResponse{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

// CreateSubject creates a new subject
func (h *Handler) CreateSubject(c *fiber.Ctx) error {
	ctx := context.Background()

	var req CreateSubjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	subject := &Subject{
		ID:         uuid.New().String(),
		TenantID:   req.TenantID,
		Code:       req.Code,
		Name:       req.Name,
		Credits:    req.Credits,
		IsElective: req.IsElective,
	}

	if req.Description != "" {
		subject.Description = &req.Description
	}

	if err := h.repo.CreateSubject(ctx, subject); err != nil {
		h.deps.Logger.Error("Failed to create subject", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create subject",
		})
	}

	eventData := map[string]interface{}{
		"subject_id": subject.ID,
		"tenant_id":  subject.TenantID,
		"code":       subject.Code,
		"name":       subject.Name,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("subject.created", eventJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(ToSubjectDTO(subject))
}

// Grades

// GetStudentGrades retrieves grades for a specific student
func (h *Handler) GetStudentGrades(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	studentID := c.Params("student_id")
	if studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Student ID is required",
		})
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	grades, total, err := h.repo.ListGradesByStudent(ctx, studentID, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list student grades", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve grades",
		})
	}

	dtos := make([]*GradeDTO, len(grades))
	for i, grade := range grades {
		dtos[i] = ToGradeDTO(grade)
	}

	return c.JSON(&PaginatedResponse{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

// CreateGrade creates a new grade entry
func (h *Handler) CreateGrade(c *fiber.Ctx) error {
	ctx := context.Background()

	var req CreateGradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	// Validate score doesn't exceed max score
	if req.Score > req.MaxScore {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Score cannot exceed max score",
		})
	}

	// Get grader ID from context
	graderID := c.Locals("user_id").(string)

	grade := &Grade{
		ID:             uuid.New().String(),
		TenantID:       req.TenantID,
		StudentID:      req.StudentID,
		ClassID:        req.ClassID,
		SubjectID:      req.SubjectID,
		AssessmentType: req.AssessmentType,
		Score:          req.Score,
		MaxScore:       req.MaxScore,
		GradedBy:       &graderID,
	}

	if req.Grade != "" {
		grade.Grade = &req.Grade
	}
	if req.GradePoint > 0 {
		grade.GradePoint = &req.GradePoint
	}
	if req.Remarks != "" {
		grade.Remarks = &req.Remarks
	}
	if req.GradedDate != "" {
		gradedDate, err := time.Parse("2006-01-02", req.GradedDate)
		if err == nil {
			grade.GradedDate = &gradedDate
		}
	} else {
		now := time.Now()
		grade.GradedDate = &now
	}

	if err := h.repo.CreateGrade(ctx, grade); err != nil {
		h.deps.Logger.Error("Failed to create grade", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create grade",
		})
	}

	eventData := map[string]interface{}{
		"grade_id":   grade.ID,
		"student_id": grade.StudentID,
		"class_id":   grade.ClassID,
		"subject_id": grade.SubjectID,
		"score":      grade.Score,
		"max_score":  grade.MaxScore,
		"tenant_id":  grade.TenantID,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("grade.created", eventJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(ToGradeDTO(grade))
}

// UpdateGrade updates an existing grade
func (h *Handler) UpdateGrade(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	gradeID := c.Params("id")
	if gradeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Grade ID is required",
		})
	}

	var req UpdateGradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	grade, err := h.repo.GetGradeByID(ctx, gradeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Grade not found",
		})
	}

	if req.Score > 0 {
		grade.Score = req.Score
	}
	if req.MaxScore > 0 {
		grade.MaxScore = req.MaxScore
	}
	if req.Grade != "" {
		grade.Grade = &req.Grade
	}
	if req.GradePoint > 0 {
		grade.GradePoint = &req.GradePoint
	}
	if req.Remarks != "" {
		grade.Remarks = &req.Remarks
	}
	if req.GradedDate != "" {
		gradedDate, err := time.Parse("2006-01-02", req.GradedDate)
		if err == nil {
			grade.GradedDate = &gradedDate
		}
	}

	// Validate score doesn't exceed max score
	if grade.Score > grade.MaxScore {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Score cannot exceed max score",
		})
	}

	if err := h.repo.UpdateGrade(ctx, grade); err != nil {
		h.deps.Logger.Error("Failed to update grade", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update grade",
		})
	}

	eventData := map[string]interface{}{
		"grade_id":   grade.ID,
		"student_id": grade.StudentID,
		"tenant_id":  grade.TenantID,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("grade.updated", eventJSON)
	}

	return c.JSON(ToGradeDTO(grade))
}

// Attendance

// GetClassAttendance retrieves attendance records for a class
func (h *Handler) GetClassAttendance(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	classID := c.Params("class_id")
	if classID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Class ID is required",
		})
	}

	// Parse date range
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start_date format, use YYYY-MM-DD",
			})
		}
	} else {
		// Default to last 30 days
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end_date format, use YYYY-MM-DD",
			})
		}
	} else {
		endDate = time.Now()
	}

	records, err := h.repo.ListAttendanceByClass(ctx, classID, startDate, endDate)
	if err != nil {
		h.deps.Logger.Error("Failed to list attendance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve attendance",
		})
	}

	dtos := make([]*AttendanceDTO, len(records))
	for i, record := range records {
		dtos[i] = ToAttendanceDTO(record)
	}

	return c.JSON(fiber.Map{
		"data":       dtos,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"total":      len(dtos),
	})
}

// MarkAttendance marks attendance for a student
func (h *Handler) MarkAttendance(c *fiber.Ctx) error {
	ctx := context.Background()

	var req MarkAttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format, use YYYY-MM-DD",
		})
	}

	// Get marker ID from context
	markerID := c.Locals("user_id").(string)
	now := time.Now()

	attendance := &Attendance{
		ID:        uuid.New().String(),
		TenantID:  req.TenantID,
		StudentID: req.StudentID,
		ClassID:   req.ClassID,
		Date:      date,
		Status:    req.Status,
		MarkedBy:  &markerID,
		MarkedAt:  &now,
	}

	if req.Remarks != "" {
		attendance.Remarks = &req.Remarks
	}

	if err := h.repo.MarkAttendance(ctx, attendance); err != nil {
		h.deps.Logger.Error("Failed to mark attendance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark attendance",
		})
	}

	eventData := map[string]interface{}{
		"attendance_id": attendance.ID,
		"student_id":    attendance.StudentID,
		"class_id":      attendance.ClassID,
		"date":          attendance.Date.Format("2006-01-02"),
		"status":        attendance.Status,
		"tenant_id":     attendance.TenantID,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("attendance.marked", eventJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(ToAttendanceDTO(attendance))
}

// BulkMarkAttendance marks attendance for multiple students
func (h *Handler) BulkMarkAttendance(c *fiber.Ctx) error {
	ctx := context.Background()

	var req BulkAttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format, use YYYY-MM-DD",
		})
	}

	// Get marker ID from context
	markerID := c.Locals("user_id").(string)
	now := time.Now()

	// Create attendance records
	records := make([]*Attendance, len(req.Records))
	for i, record := range req.Records {
		attendance := &Attendance{
			ID:        uuid.New().String(),
			TenantID:  req.TenantID,
			StudentID: record.StudentID,
			ClassID:   req.ClassID,
			Date:      date,
			Status:    record.Status,
			MarkedBy:  &markerID,
			MarkedAt:  &now,
		}
		if record.Remarks != "" {
			attendance.Remarks = &record.Remarks
		}
		records[i] = attendance
	}

	if err := h.repo.BulkMarkAttendance(ctx, records); err != nil {
		h.deps.Logger.Error("Failed to bulk mark attendance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark attendance",
		})
	}

	eventData := map[string]interface{}{
		"class_id":       req.ClassID,
		"date":           date.Format("2006-01-02"),
		"student_count":  len(records),
		"tenant_id":      req.TenantID,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("attendance.bulk_marked", eventJSON)
	}

	dtos := make([]*AttendanceDTO, len(records))
	for i, record := range records {
		dtos[i] = ToAttendanceDTO(record)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Attendance marked successfully",
		"count":   len(dtos),
		"data":    dtos,
	})
}

// GetAttendanceStats retrieves attendance statistics for a student
func (h *Handler) GetAttendanceStats(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	studentID := c.Params("student_id")
	classID := c.Query("class_id")

	if studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Student ID is required",
		})
	}

	// Parse date range
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start_date format",
			})
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end_date format",
			})
		}
	} else {
		endDate = time.Now()
	}

	records, err := h.repo.ListAttendanceByStudent(ctx, studentID, classID, startDate, endDate)
	if err != nil {
		h.deps.Logger.Error("Failed to list student attendance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve attendance",
		})
	}

	// Calculate statistics
	totalDays := len(records)
	presentDays := 0
	absentDays := 0
	lateDays := 0
	excusedDays := 0

	for _, record := range records {
		switch record.Status {
		case "present":
			presentDays++
		case "absent":
			absentDays++
		case "late":
			lateDays++
		case "excused":
			excusedDays++
		}
	}

	attendanceRate := 0.0
	if totalDays > 0 {
		attendanceRate = (float64(presentDays) / float64(totalDays)) * 100
	}

	stats := &AttendanceStatsDTO{
		StudentID:      studentID,
		ClassID:        classID,
		TotalDays:      totalDays,
		PresentDays:    presentDays,
		AbsentDays:     absentDays,
		LateDays:       lateDays,
		ExcusedDays:    excusedDays,
		AttendanceRate: attendanceRate,
	}

	return c.JSON(stats)
}

// ListAssignments retrieves assignments for a class
func (h *Handler) ListAssignments(c *fiber.Ctx) error {
	ctx := context.Background()

	tenantID := c.Locals("tenant_id").(string)
	h.deps.DB.SetTenant(ctx, tenantID)

	classID := c.Query("class_id")
	if classID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "class_id query parameter is required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	assignments, total, err := h.repo.ListAssignments(ctx, classID, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list assignments", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve assignments",
		})
	}

	dtos := make([]*AssignmentDTO, len(assignments))
	for i, assignment := range assignments {
		dtos[i] = ToAssignmentDTO(assignment)
	}

	return c.JSON(&PaginatedResponse{
		Data:       dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

// CreateAssignment creates a new assignment
func (h *Handler) CreateAssignment(c *fiber.Ctx) error {
	ctx := context.Background()

	var req CreateAssignmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.deps.DB.SetTenant(ctx, req.TenantID)

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid due_date format, use RFC3339",
		})
	}

	assignment := &Assignment{
		ID:        uuid.New().String(),
		TenantID:  req.TenantID,
		ClassID:   req.ClassID,
		SubjectID: req.SubjectID,
		TeacherID: req.TeacherID,
		Title:     req.Title,
		MaxScore:  req.MaxScore,
		DueDate:   dueDate,
	}

	if req.Description != "" {
		assignment.Description = &req.Description
	}
	if req.AssignmentType != "" {
		assignment.AssignmentType = &req.AssignmentType
	}
	if req.AttachmentURL != "" {
		assignment.AttachmentURL = &req.AttachmentURL
	}

	if err := h.repo.CreateAssignment(ctx, assignment); err != nil {
		h.deps.Logger.Error("Failed to create assignment", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create assignment",
		})
	}

	eventData := map[string]interface{}{
		"assignment_id": assignment.ID,
		"class_id":      assignment.ClassID,
		"subject_id":    assignment.SubjectID,
		"title":         assignment.Title,
		"tenant_id":     assignment.TenantID,
	}
	if eventJSON, err := json.Marshal(eventData); err == nil {
		h.deps.EventBus.Publish("assignment.created", eventJSON)
	}

	return c.Status(fiber.StatusCreated).JSON(ToAssignmentDTO(assignment))
}
