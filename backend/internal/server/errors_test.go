package server

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError("TEST_CODE", "Test message", fiber.StatusBadRequest)
	assert.Equal(t, "TEST_CODE", err.Code)
	assert.Equal(t, "Test message", err.Message)
	assert.Equal(t, fiber.StatusBadRequest, err.Status)
	assert.Equal(t, "Test message", err.Error())
}

func TestValidationError(t *testing.T) {
	err := ValidationError("Email is required")
	assert.Equal(t, ErrCodeValidation, err.Code)
	assert.Equal(t, "Validation failed", err.Message)
	assert.Equal(t, "Email is required", err.Details)
	assert.Equal(t, fiber.StatusBadRequest, err.Status)
}

func TestNotFoundError(t *testing.T) {
	err := NotFoundError("Student")
	assert.Equal(t, ErrCodeNotFound, err.Code)
	assert.Equal(t, "Student not found", err.Message)
	assert.Equal(t, fiber.StatusNotFound, err.Status)
}

func TestConflictError(t *testing.T) {
	err := ConflictError("Email already exists")
	assert.Equal(t, ErrCodeConflict, err.Code)
	assert.Equal(t, "Resource already exists", err.Message)
	assert.Equal(t, "Email already exists", err.Details)
	assert.Equal(t, fiber.StatusConflict, err.Status)
}

func TestUnauthorizedError(t *testing.T) {
	err := UnauthorizedError("Invalid credentials")
	assert.Equal(t, ErrCodeUnauthorized, err.Code)
	assert.Equal(t, "Invalid credentials", err.Message)
	assert.Equal(t, fiber.StatusUnauthorized, err.Status)
}

func TestForbiddenError(t *testing.T) {
	err := ForbiddenError("Access denied")
	assert.Equal(t, ErrCodeForbidden, err.Code)
	assert.Equal(t, "Access denied", err.Message)
	assert.Equal(t, fiber.StatusForbidden, err.Status)
}

func TestErrorHandler_AppError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return ValidationError("Invalid input")
	})

	req := fiber.AcquireRequest()
	req.SetRequestURI("/test")
	req.Header.SetMethod("GET")
	defer fiber.ReleaseRequest(req)

	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)

	err := app.Test(req, -1)
	assert.NoError(t, err)
}

func TestErrorHandler_PgxNoRows(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return pgx.ErrNoRows
	})

	req := fiber.AcquireRequest()
	req.SetRequestURI("/test")
	req.Header.SetMethod("GET")
	defer fiber.ReleaseRequest(req)

	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)

	err := app.Test(req, -1)
	assert.NoError(t, err)
}

func TestErrorHandler_FiberError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusTeapot, "I'm a teapot")
	})

	req := fiber.AcquireRequest()
	req.SetRequestURI("/test")
	req.Header.SetMethod("GET")
	defer fiber.ReleaseRequest(req)

	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)

	err := app.Test(req, -1)
	assert.NoError(t, err)
}

func TestErrorHandler_GenericError(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return errors.New("unexpected error")
	})

	req := fiber.AcquireRequest()
	req.SetRequestURI("/test")
	req.Header.SetMethod("GET")
	defer fiber.ReleaseRequest(req)

	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)

	err := app.Test(req, -1)
	assert.NoError(t, err)
}
