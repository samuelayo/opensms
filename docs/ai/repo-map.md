# Repository Map

## What This Repo Is

- Production-grade school management system for primary through university level institutions
- Built as a modular monolith with microservices evolution path, designed for extreme scale (millions of users, thousands of institutions)
- Multi-tenant SaaS platform with row-level security, supporting 10K+ institutions on shared infrastructure

## Tech Stack

- **Backend**: Go 1.22+ with Fiber web framework, Clean Architecture, Domain-Driven Design
- **Frontend**: Vue.js 3 (Composition API), Pinia state management, TypeScript, Vite, TailwindCSS
- **Databases**: PostgreSQL 16+ (primary with multi-tenancy), Redis 7+ (cache/sessions)
- **Event Bus**: NATS JetStream for async event-driven communication
- **Storage**: MinIO/S3 for object storage
- **Observability**: Prometheus (metrics), Jaeger (tracing), Zap (structured logging)

## High-Level Module Map

```
opensms/
├── backend/                           # Go backend (modular monolith)
│   ├── cmd/api/                      # HTTP API server entrypoint
│   ├── internal/
│   │   ├── infrastructure/           # Infrastructure layer (databases, caching, observability)
│   │   ├── modules/                  # Business domains (auth, users, academic)
│   │   ├── server/                   # HTTP server, routing, middleware
│   │   └── shared/                   # Shared kernel (security, utilities)
│   ├── migrations/                   # Database schema migrations (Goose)
│   └── tests/                        # Integration tests
├── frontend/
│   └── admin-portal/                 # Vue.js admin dashboard (only portal currently implemented)
│       ├── src/
│       │   ├── components/           # Reusable Vue components
│       │   ├── stores/               # Pinia stores (state management)
│       │   ├── services/             # API clients, utilities
│       │   ├── views/                # Page-level components
│       │   └── router/               # Vue Router configuration
│       └── e2e/                      # Playwright E2E tests
├── infrastructure/                   # Infrastructure as Code
│   └── prometheus/                   # Prometheus monitoring config
├── docker-compose.yml                # Local development environment
└── docs/                            # Documentation
```

## Key Choke Points

### Request Entry Points (HTTP)
1. **`backend/cmd/api/main.go:30-193`** - Main HTTP server initialization and startup
2. **`backend/internal/server/server.go:57-179`** - Route registration for all API endpoints
3. **`backend/internal/server/middleware.go:14-105`** - Auth middleware (JWT validation, tenant context)
4. **`backend/internal/server/middleware.go:107-149`** - Permission middleware (RBAC enforcement)

### Database Persistence
5. **`backend/internal/infrastructure/database/postgres.go:25-119`** - PostgreSQL connection management with multi-tenancy
6. **`backend/internal/infrastructure/database/postgres.go:121-141`** - Tenant context setting (row-level security via `app.current_tenant`)
7. **`backend/migrations/001_initial_schema.sql`** - Core schema with users, roles, permissions
8. **`backend/migrations/002_students_and_enrollment.sql`** - Academic entities (students, classes, grades)

### Caching Layer
9. **`backend/internal/infrastructure/cache/redis.go:25-65`** - Redis client initialization and operations
10. **`backend/internal/infrastructure/cache/redis.go:142-212`** - Rate limiter implementation (sliding window)

### Event Processing
11. **`backend/internal/infrastructure/eventbus/nats.go:23-62`** - NATS connection and pub/sub
12. **Module handlers** - Publish events after state changes (e.g., student enrollment triggers emails)

### External Service Calls
13. **`backend/internal/infrastructure/storage/minio.go:24-78`** - MinIO/S3 object storage client
14. **`backend/internal/infrastructure/observability/tracing.go:18-61`** - Jaeger tracing exporter
15. **`backend/internal/infrastructure/observability/metrics.go:14-72`** - Prometheus metrics collection

### Authentication & Authorization
16. **`backend/internal/shared/security/jwt.go:40-98`** - JWT token generation and validation
17. **`backend/internal/shared/security/rbac.go:51-200`** - RBAC permission checking (13 roles defined)
18. **`backend/internal/shared/security/crypto.go:18-74`** - Encryption (AES-256) and password hashing (bcrypt)

### API Modules (Business Logic)
19. **`backend/internal/modules/auth/handler.go`** - Authentication endpoints (login, register, refresh, 2FA)
20. **`backend/internal/modules/users/handler.go`** - User/student/teacher CRUD operations
21. **`backend/internal/modules/academic/handler.go`** - Academic operations (grades, attendance, classes)

### Frontend Entry Points
22. **`frontend/admin-portal/src/main.ts`** - Vue.js app initialization
23. **`frontend/admin-portal/src/router/index.ts`** - Route definitions and navigation guards
24. **`frontend/admin-portal/src/services/api.ts`** - Axios HTTP client with interceptors (auto token refresh)
25. **`frontend/admin-portal/src/stores/auth.ts`** - Authentication state management (Pinia)

## Dependency Notes

### Internal Libraries
- **`backend/internal/shared/security/`** - Used by all modules for JWT, RBAC, encryption
- **`backend/internal/infrastructure/`** - Injected into all modules via dependency injection pattern
- **`backend/internal/server/`** - Server package coordinates all modules and infrastructure

### External Dependencies (Backend)
- **Fiber v2** (`github.com/gofiber/fiber/v2`) - HTTP framework
- **pgx/v5** (`github.com/jackc/pgx/v5`) - PostgreSQL driver
- **go-redis/v9** - Redis client
- **nats.go** - NATS messaging client
- **jwt/v5** (`github.com/golang-jwt/jwt/v5`) - JWT implementation
- **zap** (`go.uber.org/zap`) - Structured logging
- **viper** (`github.com/spf13/viper`) - Configuration management
- **goose** (`github.com/pressly/goose/v3`) - Database migrations
- **testify** (`github.com/stretchr/testify`) - Testing framework

### External Dependencies (Frontend)
- **Vue 3** - UI framework
- **Pinia** - State management
- **Vue Router** - Routing
- **Axios** - HTTP client
- **Vite** - Build tool
- **Vitest** - Testing framework
- **Playwright** - E2E testing

### Generated Code
- No code generation currently used
- Future consideration for OpenAPI/Swagger spec generation

### Configuration Sources
- **Environment variables** - Primary configuration method (see `backend/internal/infrastructure/config/config.go:104-210`)
- **YAML config files** (optional) - Located at `./config/config.yaml` or `.`
- **Defaults** - Hardcoded in `backend/internal/infrastructure/config/config.go:212-267`
- **Docker Compose** - `docker-compose.yml` for local development environment variables

### Build System
- **Backend**: Go modules (`go.mod`), Makefile for common tasks
- **Frontend**: npm/package.json, Vite for bundling
- **CI/CD**: GitHub Actions (`.github/workflows/test.yml`)
