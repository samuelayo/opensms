# Data and Contracts

## Datastores

### PostgreSQL (Primary Database)

**Connection**: `backend/internal/infrastructure/database/postgres.go:25-119`

#### Core Tables (001_initial_schema.sql)

**Multi-Tenancy**:
- `tenants` - Top-level tenant isolation (schools/institutions)
  - Columns: id, name, slug, domain, academic_level, timezone, country, settings (JSONB), subscription_plan, max_students, max_teachers, is_active
  - Indexes: slug, domain, is_active

- `schools` - Multi-campus support within tenant
  - Columns: id, tenant_id, name, code, address, phone, email, principal_id, is_active
  - RLS: `tenant_id = current_setting('app.current_tenant')::UUID`
  - Indexes: tenant_id, code

**Users & Authentication**:
- `users` - Core authentication table
  - Columns: id, tenant_id, email, password_hash, role (ENUM), status (ENUM), first_name, last_name, phone, avatar_url, last_login_at, failed_login_attempts, email_verified, two_factor_enabled, preferences (JSONB)
  - RLS: `tenant_id = current_setting('app.current_tenant')::UUID`
  - Indexes: tenant_id, email, role, status
  - Unique: (tenant_id, email)

- `refresh_tokens` - JWT refresh token storage
  - Columns: id, user_id, token_hash, expires_at, revoked_at, replaced_by_token
  - Indexes: user_id, token_hash, expires_at

**Academic Structure**:
- `academic_years` - Academic year/session tracking
  - Columns: id, tenant_id, school_id, name, start_date, end_date, is_current
  - RLS: tenant_id isolation
  - Indexes: tenant_id, school_id, is_current

- `terms` - Terms/semesters within academic years
  - Columns: id, tenant_id, academic_year_id, name, term_number, start_date, end_date, is_current
  - RLS: tenant_id isolation

- `grade_levels` - Flexible grade levels for all education types
  - Columns: id, tenant_id, school_id, name, level_order, academic_level (ENUM: primary/secondary/tertiary)
  - RLS: tenant_id isolation

- `departments` - Academic departments/faculties
  - Columns: id, tenant_id, school_id, name, code, head_id, description
  - RLS: tenant_id isolation

- `classes` - Classes/sections
  - Columns: id, tenant_id, school_id, grade_level_id, department_id, name, section, class_teacher_id, capacity, room_number, academic_year_id
  - RLS: tenant_id isolation
  - Unique: (tenant_id, school_id, grade_level_id, section, academic_year_id)

- `subjects` - Subjects/courses
  - Columns: id, tenant_id, school_id, name, code, description, department_id, credit_hours, is_elective
  - RLS: tenant_id isolation
  - Unique: (tenant_id, school_id, code)

#### Student & Teacher Tables (002_students_and_enrollment.sql)

**Students**:
- `students` - Student records
  - Columns: id, tenant_id, user_id, school_id, admission_number, admission_date, date_of_birth, gender (ENUM), blood_group, nationality, address, emergency_contact_*, medical_conditions, current_grade_level_id, current_class_id, is_active
  - RLS: tenant_id isolation
  - Unique: (tenant_id, admission_number)

- `student_guardians` - Student-parent relationships (many-to-many)
  - Columns: id, student_id, guardian_user_id, relationship, is_primary, can_pickup, can_receive_communication

- `student_documents` - Student document storage
  - Columns: id, student_id, document_type, document_name, file_path, file_size, mime_type, uploaded_by

**Teachers**:
- `teachers` - Teacher records
  - Columns: id, tenant_id, user_id, school_id, employee_id, department_id, date_of_birth, gender, date_of_joining, qualification, specialization, experience_years, employment_type
  - RLS: tenant_id isolation

**Expected Additional Tables** (not yet in migrations):
- `grades` - Student grades/marks
- `attendance` - Student attendance records
- `assignments` - Homework/assignments
- `exams` - Examination records
- `fee_structures` - Fee management
- `payments` - Payment tracking

#### Custom Types (ENUMs)

- `user_role`: super_admin, school_admin, principal, teacher, student, parent, librarian, accountant, registrar, counselor, nurse, transport_admin, hostel_warden
- `user_status`: active, inactive, suspended, deleted
- `gender`: male, female, other, prefer_not_to_say
- `academic_level`: primary, secondary, tertiary
- `payment_status`: pending, partial, paid, overdue, cancelled
- `attendance_status`: present, absent, late, excused, sick_leave

#### Migrations

**Location**: `backend/migrations/`
**Tool**: Goose (`github.com/pressly/goose/v3`)
**Commands**:
- Up: `make migrate-up` or `goose -dir migrations postgres "connection_string" up`
- Down: `make migrate-down`
- Create: `make migrate-create NAME=migration_name`

**Existing Migrations**:
1. `001_initial_schema.sql` - Tenants, users, academic structure
2. `002_students_and_enrollment.sql` - Students, teachers, guardians

### Redis (Cache & Sessions)

**Connection**: `backend/internal/infrastructure/cache/redis.go:25-65`

**Key Patterns**:
- `session:{user_id}` - User session data (TTL: 30 minutes)
- `user:{id}` - Cached user data (TTL: 10 minutes)
- `student:{id}` - Cached student data (TTL: 10 minutes)
- `rate_limit:{client_id}:{endpoint}` - Rate limiting (TTL: 1 minute, sliding window)
- `token_blacklist:{token_hash}` - Revoked tokens (TTL: token expiry time)
- `cache:{resource}:{id}` - Generic cache pattern (TTL: varies)

**Rate Limiter**: `backend/internal/infrastructure/cache/redis.go:142-212` - Sliding window implementation using ZADD/ZCOUNT

### MinIO/S3 (Object Storage)

**Connection**: `backend/internal/infrastructure/storage/minio.go:24-78`

**Buckets**:
- `opensms` (default) - Student documents, avatars, attachments

**Object Paths**:
- `documents/{student_id}/{document_type}/{filename}` - Student documents
- `avatars/{user_id}/{filename}` - User profile pictures
- `attachments/{entity_type}/{entity_id}/{filename}` - General attachments

## Message Contracts (Event Bus)

### NATS Topics

**Connection**: `backend/internal/infrastructure/eventbus/nats.go:23-62`

**Event Patterns** (from ARCHITECTURE.md, not yet fully implemented):
- `student.enrolled` - Published when student is created/enrolled
- `student.updated` - Published when student details are updated
- `student.withdrawn` - Published when student leaves school
- `user.registered` - Published when new user is registered
- `user.login` - Published on successful login
- `grade.created` - Published when grade is entered
- `attendance.marked` - Published when attendance is recorded
- `payment.received` - Published when fee payment is made

**Event Schema** (typical structure):
```json
{
  "event_id": "uuid",
  "event_type": "student.enrolled",
  "tenant_id": "uuid",
  "timestamp": "ISO-8601",
  "data": {
    "student_id": "uuid",
    "school_id": "uuid",
    ...
  }
}
```

**Schemas Location**: Not yet formalized, defined per module

## API Contracts

### REST API

**Base Path**: `/api/v1`
**Format**: JSON
**Authentication**: Bearer JWT tokens in `Authorization` header

#### Common Request Headers
- `Authorization: Bearer {access_token}` - Required for protected routes
- `X-Tenant-ID: {tenant_uuid}` - Tenant context (extracted from JWT)
- `Content-Type: application/json`
- `X-Request-ID: {uuid}` - Auto-generated request ID

#### Common Response Format

**Success** (2xx):
```json
{
  "data": { ... },
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

**Error** (4xx/5xx):
```json
{
  "code": "ERROR_CODE",
  "message": "Human readable message",
  "details": "Detailed error information"
}
```

**Error Handler**: `backend/internal/server/errors.go:14-95`

#### Pagination Query Params
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 20, max: 100)
- `sort` - Sort field (default: created_at)
- `order` - Sort order (asc/desc, default: desc)

#### API Documentation
- **Format**: Not yet generated
- **Future**: OpenAPI 3.0/Swagger spec generation planned

### API Modules

See `docs/ai/entrypoints.md` for complete route listing by module:
- Auth Module: `/api/v1/auth/*`
- Users Module: `/api/v1/users/*`, `/api/v1/students/*`, `/api/v1/teachers/*`
- Academic Module: `/api/v1/academic-years/*`, `/api/v1/classes/*`, `/api/v1/subjects/*`, `/api/v1/grades/*`, `/api/v1/attendance/*`

## Configuration Keys

### Environment Variables

**Location**: Loaded via `backend/internal/infrastructure/config/config.go:104-210`
**Defaults**: `backend/internal/infrastructure/config/config.go:212-267`

#### Server Configuration
- `SERVER_HOST` - Server bind address (default: "0.0.0.0")
- `SERVER_PORT` - Server port (default: 8080)
- `ALLOWED_ORIGINS` - CORS allowed origins (default: "*")
- `TRUSTED_PROXIES` - Trusted proxy IPs (default: [])
- `SERVER_READ_TIMEOUT` - Read timeout (default: 30s)
- `SERVER_WRITE_TIMEOUT` - Write timeout (default: 30s)

#### Database Configuration
- `DB_HOST` - PostgreSQL host (default: "localhost") **[REQUIRED]**
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database user (default: "") **[REQUIRED]**
- `DB_PASSWORD` - Database password (default: "")
- `DB_NAME` - Database name (default: "") **[REQUIRED]**
- `DB_SSLMODE` - SSL mode (default: "disable")
- `DB_MAX_OPEN_CONNS` - Max open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: 5)
- `DB_CONN_MAX_LIFETIME` - Connection max lifetime (default: 5m)
- `DB_CONN_MAX_IDLE_TIME` - Connection max idle time (default: 10m)

#### Redis Configuration
- `REDIS_HOST` - Redis host (default: "localhost")
- `REDIS_PORT` - Redis port (default: 6379)
- `REDIS_PASSWORD` - Redis password (default: "")
- `REDIS_DB` - Redis database number (default: 0)
- `REDIS_POOL_SIZE` - Connection pool size (default: 10)

#### NATS Configuration
- `NATS_URL` - NATS server URL (default: "nats://localhost:4222")
- `NATS_MAX_RECONNECTS` - Max reconnect attempts (default: 10)
- `NATS_RECONNECT_WAIT` - Reconnect wait time (default: 2s)

#### MinIO/S3 Configuration
- `MINIO_ENDPOINT` - MinIO endpoint (default: "localhost:9000")
- `MINIO_ACCESS_KEY` - Access key (default: "")
- `MINIO_SECRET_KEY` - Secret key (default: "")
- `MINIO_USE_SSL` - Use SSL (default: false)
- `MINIO_BUCKET` - Bucket name (default: "opensms")
- `MINIO_REGION` - Region (default: "us-east-1")

#### JWT Configuration
- `JWT_ACCESS_SECRET` - Access token secret (default: "") **[REQUIRED, min 32 chars]**
- `JWT_REFRESH_SECRET` - Refresh token secret (default: "") **[REQUIRED, min 32 chars]**
- `JWT_ACCESS_EXPIRY` - Access token expiry (default: 15m)
- `JWT_REFRESH_EXPIRY` - Refresh token expiry (default: 7d)
- `JWT_ISSUER` - Token issuer (default: "opensms")

#### Security Configuration
- `ENCRYPTION_KEY` - AES-256 encryption key (default: "") **[REQUIRED, exactly 32 bytes]**
- `RATE_LIMIT_PER_MINUTE` - Rate limit per minute (default: 100)
- `MAX_LOGIN_ATTEMPTS` - Max failed login attempts (default: 5)
- `LOGIN_ATTEMPT_WINDOW` - Login attempt window (default: 15m)
- `PASSWORD_MIN_LENGTH` - Minimum password length (default: 12)
- `REQUIRE_STRONG_PASSWORDS` - Require strong passwords (default: true)
- `SESSION_TIMEOUT` - Session timeout (default: 30m)
- `CSRF_ENABLED` - Enable CSRF protection (default: true)

#### Observability Configuration
- `JAEGER_ENDPOINT` - Jaeger collector endpoint (default: "http://localhost:14268/api/traces")
- `PROMETHEUS_PORT` - Prometheus port (default: 9090)
- `LOG_LEVEL` - Log level (default: "info")
- `ENABLE_TRACING` - Enable distributed tracing (default: true)
- `ENABLE_METRICS` - Enable metrics collection (default: true)
- `TRACING_SAMPLING_RATE` - Trace sampling rate (default: 0.1)

### Configuration Loading Priority

1. Environment variables (highest priority)
2. YAML config file (`./config/config.yaml` or `.`)
3. Default values (lowest priority)

### Validation

**Validation**: `backend/internal/infrastructure/config/config.go:269-294`

**Required Fields**:
- `DB_HOST`, `DB_USER`, `DB_NAME` - Database connection
- `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET` - JWT signing
- `ENCRYPTION_KEY` - Must be exactly 32 bytes for AES-256
