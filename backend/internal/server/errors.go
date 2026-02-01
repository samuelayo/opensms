package server

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// AppError represents a custom application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
)

// NewAppError creates a new application error
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Common errors
var (
	ErrValidation      = NewAppError(ErrCodeValidation, "Validation failed", fiber.StatusBadRequest)
	ErrNotFound        = NewAppError(ErrCodeNotFound, "Resource not found", fiber.StatusNotFound)
	ErrUnauthorized    = NewAppError(ErrCodeUnauthorized, "Unauthorized", fiber.StatusUnauthorized)
	ErrForbidden       = NewAppError(ErrCodeForbidden, "Forbidden", fiber.StatusForbidden)
	ErrConflict        = NewAppError(ErrCodeConflict, "Resource already exists", fiber.StatusConflict)
	ErrInternal        = NewAppError(ErrCodeInternal, "Internal server error", fiber.StatusInternalServerError)
	ErrBadRequest      = NewAppError(ErrCodeBadRequest, "Bad request", fiber.StatusBadRequest)
	ErrTooManyRequests = NewAppError(ErrCodeTooManyRequests, "Too many requests", fiber.StatusTooManyRequests)
)

// ErrorHandler is the global error handler for Fiber
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default to 500
	code := fiber.StatusInternalServerError
	response := fiber.Map{
		"error": "Internal server error",
	}

	// Handle AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		code = appErr.Status
		response = fiber.Map{
			"code":    appErr.Code,
			"message": appErr.Message,
		}
		if appErr.Details != "" {
			response["details"] = appErr.Details
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		// Database not found error
		code = fiber.StatusNotFound
		response = fiber.Map{
			"code":    ErrCodeNotFound,
			"message": "Resource not found",
		}
	} else {
		// Fiber errors
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			response = fiber.Map{
				"code":    fmt.Sprintf("HTTP_%d", fiberErr.Code),
				"message": fiberErr.Message,
			}
		} else {
			// Log unknown errors
			logger, _ := zap.NewProduction()
			logger.Error("Unhandled error",
				zap.Error(err),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)

			response = fiber.Map{
				"code":    ErrCodeInternal,
				"message": "An unexpected error occurred",
			}
		}
	}

	// Don't expose internal errors in production
	if code == fiber.StatusInternalServerError {
		response = fiber.Map{
			"code":    ErrCodeInternal,
			"message": "An unexpected error occurred",
		}
	}

	return c.Status(code).JSON(response)
}

// ValidationError creates a validation error with details
func ValidationError(details string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: "Validation failed",
		Details: details,
		Status:  fiber.StatusBadRequest,
	}
}

// NotFoundError creates a not found error with custom message
func NotFoundError(resource string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Status:  fiber.StatusNotFound,
	}
}

// ConflictError creates a conflict error with custom message
func ConflictError(details string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: "Resource already exists",
		Details: details,
		Status:  fiber.StatusConflict,
	}
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
		Status:  fiber.StatusUnauthorized,
	}
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
		Status:  fiber.StatusForbidden,
	}
}
