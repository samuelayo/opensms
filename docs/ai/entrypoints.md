# Entrypoints

**⚠️ CRITICAL**: All routes are **REGISTERED** but return **PLACEHOLDER** responses. See `docs/ai/implementation-status.md` for details.

- Auth endpoints return: `{"message": "Login endpoint - to be implemented"}` with 501 status
- Users/Students/Teachers endpoints return: `{"message": "List users - to be implemented"}`
- Academic endpoints return: `{"message": "Get grades - to be implemented"}`
- **ONLY** health checks (`/health`, `/api/v1/public/health`) and metrics (`/metrics`) are functional

## Services/Binaries

### API Server
- **Name**: OpenSMS API Server
- **Path**: `backend/cmd/api/main.go`
- **How Started**:
  - Development: `make run` or `make dev` (with hot reload via air)
  - Production: `./bin/api` (after `make build`)
  - Docker: `docker-compose up` (not containerized yet, runs directly)
- **Listen Address**: `SERVER_HOST:SERVER_PORT` (default: `0.0.0.0:8080`)
- **Health Check**: `GET /health` at `backend/cmd/api/main.go:160-165`
- **Metrics**: `GET /metrics` at `backend/cmd/api/main.go:168`

## HTTP/RPC Entrypoints

### Router Registration
- **Main Router**: `backend/internal/server/server.go:57-79` - `RegisterRoutes()` method
- **Base Path**: `/api/v1` (defined at `backend/internal/server/server.go:59`)

### Public Routes (No Auth Required)
- `GET /api/v1/public/health` - Health check (`backend/internal/server/server.go:63`)
- `GET /api/v1/public/ping` - Ping test (`backend/internal/server/server.go:64-66`)
- `GET /health` - Top-level health check (`backend/cmd/api/main.go:160-165`)
- `GET /metrics` - Prometheus metrics (`backend/cmd/api/main.go:168`)

### Authentication Module
- **Handler**: `backend/internal/modules/auth/handler.go`
- **Registration**: `backend/internal/server/server.go:83-108`
- **Routes**:
  - `POST /api/v1/auth/register` - User registration
  - `POST /api/v1/auth/login` - User login
  - `POST /api/v1/auth/logout` - User logout
  - `POST /api/v1/auth/refresh` - Refresh access token
  - `POST /api/v1/auth/forgot-password` - Forgot password
  - `POST /api/v1/auth/reset-password` - Reset password
  - `GET /api/v1/auth/verify-email/:token` - Email verification
  - `GET /api/v1/auth/me` - Get current user (requires auth)
  - `PUT /api/v1/auth/change-password` - Change password (requires auth)
  - `POST /api/v1/auth/enable-2fa` - Enable 2FA (requires auth)
  - `POST /api/v1/auth/verify-2fa` - Verify 2FA (requires auth)

### Users Module
- **Handler**: `backend/internal/modules/users/handler.go`
- **Registration**: `backend/internal/server/server.go:111-140`
- **Routes** (all require authentication):
  - `GET /api/v1/users` - List users (requires `user:read` permission)
  - `GET /api/v1/users/:id` - Get user (requires `user:read`)
  - `POST /api/v1/users` - Create user (requires `user:create`)
  - `PUT /api/v1/users/:id` - Update user (requires `user:update`)
  - `DELETE /api/v1/users/:id` - Delete user (requires `user:delete`)
  - `GET /api/v1/students` - List students (requires `student:read`)
  - `GET /api/v1/students/:id` - Get student (requires `student:read`)
  - `POST /api/v1/students` - Create student (requires `student:create`)
  - `PUT /api/v1/students/:id` - Update student (requires `student:update`)
  - `GET /api/v1/teachers` - List teachers (requires `user:read`)
  - `GET /api/v1/teachers/:id` - Get teacher (requires `user:read`)
  - `POST /api/v1/teachers` - Create teacher (requires `user:create`)
  - `PUT /api/v1/teachers/:id` - Update teacher (requires `user:update`)

### Academic Module
- **Handler**: `backend/internal/modules/academic/handler.go`
- **Registration**: `backend/internal/server/server.go:143-179`
- **Routes** (all require authentication):
  - `GET /api/v1/academic-years` - List academic years
  - `POST /api/v1/academic-years` - Create academic year (requires `school:update`)
  - `GET /api/v1/classes` - List classes
  - `GET /api/v1/classes/:id` - Get class
  - `POST /api/v1/classes` - Create class (requires `class:create`)
  - `PUT /api/v1/classes/:id` - Update class (requires `class:manage`)
  - `GET /api/v1/subjects` - List subjects
  - `POST /api/v1/subjects` - Create subject (requires `class:create`)
  - `GET /api/v1/grades/student/:student_id` - Get student grades (requires `grade:read`)
  - `POST /api/v1/grades` - Create grade (requires `grade:create`)
  - `PUT /api/v1/grades/:id` - Update grade (requires `grade:update`)
  - `GET /api/v1/attendance/class/:class_id` - Get class attendance (requires `attendance:read`)
  - `POST /api/v1/attendance` - Mark attendance (requires `attendance:mark`)

### 404 Handler
- **Location**: `backend/internal/server/server.go:75-79`
- **Response**: `{"error": "Route not found"}` with 404 status

## Background Workers/Consumers

### Event Consumers
- **Connection**: `backend/internal/infrastructure/eventbus/nats.go:23-62`
- **Topics/Events**: Defined per module, not centralized (future enhancement)
- **Example Event Flow** (from `ARCHITECTURE.md:82-93`):
  - `student.enrolled` - Triggers welcome email, fee invoice, teacher notifications, analytics updates
- **Implementation**: Modules publish events via `EventBus.Publish()`, consumers subscribe via `EventBus.Subscribe()`

### Scheduled Jobs
- **Status**: Not implemented yet
- **Future**: Cron jobs for report generation, cleanup tasks, reminder emails

## CLI Commands

### Available via Makefile
- **Location**: `backend/Makefile`
- **Commands**:
  - `make build` - Build the API binary to `bin/api`
  - `make run` - Build and run the API server
  - `make dev` - Run with hot reload (using air)
  - `make test` - Run all tests with coverage
  - `make lint` - Run golangci-lint
  - `make clean` - Remove build artifacts
  - `make migrate-up` - Run database migrations up
  - `make migrate-down` - Run database migrations down
  - `make migrate-status` - Show migration status
  - `make migrate-create NAME=<name>` - Create new migration
  - `make docker-up` - Start Docker services (postgres, redis, nats, minio)
  - `make docker-down` - Stop Docker services
  - `make docker-logs` - Show Docker logs
  - `make setup` - Initial development environment setup
  - `make fmt` - Format Go code
  - `make tidy` - Tidy Go dependencies

### Database Migration Tool
- **Tool**: Goose (`github.com/pressly/goose/v3`)
- **Migrations Dir**: `backend/migrations/`
- **Connection String**: Configured in Makefile targets (lines 59, 63, 67)
- **Commands**: Wrapped in Makefile (migrate-up, migrate-down, migrate-status, migrate-create)

## Frontend Entrypoints

### Main Application
- **Entry File**: `frontend/admin-portal/src/main.ts`
- **Dev Server**: `npm run dev` (Vite dev server on `http://localhost:5173`)
- **Build**: `npm run build` (outputs to `dist/`)
- **Preview**: `npm run preview` (preview production build)

### Routes (Vue Router)
- **Router File**: `frontend/admin-portal/src/router/index.ts`
- **Navigation Guards**: Authentication check on all routes except login
- **Routes**:
  - `/login` - Login page
  - `/` - Dashboard (requires auth)
  - Additional routes defined in router file

### API Client
- **File**: `frontend/admin-portal/src/services/api.ts`
- **Base URL**: Configured via environment variable
- **Interceptors**: Auto token refresh on 401 responses

## Observability Endpoints

### Health Checks
- **Detailed Health**: `GET /api/v1/public/health` - Checks DB, cache, event bus (`backend/internal/server/server.go:182-207`)
- **Simple Health**: `GET /health` - Returns `{"status": "healthy"}` (`backend/cmd/api/main.go:160-165`)

### Metrics
- **Endpoint**: `GET /metrics` (`backend/cmd/api/main.go:168`)
- **Format**: Prometheus text format
- **Metrics**: Defined in `backend/internal/infrastructure/observability/metrics.go:14-72`

### Tracing
- **Protocol**: OpenTelemetry/Jaeger
- **Exporter**: `backend/internal/infrastructure/observability/tracing.go:18-61`
- **Endpoint**: `JAEGER_ENDPOINT` (default: `http://localhost:14268/api/traces`)
