# Invariants and Risk

## Security Invariants

### 1. Multi-Tenancy Isolation (CRITICAL)

**Invariant**: All database queries MUST enforce tenant isolation via row-level security
**Code Pointers**:
- `backend/internal/infrastructure/database/postgres.go:121-141` - `SetTenant()` sets `app.current_tenant`
- `backend/internal/server/middleware.go:14-105` - Auth middleware extracts tenant_id from JWT
- `backend/migrations/001_initial_schema.sql:72-75` - RLS policies on all tenant tables

**How to Test**:
1. Create two tenants with different data
2. Authenticate as user from tenant A
3. Attempt to access tenant B's data via API
4. Verify: Response returns 404/empty, not tenant B's data

**Risk if Broken**: **CRITICAL** - Cross-tenant data leakage, GDPR/compliance violation

### 2. Authentication Required (HIGH)

**Invariant**: All routes except `/api/v1/public/*`, `/health`, `/metrics`, `/auth/login`, `/auth/register` MUST require valid JWT
**Code Pointers**:
- `backend/internal/server/middleware.go:14-105` - `authMiddleware()` validates JWT on all protected routes
- `backend/internal/server/server.go:63-66` - Public routes registration
- `backend/internal/server/server.go:103-179` - Protected routes use `s.authMiddleware()`

**How to Test**:
1. Call protected endpoint without Authorization header
2. Call with invalid/expired token
3. Verify: Returns 401 Unauthorized

**Risk if Broken**: **CRITICAL** - Unauthorized access to sensitive student/academic data

### 3. RBAC Permission Enforcement (HIGH)

**Invariant**: Users can only perform actions allowed by their role's permissions
**Code Pointers**:
- `backend/internal/shared/security/rbac.go:51-200` - Role permission definitions
- `backend/internal/server/middleware.go:107-149` - `requirePermission()` middleware
- `backend/internal/server/server.go:121-179` - Routes protected with permission checks

**How to Test**:
1. Create user with 'student' role
2. Attempt to create another user (requires `user:create`)
3. Verify: Returns 403 Forbidden

**Risk if Broken**: **HIGH** - Privilege escalation, unauthorized data modification

### 4. Password Security (HIGH)

**Invariant**: Passwords MUST be hashed with bcrypt (cost 12) before storage, never stored in plaintext
**Code Pointers**:
- `backend/internal/shared/security/crypto.go:18-28` - `HashPassword()` uses bcrypt cost 12
- `backend/internal/shared/security/crypto.go:76-91` - `ValidatePassword()` compares hashes

**How to Test**:
1. Register new user
2. Query database directly
3. Verify: password_hash column starts with "$2a$12$" (bcrypt with cost 12)

**Risk if Broken**: **CRITICAL** - Mass credential compromise if database is breached

### 5. Token Expiry and Rotation (MEDIUM)

**Invariant**: Access tokens expire in 15 minutes, refresh tokens expire in 7 days, refresh tokens must be rotated on use
**Code Pointers**:
- `backend/internal/shared/security/jwt.go:40-98` - Token generation with expiry
- `backend/internal/shared/security/jwt.go:167-241` - Refresh token rotation
- `backend/internal/infrastructure/config/config.go:247-248` - Default expiry times

**How to Test**:
1. Login to get tokens
2. Wait 16 minutes
3. Call protected endpoint with access token
4. Verify: Returns 401, token expired

**Risk if Broken**: **MEDIUM** - Increased window for token theft/replay attacks

### 6. Data Encryption at Rest (MEDIUM)

**Invariant**: Sensitive fields (SSN, medical records, etc.) MUST be encrypted with AES-256
**Code Pointers**:
- `backend/internal/shared/security/crypto.go:30-74` - `EncryptData()` / `DecryptData()` using AES-256
- `backend/internal/infrastructure/config/config.go:183-191` - Encryption key validation (must be 32 bytes)

**How to Test**:
1. Create student with medical_conditions
2. Verify encrypted data is not readable in DB
3. Retrieve via API, verify decrypted correctly

**Risk if Broken**: **MEDIUM** - Sensitive data exposure in database backups/logs

## Correctness Invariants

### 7. Unique Constraints (HIGH)

**Invariant**: Email must be unique per tenant, admission numbers must be unique per tenant
**Code Pointers**:
- `backend/migrations/001_initial_schema.sql:108` - `UNIQUE(tenant_id, email)` on users table
- `backend/migrations/002_students_and_enrollment.sql:38` - `UNIQUE(tenant_id, admission_number)` on students table

**How to Test**:
1. Create user with email test@example.com in tenant A
2. Attempt to create another user with same email in tenant A
3. Verify: Returns 409 Conflict

**Risk if Broken**: **MEDIUM** - Data integrity issues, duplicate students/users

### 8. Referential Integrity (MEDIUM)

**Invariant**: Foreign keys must reference existing records, cascading deletes must clean up related data
**Code Pointers**:
- All migrations use `REFERENCES table(id) ON DELETE CASCADE` for proper cleanup
- Example: `backend/migrations/001_initial_schema.sql:52` - schools reference tenants with CASCADE

**How to Test**:
1. Create tenant with students
2. Delete tenant
3. Verify: Students and all related records are also deleted

**Risk if Broken**: **MEDIUM** - Orphaned records, database corruption

### 9. Transaction Atomicity (MEDIUM)

**Invariant**: Multi-step operations (e.g., enrollment with grade creation) must be atomic
**Code Pointers**:
- `backend/internal/infrastructure/database/postgres.go` - Uses pgx transactions
- Modules should use `BeginTx()` for multi-step operations

**How to Test**:
1. Enroll student with grade creation
2. Simulate error midway (e.g., invalid grade value)
3. Verify: Student enrollment is rolled back

**Risk if Broken**: **MEDIUM** - Inconsistent data state

## Consistency Invariants

### 10. Tenant Context Propagation (CRITICAL)

**Invariant**: Every database operation must have tenant context set via `SetTenant()`
**Code Pointers**:
- `backend/internal/server/middleware.go:14-105` - Auth middleware sets tenant_id in context
- `backend/internal/infrastructure/database/postgres.go:121-141` - `SetTenant()` must be called before queries

**How to Test**:
1. Review module handlers
2. Verify all DB queries preceded by `SetTenant(ctx, tenantID)`
3. Integration test: Query without SetTenant should return empty results

**Risk if Broken**: **CRITICAL** - Data leakage across tenants

## Failure Modes & Recovery

### 11. Database Connection Failures (HIGH)

**Handling**:
- Connection pooling with automatic reconnection (pgx driver)
- Health check endpoint monitors DB status: `/api/v1/public/health`
- Max 25 connections, 5 idle, 5min max lifetime

**Code Pointers**:
- `backend/internal/infrastructure/database/postgres.go:25-119` - Connection pool configuration
- `backend/internal/server/server.go:182-207` - Health check includes DB check

**Recovery**: Automatic via pgx connection pool, manual restart if needed

**Test**: Stop database, verify health check fails, restart database, verify recovery

### 12. Redis Cache Failures (MEDIUM)

**Handling**:
- Cache-aside pattern: DB is source of truth
- Redis failure should not prevent reads (fall back to DB)
- Session data loss requires re-login

**Code Pointers**:
- `backend/internal/infrastructure/cache/redis.go:25-65` - Redis client with error handling
- Modules should implement graceful degradation on cache miss

**Recovery**: Application continues without cache, performance degrades

**Test**: Stop Redis, verify API still responds (slower), sessions may fail

### 13. Rate Limiting (MEDIUM)

**Handling**:
- Global rate limit: 100 requests/minute per IP (Fiber middleware)
- Per-endpoint rate limits via Redis-based rate limiter
- Returns 429 Too Many Requests when exceeded

**Code Pointers**:
- `backend/cmd/api/main.go:136-144` - Global rate limiter (Fiber middleware)
- `backend/internal/infrastructure/cache/redis.go:142-212` - Sliding window rate limiter

**Test**: Send 101 requests in 1 minute, verify 101st returns 429

### 14. Token Refresh Idempotency (MEDIUM)

**Handling**:
- Refresh tokens are single-use (replaced_by_token tracking)
- Replay attacks prevented by checking if token already used
- 5-second grace period for concurrent requests

**Code Pointers**:
- `backend/internal/shared/security/jwt.go:167-241` - Refresh token rotation logic
- `backend/migrations/001_initial_schema.sql:123-131` - refresh_tokens table with replaced_by_token

**Test**: Use same refresh token twice, verify second attempt fails

### 15. Graceful Shutdown (MEDIUM)

**Handling**:
- 30-second timeout for in-flight requests
- Database connections closed cleanly
- Signal handling for SIGINT/SIGTERM

**Code Pointers**:
- `backend/cmd/api/main.go:181-193` - Graceful shutdown with timeout
- Cleanup deferred in main(): DB, cache, event bus connections

**Test**: Send SIGTERM, verify active requests complete, server exits cleanly

## "Do Not Break" List

### 1. Row-Level Security Policies (CRITICAL)
**Why**: Tenant isolation is the foundation of multi-tenancy security
**How to Test**: Integration tests in `backend/tests/integration/` verify RLS enforcement
**Files**: All migration files with `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`

### 2. JWT Signature Validation (CRITICAL)
**Why**: Prevents token forgery and impersonation
**How to Test**: `backend/internal/shared/security/jwt_test.go` - Token validation tests
**Files**: `backend/internal/shared/security/jwt.go:100-165`

### 3. RBAC Permission Definitions (HIGH)
**Why**: Changing permissions can grant/revoke access unexpectedly
**How to Test**: `backend/internal/shared/security/rbac_test.go` - All role tests
**Files**: `backend/internal/shared/security/rbac.go:51-200`

### 4. Database Migration Order (HIGH)
**Why**: Migrations must be applied in sequence for schema consistency
**How to Test**: Fresh database migration run in CI: `.github/workflows/test.yml`
**Files**: `backend/migrations/*.sql` numbered sequentially

### 5. Password Hash Cost Factor (HIGH)
**Why**: Lowering bcrypt cost reduces password security
**How to Test**: `backend/internal/shared/security/crypto_test.go` - Hash format validation
**Files**: `backend/internal/shared/security/crypto.go:18` - bcrypt.DefaultCost (12)

### 6. Configuration Validation (MEDIUM)
**Why**: Missing required config can cause runtime failures
**How to Test**: `backend/internal/infrastructure/config/config_test.go` - Validation tests
**Files**: `backend/internal/infrastructure/config/config.go:269-294`

### 7. CORS Configuration (MEDIUM)
**Why**: Incorrect CORS allows unauthorized origins to call API
**How to Test**: Manual testing from different origins
**Files**: `backend/cmd/api/main.go:111-117` - CORS middleware config

### 8. API Response Format (MEDIUM)
**Why**: Frontend depends on consistent response structure
**How to Test**: Integration tests verify response structure
**Files**: `backend/internal/server/errors.go:14-95` - Error response format

### 9. Middleware Chain Order (MEDIUM)
**Why**: Security middleware must run before business logic
**How to Test**: Review `backend/cmd/api/main.go:100-144` - middleware registration order
**Files**: `backend/cmd/api/main.go:100-144`

### 10. Event Schema Compatibility (LOW)
**Why**: Breaking changes to event payloads break consumers
**How to Test**: Currently no versioning, manual review required
**Files**: Event publishing in module handlers

## Areas of High Coupling / Blast Radius

### 1. Shared Security Package (HIGH BLAST RADIUS)
**Location**: `backend/internal/shared/security/`
**Used By**: All modules (auth, users, academic)
**Risk**: Changes to JWT/RBAC/Crypto affect entire application
**Mitigation**: Comprehensive test suite (`*_test.go`), avoid breaking changes

### 2. Database Infrastructure (HIGH BLAST RADIUS)
**Location**: `backend/internal/infrastructure/database/`
**Used By**: All modules
**Risk**: Connection pool changes, transaction handling bugs affect all data operations
**Mitigation**: Integration tests, health checks, connection pool monitoring

### 3. Middleware Chain (HIGH BLAST RADIUS)
**Location**: `backend/internal/server/middleware.go`
**Used By**: All API requests
**Risk**: Auth/permission middleware bugs block all requests
**Mitigation**: Unit tests for each middleware, integration tests for chains

### 4. Configuration Loading (MEDIUM BLAST RADIUS)
**Location**: `backend/internal/infrastructure/config/`
**Used By**: Server initialization, all infrastructure components
**Risk**: Config parsing errors prevent server startup
**Mitigation**: Validation tests, fail-fast on invalid config

### 5. User Table Schema (MEDIUM BLAST RADIUS)
**Location**: `backend/migrations/001_initial_schema.sql:82-121`
**Used By**: Auth, users, students, teachers modules
**Risk**: Schema changes require coordinated updates across modules
**Mitigation**: Database migrations, backward compatibility considerations

### 6. Tenant Context (HIGH BLAST RADIUS)
**Location**: Request context propagation throughout application
**Used By**: Every database query
**Risk**: Missing tenant context causes data leakage
**Mitigation**: Middleware enforcement, linting rules (future), integration tests

### 7. Redis Client (MEDIUM BLAST RADIUS)
**Location**: `backend/internal/infrastructure/cache/`
**Used By**: Sessions, rate limiting, caching across all modules
**Risk**: Redis connection issues degrade performance or block requests
**Mitigation**: Graceful degradation, health checks, circuit breaker pattern (future)

### 8. Router Registration (MEDIUM BLAST RADIUS)
**Location**: `backend/internal/server/server.go:57-79`
**Used By**: All API endpoints
**Risk**: Route conflicts, incorrect middleware application
**Mitigation**: Route listing in entrypoints.md, integration tests for all routes
