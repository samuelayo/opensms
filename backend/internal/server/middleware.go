package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// authMiddleware validates JWT tokens and sets user context
func (s *Server) authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := s.jwt.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Set user context
		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions)

		// Set tenant in context for row-level security
		ctx := database.WithTenant(c.Context(), claims.TenantID)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// tenantMiddleware extracts tenant from subdomain or header
func (s *Server) tenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tenantID string

		// Try to get tenant from header first (for API clients)
		tenantID = c.Get("X-Tenant-ID")

		// If not in header, try to extract from subdomain
		if tenantID == "" {
			host := c.Hostname()
			// Extract subdomain (e.g., school1.opensms.com -> school1)
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				tenantID = parts[0]
			}
		}

		if tenantID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant identification required",
			})
		}

		c.Locals("tenant_id", tenantID)
		return c.Next()
	}
}

// requirePermission creates middleware to check for specific permission
func (s *Server) requirePermission(permission security.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user role from context
		roleStr, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		role := security.Role(roleStr)

		// Check if role has permission
		if !s.rbac.HasPermission(role, permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// requireAnyPermission creates middleware to check for any of the specified permissions
func (s *Server) requireAnyPermission(permissions []security.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleStr, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		role := security.Role(roleStr)

		if !s.rbac.HasAnyPermission(role, permissions) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// requireRole creates middleware to check for specific role
func (s *Server) requireRole(requiredRole security.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleStr, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		role := security.Role(roleStr)

		if role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient role privileges",
			})
		}

		return c.Next()
	}
}

// rateLimitMiddleware implements rate limiting per user
func (s *Server) rateLimitMiddleware(limit int, window int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID or IP
		key := c.IP()
		if userID, ok := c.Locals("user_id").(string); ok {
			key = userID
		}

		// Check rate limit using Redis
		// Implementation would use cache.RateLimiter
		// For now, just pass through
		return c.Next()
	}
}

// auditLogMiddleware logs important actions for compliance
func (s *Server) auditLogMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log the request
		userID, _ := c.Locals("user_id").(string)
		tenantID, _ := c.Locals("tenant_id").(string)

		s.deps.Logger.Info("API Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.String("ip", c.IP()),
		)

		return c.Next()
	}
}
