package server

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/infrastructure/storage"
	"github.com/samuelayo/opensms/internal/modules/academic"
	"github.com/samuelayo/opensms/internal/modules/auth"
	"github.com/samuelayo/opensms/internal/modules/users"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// Dependencies holds all infrastructure dependencies
type Dependencies struct {
	Config   *config.Config
	DB       *database.DB
	Cache    *cache.RedisClient
	EventBus *eventbus.NATSConnection
	Storage  *storage.MinIOClient
	Logger   *zap.Logger
}

// Server represents the HTTP server
type Server struct {
	deps Dependencies
	jwt  *security.JWTManager
	rbac *security.RBAC
}

// NewServer creates a new server instance
func NewServer(deps Dependencies) *Server {
	// Initialize JWT manager
	jwt := security.NewJWTManager(
		deps.Config.JWT.AccessSecret,
		deps.Config.JWT.RefreshSecret,
		deps.Config.JWT.AccessTokenExpiry,
		deps.Config.JWT.RefreshTokenExpiry,
		deps.Config.JWT.Issuer,
	)

	// Initialize RBAC
	rbac := security.NewRBAC()

	return &Server{
		deps: deps,
		jwt:  jwt,
		rbac: rbac,
	}
}

// RegisterRoutes registers all API routes
func (s *Server) RegisterRoutes(app *fiber.App) {
	// API version 1
	v1 := app.Group("/api/v1")

	// Public routes (no authentication required)
	public := v1.Group("/public")
	public.Get("/health", s.healthCheck)
	public.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	// Initialize modules
	s.registerAuthModule(v1)
	s.registerUsersModule(v1)
	s.registerAcademicModule(v1)
	// Additional modules will be registered here

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	})
}

// registerAuthModule registers authentication routes
func (s *Server) registerAuthModule(router fiber.Router) {
	authHandler := auth.NewHandler(auth.HandlerDeps{
		DB:       s.deps.DB,
		Cache:    s.deps.Cache,
		EventBus: s.deps.EventBus,
		JWT:      s.jwt,
		Config:   s.deps.Config,
		Logger:   s.deps.Logger,
	})

	authGroup := router.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Post("/refresh", authHandler.RefreshToken)
	authGroup.Post("/forgot-password", authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authHandler.ResetPassword)
	authGroup.Get("/verify-email/:token", authHandler.VerifyEmail)

	// Protected routes
	protected := authGroup.Group("", s.authMiddleware())
	protected.Get("/me", authHandler.GetCurrentUser)
	protected.Put("/change-password", authHandler.ChangePassword)
	protected.Post("/enable-2fa", authHandler.Enable2FA)
	protected.Post("/verify-2fa", authHandler.Verify2FA)
}

// registerUsersModule registers user management routes
func (s *Server) registerUsersModule(router fiber.Router) {
	crypto := security.NewCrypto(s.deps.Config.Security.EncryptionKey)

	usersHandler := users.NewHandler(users.HandlerDeps{
		DB:       s.deps.DB,
		Cache:    s.deps.Cache,
		EventBus: s.deps.EventBus,
		RBAC:     s.rbac,
		Logger:   s.deps.Logger,
		Crypto:   crypto,
	})

	usersGroup := router.Group("/users", s.authMiddleware())
	usersGroup.Get("/", s.requirePermission(security.PermUserRead), usersHandler.ListUsers)
	usersGroup.Get("/:id", s.requirePermission(security.PermUserRead), usersHandler.GetUser)
	usersGroup.Post("/", s.requirePermission(security.PermUserCreate), usersHandler.CreateUser)
	usersGroup.Put("/:id", s.requirePermission(security.PermUserUpdate), usersHandler.UpdateUser)
	usersGroup.Delete("/:id", s.requirePermission(security.PermUserDelete), usersHandler.DeleteUser)

	// Students
	studentsGroup := router.Group("/students", s.authMiddleware())
	studentsGroup.Get("/", s.requirePermission(security.PermStudentRead), usersHandler.ListStudents)
	studentsGroup.Get("/:id", s.requirePermission(security.PermStudentRead), usersHandler.GetStudent)
	studentsGroup.Post("/", s.requirePermission(security.PermStudentCreate), usersHandler.CreateStudent)
	studentsGroup.Put("/:id", s.requirePermission(security.PermStudentUpdate), usersHandler.UpdateStudent)

	// Teachers
	teachersGroup := router.Group("/teachers", s.authMiddleware())
	teachersGroup.Get("/", s.requirePermission(security.PermUserRead), usersHandler.ListTeachers)
	teachersGroup.Get("/:id", s.requirePermission(security.PermUserRead), usersHandler.GetTeacher)
	teachersGroup.Post("/", s.requirePermission(security.PermUserCreate), usersHandler.CreateTeacher)
	teachersGroup.Put("/:id", s.requirePermission(security.PermUserUpdate), usersHandler.UpdateTeacher)
}

// registerAcademicModule registers academic management routes
func (s *Server) registerAcademicModule(router fiber.Router) {
	academicHandler := academic.NewHandler(academic.HandlerDeps{
		DB:       s.deps.DB,
		Cache:    s.deps.Cache,
		EventBus: s.deps.EventBus,
		RBAC:     s.rbac,
		Logger:   s.deps.Logger,
	})

	// Academic years
	yearsGroup := router.Group("/academic-years", s.authMiddleware())
	yearsGroup.Get("/", academicHandler.ListAcademicYears)
	yearsGroup.Post("/", s.requirePermission(security.PermSchoolUpdate), academicHandler.CreateAcademicYear)

	// Classes
	classesGroup := router.Group("/classes", s.authMiddleware())
	classesGroup.Get("/", academicHandler.ListClasses)
	classesGroup.Get("/:id", academicHandler.GetClass)
	classesGroup.Post("/", s.requirePermission(security.PermClassCreate), academicHandler.CreateClass)
	classesGroup.Put("/:id", s.requirePermission(security.PermClassManage), academicHandler.UpdateClass)

	// Subjects
	subjectsGroup := router.Group("/subjects", s.authMiddleware())
	subjectsGroup.Get("/", academicHandler.ListSubjects)
	subjectsGroup.Post("/", s.requirePermission(security.PermClassCreate), academicHandler.CreateSubject)

	// Grades
	gradesGroup := router.Group("/grades", s.authMiddleware())
	gradesGroup.Get("/student/:student_id", s.requirePermission(security.PermGradeRead), academicHandler.GetStudentGrades)
	gradesGroup.Post("/", s.requirePermission(security.PermGradeCreate), academicHandler.CreateGrade)
	gradesGroup.Put("/:id", s.requirePermission(security.PermGradeUpdate), academicHandler.UpdateGrade)

	// Attendance
	attendanceGroup := router.Group("/attendance", s.authMiddleware())
	attendanceGroup.Get("/class/:class_id", s.requirePermission(security.PermAttendanceRead), academicHandler.GetClassAttendance)
	attendanceGroup.Get("/student/:student_id/stats", s.requirePermission(security.PermAttendanceRead), academicHandler.GetAttendanceStats)
	attendanceGroup.Post("/", s.requirePermission(security.PermAttendanceMark), academicHandler.MarkAttendance)
	attendanceGroup.Post("/bulk", s.requirePermission(security.PermAttendanceMark), academicHandler.BulkMarkAttendance)

	// Assignments
	assignmentsGroup := router.Group("/assignments", s.authMiddleware())
	assignmentsGroup.Get("/", s.requirePermission(security.PermGradeRead), academicHandler.ListAssignments)
	assignmentsGroup.Post("/", s.requirePermission(security.PermGradeCreate), academicHandler.CreateAssignment)
}

// healthCheck returns the health status of the server
func (s *Server) healthCheck(c *fiber.Ctx) error {
	// Check database health
	dbHealth := "healthy"
	if err := s.deps.DB.Health(c.Context()); err != nil {
		dbHealth = "unhealthy: " + err.Error()
	}

	// Check cache health
	cacheHealth := "healthy"
	if err := s.deps.Cache.Health(c.Context()); err != nil {
		cacheHealth = "unhealthy: " + err.Error()
	}

	// Check event bus health
	eventBusHealth := "healthy"
	if err := s.deps.EventBus.Health(); err != nil {
		eventBusHealth = "unhealthy: " + err.Error()
	}

	return c.JSON(fiber.Map{
		"status":    "healthy",
		"database":  dbHealth,
		"cache":     cacheHealth,
		"event_bus": eventBusHealth,
	})
}
