# Run and Test

## How to Build

### Backend (Go)

**Prerequisites**:
- Go 1.22+ installed
- Make utility installed

**Build Commands**:
```bash
cd backend

# Build binary to bin/api
make build

# Build manually
go build -o bin/api cmd/api/main.go
```

**Output**: Compiled binary at `backend/bin/api`

### Frontend (Vue.js)

**Prerequisites**:
- Node.js 20+ installed
- npm or yarn

**Build Commands**:
```bash
cd frontend/admin-portal

# Install dependencies
npm install

# Build for production
npm run build

# Type check
npm run type-check
```

**Output**: Production build in `frontend/admin-portal/dist/`

## How to Run Locally

### Quick Start (All Services)

**Prerequisites**:
- Docker and Docker Compose installed
- Go 1.22+ and Node.js 20+ installed

**Steps**:

1. **Start infrastructure services** (PostgreSQL, Redis, NATS, MinIO):
```bash
# From repo root
docker-compose up -d

# Verify services are running
docker-compose ps
```

2. **Run database migrations**:
```bash
cd backend
make migrate-up
```

3. **Start backend API**:
```bash
cd backend

# Development mode with hot reload (recommended)
make dev

# OR standard run
make run

# OR run directly
go run cmd/api/main.go
```

Backend will start on `http://localhost:8080`

4. **Start frontend** (in separate terminal):
```bash
cd frontend/admin-portal

# Install dependencies (first time only)
npm install

# Start dev server
npm run dev
```

Frontend will start on `http://localhost:5173`

### Environment Variables

**Backend**: Create `.env` in `backend/` directory (optional, env vars can be set directly):
```bash
# Required
DB_HOST=localhost
DB_PORT=5432
DB_USER=opensms
DB_PASSWORD=opensms_dev_password
DB_NAME=opensms
JWT_ACCESS_SECRET=your-secret-key-min-32-characters-long
JWT_REFRESH_SECRET=your-refresh-secret-min-32-chars
ENCRYPTION_KEY=12345678901234567890123456789012

# Optional (defaults provided)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=opensms_redis_password
NATS_URL=nats://localhost:4222
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=opensms
MINIO_SECRET_KEY=opensms_minio_password
```

**Frontend**: Create `.env` in `frontend/admin-portal/` directory:
```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### Infrastructure Services (Docker Compose)

**Services Defined**: `docker-compose.yml` at repo root

**Start all services**:
```bash
docker-compose up -d
```

**Individual services**:
```bash
# Start only database
docker-compose up -d postgres

# Start only cache
docker-compose up -d redis
```

**View logs**:
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f postgres
```

**Stop services**:
```bash
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

**Service URLs**:
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- NATS: `localhost:4222` (client), `localhost:8222` (monitoring)
- MinIO: `localhost:9000` (API), `localhost:9001` (console)
- Jaeger UI: `http://localhost:16686`

## How to Run Unit Tests

### Backend Tests

**Run all tests**:
```bash
cd backend

# Via Makefile (recommended)
make test

# Via go test directly
go test -v -race -coverprofile=coverage.out ./...
```

**Run specific package**:
```bash
# Security package tests
go test -v ./internal/shared/security/...

# Cache tests
go test -v ./internal/infrastructure/cache/...

# Middleware tests
go test -v ./internal/server/...
```

**Run with coverage report**:
```bash
make test
# Opens coverage.html in browser
open coverage.html
```

**Run specific test**:
```bash
# Run single test function
go test -v -run TestJWTManager_GenerateTokenPair ./internal/shared/security/

# Run tests matching pattern
go test -v -run TestJWT ./internal/shared/security/
```

**Run with race detection**:
```bash
go test -race ./...
```

**Run benchmarks**:
```bash
# All benchmarks
go test -bench=. ./internal/shared/security/...

# Specific benchmark
go test -bench=BenchmarkHashPassword ./internal/shared/security/
```

**Coverage by Package**:
- Security: >95% (crypto, JWT, RBAC)
- Server: >85% (errors, middleware)
- Infrastructure: >85% (database, cache, config)

**Test Files**:
- `backend/internal/shared/security/*_test.go` - Crypto, JWT, RBAC tests (282+267+302 = 851 lines)
- `backend/internal/server/*_test.go` - Error handling, middleware tests (139+400 = 539 lines)
- `backend/internal/infrastructure/database/postgres_test.go` - Database tests (340 lines)
- `backend/internal/infrastructure/cache/redis_test.go` - Cache tests (430 lines)
- `backend/internal/infrastructure/config/config_test.go` - Config tests (370 lines)

### Frontend Tests

**Run all tests**:
```bash
cd frontend/admin-portal

# Run tests
npm run test

# Run with UI
npm run test:ui

# Run with coverage
npm run test:coverage
```

**Watch mode**:
```bash
npm run test -- --watch
```

**Run specific test file**:
```bash
npm run test auth.spec.ts
```

**Coverage report**:
```bash
npm run test:coverage
# Opens coverage/index.html in browser
open coverage/index.html
```

**Test Files**:
- `frontend/admin-portal/src/stores/auth.spec.ts` - Auth store tests (312 lines)
- `frontend/admin-portal/src/stores/ui.spec.ts` - UI store tests (310 lines)
- `frontend/admin-portal/src/stores/students.spec.ts` - Students store tests (380 lines)
- `frontend/admin-portal/src/services/api.spec.ts` - API client tests (81 lines)
- `frontend/admin-portal/src/views/auth/LoginView.spec.ts` - Login component tests (270 lines)
- `frontend/admin-portal/src/views/DashboardView.spec.ts` - Dashboard tests (190 lines)
- `frontend/admin-portal/src/router/index.spec.ts` - Router tests (370 lines)
- `frontend/admin-portal/src/components/*.spec.ts` - Component tests (390 lines)

## How to Run Integration Tests

### Backend Integration Tests

**Prerequisites**:
- Docker services running (postgres, redis, nats)
- Test database created

**Setup test database**:
```bash
# Start services
docker-compose up -d postgres redis

# Create test database (if not exists)
docker exec opensms-postgres psql -U opensms -c "CREATE DATABASE opensms_test;"

# Run migrations on test database
goose -dir backend/migrations postgres \
  "host=localhost port=5432 user=opensms password=opensms_dev_password dbname=opensms_test sslmode=disable" up
```

**Run integration tests**:
```bash
cd backend

# Set test environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=opensms
export DB_PASSWORD=opensms_dev_password
export DB_NAME=opensms_test
export REDIS_HOST=localhost
export REDIS_PORT=6379
export JWT_ACCESS_SECRET=test-access-secret-key-min-32-characters
export JWT_REFRESH_SECRET=test-refresh-secret-key-min-32-chars
export ENCRYPTION_KEY=12345678901234567890123456789012

# Run integration tests
go test -v ./tests/integration/...
```

**Integration Test Files**:
- `backend/tests/integration/api_integration_test.go` - Full API integration tests (680 lines)
  - Authentication flow
  - RBAC enforcement
  - Database transactions
  - Rate limiting
  - Token refresh
  - Multi-tenancy isolation

### Frontend E2E Tests

**Prerequisites**:
- Backend API running (`make dev` in backend/)
- Playwright installed

**Install Playwright**:
```bash
cd frontend/admin-portal

# Install Playwright browsers (first time only)
npx playwright install
```

**Run E2E tests**:
```bash
cd frontend/admin-portal

# Run all E2E tests
npm run test:e2e

# Run with UI mode
npm run test:e2e:ui

# Run in debug mode
npm run test:e2e:debug

# Run specific browser
npx playwright test --project=chromium
```

**E2E Test Files**:
- `frontend/admin-portal/e2e/auth.spec.ts` - Authentication flow tests (200 lines)
- `frontend/admin-portal/e2e/students.spec.ts` - Student management tests (300 lines)

**Browsers Tested**: Chrome, Firefox, Safari, Mobile (Pixel 5, iPhone 12)

## Lint/Format Commands

### Backend (Go)

**Format code**:
```bash
cd backend

# Format all Go files
make fmt

# Or use go fmt directly
go fmt ./...

# With goimports (if installed)
goimports -w .
```

**Lint**:
```bash
cd backend

# Run golangci-lint
make lint

# Or directly
golangci-lint run --timeout=5m
```

**Tidy dependencies**:
```bash
make tidy
# Or: go mod tidy
```

### Frontend (TypeScript/Vue)

**Lint**:
```bash
cd frontend/admin-portal

# Run ESLint
npm run lint

# Auto-fix issues
npm run lint -- --fix
```

**Type check**:
```bash
npm run type-check
```

**Format** (if prettier configured):
```bash
npm run format
```

## CI Entrypoint

### GitHub Actions

**Configuration**: `.github/workflows/test.yml` (located in `backend/.github/workflows/`)

**Triggers**:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

**Jobs**:

1. **Backend Tests** (`test` job):
   - Setup: Ubuntu, Go 1.22, PostgreSQL 16, Redis 7
   - Steps:
     - Checkout code
     - Cache Go modules
     - Download dependencies: `go mod download`
     - Run tests: `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...`
     - Generate coverage: `go tool cover -html=coverage.out`
     - Upload to Codecov
     - Run linter: `golangci-lint run --timeout=5m`

2. **Frontend Tests** (`frontend-test` job):
   - Setup: Ubuntu, Node.js 20
   - Steps:
     - Checkout code
     - Cache npm modules
     - Install dependencies: `npm ci`
     - Type check: `npm run type-check`
     - Run tests: `npm run test:coverage`
     - Lint: `npm run lint`

**Environment Variables** (set in CI):
```yaml
DB_HOST: localhost
DB_PORT: 5432
DB_USER: opensms_test
DB_PASSWORD: test_password
DB_NAME: opensms_test
REDIS_HOST: localhost
REDIS_PORT: 6379
JWT_ACCESS_SECRET: test-access-secret-key-min-32-characters
JWT_REFRESH_SECRET: test-refresh-secret-key-min-32-chars
ENCRYPTION_KEY: 12345678901234567890123456789012
```

**View CI Results**:
- GitHub Actions tab in repository
- Check runs on pull requests
- Codecov reports for coverage

### Manual CI Simulation

**Run tests locally as CI would**:

Backend:
```bash
cd backend
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
golangci-lint run --timeout=5m
```

Frontend:
```bash
cd frontend/admin-portal
npm ci
npm run type-check
npm run test:coverage
npm run lint
```

## Quick Reference

### Common Commands

**Development**:
```bash
# Start everything
docker-compose up -d && cd backend && make dev

# Run tests while developing
cd backend && go test -v ./... --watch  # (with gotestsum)
cd frontend/admin-portal && npm run test -- --watch
```

**Database**:
```bash
# Migrate up
make migrate-up

# Migrate down
make migrate-down

# Check migration status
make migrate-status

# Create new migration
make migrate-create NAME=add_new_table
```

**Debugging**:
```bash
# Backend logs (if running with make dev)
# Logs output to stdout

# Docker service logs
docker-compose logs -f postgres
docker-compose logs -f redis

# Check service health
curl http://localhost:8080/health
```

**Clean up**:
```bash
# Clean backend build artifacts
cd backend && make clean

# Clean frontend build
cd frontend/admin-portal && rm -rf dist/ node_modules/

# Reset Docker services
docker-compose down -v
```
