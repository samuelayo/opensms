package users

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// HandlerDeps holds dependencies for the users handler
type HandlerDeps struct {
	DB       *database.DB
	Cache    *cache.RedisClient
	EventBus *eventbus.NATSConnection
	RBAC     *security.RBAC
	Logger   *zap.Logger
	Crypto   *security.Crypto
}

// Handler handles user management requests
type Handler struct {
	deps HandlerDeps
	repo Repository
}

// NewHandler creates a new users handler
func NewHandler(deps HandlerDeps) *Handler {
	repo := NewPostgresRepository(deps.DB.Pool())

	return &Handler{
		deps: deps,
		repo: repo,
	}
}

// ListUsers retrieves a paginated list of users
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Parse pagination params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Get users
	users, total, err := h.repo.ListUsers(ctx, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list users", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve users",
		})
	}

	// Convert to DTOs
	userDTOs := make([]*UserDTO, 0, len(users))
	for _, user := range users {
		userDTOs = append(userDTOs, ToUserDTO(user))
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return c.JSON(&PaginatedResponse{
		Data:       userDTOs,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetUser retrieves a single user by ID
func (h *Handler) GetUser(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Get user
	user, err := h.repo.GetUser(ctx, userID)
	if err != nil {
		if err == ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			})
		}
		h.deps.Logger.Error("Failed to get user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve user",
		})
	}

	return c.JSON(ToUserDTO(user))
}

// CreateUser creates a new user
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, req.TenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Hash password
	passwordHash, err := h.deps.Crypto.HashPassword(req.Password)
	if err != nil {
		h.deps.Logger.Error("Failed to hash password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create user",
		})
	}

	// Create user
	user := &User{
		ID:               uuid.New().String(),
		TenantID:         req.TenantID,
		Email:            req.Email,
		Role:             req.Role,
		Status:           "active",
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		MiddleName:       stringPtr(req.MiddleName),
		Phone:            stringPtr(req.Phone),
		EmailVerified:    false,
		TwoFactorEnabled: false,
	}

	if err := h.repo.CreateUser(ctx, user, passwordHash); err != nil {
		h.deps.Logger.Error("Failed to create user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create user",
		})
	}

	// Publish user.created event
	if h.deps.EventBus != nil {
		eventData := map[string]interface{}{
			"user_id":   user.ID,
			"email":     user.Email,
			"tenant_id": user.TenantID,
			"role":      user.Role,
		}
		if err := h.deps.EventBus.Publish("user.created", eventData); err != nil {
			h.deps.Logger.Warn("Failed to publish user.created event", zap.Error(err))
		}
	}

	h.deps.Logger.Info("User created", zap.String("user_id", user.ID), zap.String("email", user.Email))

	return c.Status(fiber.StatusCreated).JSON(ToUserDTO(user))
}

// UpdateUser updates an existing user
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()
	userID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Get existing user
	user, err := h.repo.GetUser(ctx, userID)
	if err != nil {
		if err == ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			})
		}
		h.deps.Logger.Error("Failed to get user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update user",
		})
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.MiddleName != "" {
		user.MiddleName = &req.MiddleName
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.AvatarURL != "" {
		user.AvatarURL = &req.AvatarURL
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	// Update user
	if err := h.repo.UpdateUser(ctx, user); err != nil {
		h.deps.Logger.Error("Failed to update user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update user",
		})
	}

	h.deps.Logger.Info("User updated", zap.String("user_id", user.ID))

	return c.JSON(ToUserDTO(user))
}

// DeleteUser soft deletes a user
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Delete user
	if err := h.repo.DeleteUser(ctx, userID); err != nil {
		if err == ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			})
		}
		h.deps.Logger.Error("Failed to delete user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to delete user",
		})
	}

	h.deps.Logger.Info("User deleted", zap.String("user_id", userID))

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// ListStudents retrieves a paginated list of students
func (h *Handler) ListStudents(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Parse pagination params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Get students
	students, total, err := h.repo.ListStudents(ctx, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list students", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve students",
		})
	}

	// Convert to DTOs
	studentDTOs := make([]*StudentDTO, 0, len(students))
	for _, student := range students {
		// Optionally get user data
		user, _ := h.repo.GetUser(ctx, student.UserID)
		studentDTOs = append(studentDTOs, ToStudentDTO(student, user))
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return c.JSON(&PaginatedResponse{
		Data:       studentDTOs,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetStudent retrieves a single student by ID
func (h *Handler) GetStudent(c *fiber.Ctx) error {
	ctx := c.Context()
	studentID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Get student
	student, err := h.repo.GetStudent(ctx, studentID)
	if err != nil {
		if err == ErrStudentNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "STUDENT_NOT_FOUND",
				"message": "Student not found",
			})
		}
		h.deps.Logger.Error("Failed to get student", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve student",
		})
	}

	// Get user data
	user, _ := h.repo.GetUser(ctx, student.UserID)

	return c.JSON(ToStudentDTO(student, user))
}

// CreateStudent creates a new student (and associated user)
func (h *Handler) CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, req.TenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Hash password
	passwordHash, err := h.deps.Crypto.HashPassword(req.Password)
	if err != nil {
		h.deps.Logger.Error("Failed to hash password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create student",
		})
	}

	// Create user first
	userID := uuid.New().String()
	user := &User{
		ID:               userID,
		TenantID:         req.TenantID,
		Email:            req.Email,
		Role:             "student",
		Status:           "active",
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		MiddleName:       stringPtr(req.MiddleName),
		Phone:            stringPtr(req.Phone),
		EmailVerified:    false,
		TwoFactorEnabled: false,
	}

	if err := h.repo.CreateUser(ctx, user, passwordHash); err != nil {
		h.deps.Logger.Error("Failed to create user for student", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create student",
		})
	}

	// Parse dates
	admissionDate, err := time.Parse("2006-01-02", req.AdmissionDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid admission date format (expected YYYY-MM-DD)",
		})
	}

	dateOfBirth, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid date of birth format (expected YYYY-MM-DD)",
		})
	}

	// Create student
	student := &Student{
		ID:                      uuid.New().String(),
		TenantID:                req.TenantID,
		UserID:                  userID,
		SchoolID:                req.SchoolID,
		AdmissionNumber:         req.AdmissionNumber,
		AdmissionDate:           admissionDate,
		DateOfBirth:             dateOfBirth,
		Gender:                  req.Gender,
		BloodGroup:              stringPtr(req.BloodGroup),
		Nationality:             stringPtr(req.Nationality),
		Religion:                stringPtr(req.Religion),
		Address:                 stringPtr(req.Address),
		City:                    stringPtr(req.City),
		State:                   stringPtr(req.State),
		PostalCode:              stringPtr(req.PostalCode),
		EmergencyContactName:    stringPtr(req.EmergencyContactName),
		EmergencyContactPhone:   stringPtr(req.EmergencyContactPhone),
		EmergencyContactRelation: stringPtr(req.EmergencyContactRelation),
		MedicalConditions:       stringPtr(req.MedicalConditions),
		Allergies:               stringPtr(req.Allergies),
		PreviousSchool:          stringPtr(req.PreviousSchool),
		CurrentGradeLevelID:     stringPtr(req.CurrentGradeLevelID),
		CurrentClassID:          stringPtr(req.CurrentClassID),
		IsActive:                true,
	}

	if err := h.repo.CreateStudent(ctx, student); err != nil {
		h.deps.Logger.Error("Failed to create student", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create student",
		})
	}

	// Publish student.created event
	if h.deps.EventBus != nil {
		eventData := map[string]interface{}{
			"student_id":       student.ID,
			"user_id":          userID,
			"admission_number": student.AdmissionNumber,
			"tenant_id":        student.TenantID,
		}
		if err := h.deps.EventBus.Publish("student.created", eventData); err != nil {
			h.deps.Logger.Warn("Failed to publish student.created event", zap.Error(err))
		}
	}

	h.deps.Logger.Info("Student created", zap.String("student_id", student.ID), zap.String("admission_number", student.AdmissionNumber))

	return c.Status(fiber.StatusCreated).JSON(ToStudentDTO(student, user))
}

// UpdateStudent updates an existing student
func (h *Handler) UpdateStudent(c *fiber.Ctx) error {
	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()
	studentID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Get existing student
	student, err := h.repo.GetStudent(ctx, studentID)
	if err != nil {
		if err == ErrStudentNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "STUDENT_NOT_FOUND",
				"message": "Student not found",
			})
		}
		h.deps.Logger.Error("Failed to get student", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update student",
		})
	}

	// Update fields
	if req.Address != "" {
		student.Address = &req.Address
	}
	if req.City != "" {
		student.City = &req.City
	}
	if req.State != "" {
		student.State = &req.State
	}
	if req.PostalCode != "" {
		student.PostalCode = &req.PostalCode
	}
	if req.EmergencyContactName != "" {
		student.EmergencyContactName = &req.EmergencyContactName
	}
	if req.EmergencyContactPhone != "" {
		student.EmergencyContactPhone = &req.EmergencyContactPhone
	}
	if req.EmergencyContactRelation != "" {
		student.EmergencyContactRelation = &req.EmergencyContactRelation
	}
	if req.MedicalConditions != "" {
		student.MedicalConditions = &req.MedicalConditions
	}
	if req.Allergies != "" {
		student.Allergies = &req.Allergies
	}
	if req.CurrentGradeLevelID != "" {
		student.CurrentGradeLevelID = &req.CurrentGradeLevelID
	}
	if req.CurrentClassID != "" {
		student.CurrentClassID = &req.CurrentClassID
	}
	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	// Update student
	if err := h.repo.UpdateStudent(ctx, student); err != nil {
		h.deps.Logger.Error("Failed to update student", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update student",
		})
	}

	// Get user data
	user, _ := h.repo.GetUser(ctx, student.UserID)

	h.deps.Logger.Info("Student updated", zap.String("student_id", student.ID))

	return c.JSON(ToStudentDTO(student, user))
}

// ListTeachers retrieves a paginated list of teachers
func (h *Handler) ListTeachers(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Parse pagination params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Get teachers
	teachers, total, err := h.repo.ListTeachers(ctx, perPage, offset)
	if err != nil {
		h.deps.Logger.Error("Failed to list teachers", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve teachers",
		})
	}

	// Convert to DTOs
	teacherDTOs := make([]*TeacherDTO, 0, len(teachers))
	for _, teacher := range teachers {
		// Optionally get user data
		user, _ := h.repo.GetUser(ctx, teacher.UserID)
		teacherDTOs = append(teacherDTOs, ToTeacherDTO(teacher, user))
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return c.JSON(&PaginatedResponse{
		Data:       teacherDTOs,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetTeacher retrieves a single teacher by ID
func (h *Handler) GetTeacher(c *fiber.Ctx) error {
	ctx := c.Context()
	teacherID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Get teacher
	teacher, err := h.repo.GetTeacher(ctx, teacherID)
	if err != nil {
		if err == ErrTeacherNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "TEACHER_NOT_FOUND",
				"message": "Teacher not found",
			})
		}
		h.deps.Logger.Error("Failed to get teacher", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve teacher",
		})
	}

	// Get user data
	user, _ := h.repo.GetUser(ctx, teacher.UserID)

	return c.JSON(ToTeacherDTO(teacher, user))
}

// CreateTeacher creates a new teacher (and associated user)
func (h *Handler) CreateTeacher(c *fiber.Ctx) error {
	var req CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, req.TenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Hash password
	passwordHash, err := h.deps.Crypto.HashPassword(req.Password)
	if err != nil {
		h.deps.Logger.Error("Failed to hash password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create teacher",
		})
	}

	// Create user first
	userID := uuid.New().String()
	user := &User{
		ID:               userID,
		TenantID:         req.TenantID,
		Email:            req.Email,
		Role:             "teacher",
		Status:           "active",
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		MiddleName:       stringPtr(req.MiddleName),
		Phone:            stringPtr(req.Phone),
		EmailVerified:    false,
		TwoFactorEnabled: false,
	}

	if err := h.repo.CreateUser(ctx, user, passwordHash); err != nil {
		h.deps.Logger.Error("Failed to create user for teacher", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create teacher",
		})
	}

	// Parse dates
	dateOfBirth, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid date of birth format (expected YYYY-MM-DD)",
		})
	}

	dateOfJoining, err := time.Parse("2006-01-02", req.DateOfJoining)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid date of joining format (expected YYYY-MM-DD)",
		})
	}

	// Create teacher
	teacher := &Teacher{
		ID:              uuid.New().String(),
		TenantID:        req.TenantID,
		UserID:          userID,
		SchoolID:        req.SchoolID,
		EmployeeID:      req.EmployeeID,
		DepartmentID:    stringPtr(req.DepartmentID),
		DateOfBirth:     dateOfBirth,
		Gender:          req.Gender,
		DateOfJoining:   dateOfJoining,
		Qualification:   stringPtr(req.Qualification),
		Specialization:  stringPtr(req.Specialization),
		ExperienceYears: req.ExperienceYears,
		EmploymentType:  stringPtr(req.EmploymentType),
		Salary:          float64Ptr(req.Salary),
		IsActive:        true,
	}

	if err := h.repo.CreateTeacher(ctx, teacher); err != nil {
		h.deps.Logger.Error("Failed to create teacher", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create teacher",
		})
	}

	// Publish teacher.created event
	if h.deps.EventBus != nil {
		eventData := map[string]interface{}{
			"teacher_id":  teacher.ID,
			"user_id":     userID,
			"employee_id": teacher.EmployeeID,
			"tenant_id":   teacher.TenantID,
		}
		if err := h.deps.EventBus.Publish("teacher.created", eventData); err != nil {
			h.deps.Logger.Warn("Failed to publish teacher.created event", zap.Error(err))
		}
	}

	h.deps.Logger.Info("Teacher created", zap.String("teacher_id", teacher.ID), zap.String("employee_id", teacher.EmployeeID))

	return c.Status(fiber.StatusCreated).JSON(ToTeacherDTO(teacher, user))
}

// UpdateTeacher updates an existing teacher
func (h *Handler) UpdateTeacher(c *fiber.Ctx) error {
	var req UpdateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()
	teacherID := c.Params("id")

	// Extract tenant from context
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Tenant not found",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, tenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Get existing teacher
	teacher, err := h.repo.GetTeacher(ctx, teacherID)
	if err != nil {
		if err == ErrTeacherNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "TEACHER_NOT_FOUND",
				"message": "Teacher not found",
			})
		}
		h.deps.Logger.Error("Failed to get teacher", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update teacher",
		})
	}

	// Update fields
	if req.DepartmentID != "" {
		teacher.DepartmentID = &req.DepartmentID
	}
	if req.Qualification != "" {
		teacher.Qualification = &req.Qualification
	}
	if req.Specialization != "" {
		teacher.Specialization = &req.Specialization
	}
	if req.ExperienceYears > 0 {
		teacher.ExperienceYears = req.ExperienceYears
	}
	if req.EmploymentType != "" {
		teacher.EmploymentType = &req.EmploymentType
	}
	if req.Salary > 0 {
		teacher.Salary = &req.Salary
	}
	if req.IsActive != nil {
		teacher.IsActive = *req.IsActive
	}

	// Update teacher
	if err := h.repo.UpdateTeacher(ctx, teacher); err != nil {
		h.deps.Logger.Error("Failed to update teacher", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update teacher",
		})
	}

	// Get user data
	user, _ := h.repo.GetUser(ctx, teacher.UserID)

	h.deps.Logger.Info("Teacher updated", zap.String("teacher_id", teacher.ID))

	return c.JSON(ToTeacherDTO(teacher, user))
}

// Helper functions

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func float64Ptr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}
