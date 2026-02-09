# Components

**⚠️ IMPLEMENTATION STATUS**:
- ✅ Infrastructure components (1-5, 9-10): **FULLY IMPLEMENTED** with tests
- ❌ Business modules (6-8): **PLACEHOLDER ONLY** - handlers exist but return "to be implemented"
- See `docs/ai/implementation-status.md` for complete breakdown

## 1. HTTP Server & Routing

### Purpose
- Initialize Fiber web server with security middleware
- Register all module routes
- Handle graceful shutdown

### Key Types/Functions
- `main()` in `backend/cmd/api/main.go:30-193` - Server initialization and startup
- `Server` struct in `backend/internal/server/server.go:29-33`
- `NewServer()` in `backend/internal/server/server.go:36-54`
- `RegisterRoutes()` in `backend/internal/server/server.go:57-79`

### Important Flows
```
Startup:
main() → Load Config → Init Infrastructure (DB, Cache, NATS, MinIO) →
NewServer() → RegisterRoutes() → app.Listen() → Graceful Shutdown Handler

Request:
HTTP Request → Middleware Chain → Router → Handler → Response
  ├─ Security Headers (helmet)
  ├─ CORS
  ├─ Request ID
  ├─ Logger
  ├─ Recovery
  ├─ Compression
  ├─ Rate Limiting
  ├─ Auth Middleware (authMiddleware)
  └─ Permission Middleware (requirePermission)
```

### Touches
- **DB**: Via dependency injection to modules
- **Cache**: Redis client for sessions/rate limiting
- **External**: NATS event bus, MinIO storage

## 2. Authentication & Authorization (Security)

### Purpose
- JWT token generation and validation
- RBAC permission checking
- Password hashing and encryption

### Key Types/Functions
- `JWTManager` in `backend/internal/shared/security/jwt.go:21-38`
- `GenerateTokenPair()` in `backend/internal/shared/security/jwt.go:40-98`
- `ValidateAccessToken()` in `backend/internal/shared/security/jwt.go:100-165`
- `RefreshAccessToken()` in `backend/internal/shared/security/jwt.go:167-241`
- `RBAC` in `backend/internal/shared/security/rbac.go:23-49`
- `CheckPermission()` in `backend/internal/shared/security/rbac.go:88-125`
- `HashPassword()` in `backend/internal/shared/security/crypto.go:18-28`
- `EncryptData()` / `DecryptData()` in `backend/internal/shared/security/crypto.go:30-74`

### Important Flows
```
Login Flow:
POST /api/v1/auth/login → authHandler.Login() →
  ValidateCredentials() → HashPassword.Compare() →
  GenerateTokenPair() → Return {access_token, refresh_token}

Protected Request:
Request → authMiddleware() →
  Extract JWT → ValidateAccessToken() → Set UserContext →
  requirePermission() → CheckPermission(role, permission) → Handler

Token Refresh:
POST /api/v1/auth/refresh → authHandler.RefreshToken() →
  ValidateRefreshToken() → RefreshAccessToken() → New access_token
```

### Touches
- **DB**: User credential validation
- **Cache**: Token blacklist, session storage

## 3. Database Layer (PostgreSQL)

### Purpose
- PostgreSQL connection management with connection pooling
- Multi-tenancy via row-level security
- Transaction management

### Key Types/Functions
- `DB` struct in `backend/internal/infrastructure/database/postgres.go:19-23`
- `NewPostgresDB()` in `backend/internal/infrastructure/database/postgres.go:25-119`
- `SetTenant()` in `backend/internal/infrastructure/database/postgres.go:121-141`
- `Health()` in `backend/internal/infrastructure/database/postgres.go:143-150`

### Important Flows
```
Connection:
NewPostgresDB() → pgx.ParseConfig() → pgxpool.NewWithConfig() →
  Connection Pooling (max 25 conns) → Health Check

Multi-Tenant Query:
Request → authMiddleware (extract tenant_id) →
  SetTenant(ctx, tenant_id) → SET app.current_tenant →
  Query → Row-Level Security Applied → Results filtered by tenant

Transaction:
BeginTx() → SetTenant() → Execute Queries → Commit/Rollback
```

### Touches
- **DB**: Direct PostgreSQL connection via pgx driver
- **RLS**: Uses `SET app.current_tenant` for tenant isolation

## 4. Caching Layer (Redis)

### Purpose
- Session storage
- API response caching
- Rate limiting with sliding window
- Token blacklist

### Key Types/Functions
- `RedisClient` struct in `backend/internal/infrastructure/cache/redis.go:17-23`
- `NewRedisClient()` in `backend/internal/infrastructure/cache/redis.go:25-65`
- `Get()` / `Set()` in `backend/internal/infrastructure/cache/redis.go:67-108`
- `RateLimiter` in `backend/internal/infrastructure/cache/redis.go:120-140`
- `Allow()` in `backend/internal/infrastructure/cache/redis.go:142-212`

### Important Flows
```
Cache Pattern:
Get(key) → Cache Hit? → Return cached data
           ↓ Cache Miss
     Query Database → Set(key, data, TTL) → Return data

Rate Limiting:
Request → Rate Limiter → Allow(client_id, limit, window) →
  Redis ZADD (sliding window) → ZCOUNT →
  Count < Limit? → Allow : Deny
```

### Touches
- **Cache**: Direct Redis connection
- **External**: Used by all modules for caching

## 5. Event Bus (NATS)

### Purpose
- Asynchronous event-driven communication between modules
- Decoupling of services
- Event sourcing capability

### Key Types/Functions
- `NATSConnection` struct in `backend/internal/infrastructure/eventbus/nats.go:15-21`
- `NewNATSConnection()` in `backend/internal/infrastructure/eventbus/nats.go:23-62`
- `Publish()` in `backend/internal/infrastructure/eventbus/nats.go:64-75`
- `Subscribe()` in `backend/internal/infrastructure/eventbus/nats.go:77-99`

### Important Flows
```
Event Publishing:
Module Action → EventBus.Publish(subject, data) →
  NATS JetStream → Subscribers notified

Event Consumption:
Subscribe(subject, handler) → NATS receives event →
  handler(msg) → Process event → ACK
```

### Touches
- **External**: NATS JetStream server
- **Queue**: Message persistence and delivery guarantees

## 6. Authentication Module ❌ PLACEHOLDER

### Purpose
- User registration, login, logout
- Password reset and email verification
- 2FA enablement

### Key Types/Functions
- `Handler` struct in `backend/internal/modules/auth/handler.go`
- **ALL METHODS RETURN "to be implemented"**: `Login()`, `Register()`, `RefreshToken()`, `ForgotPassword()`, etc.

### Important Flows
```
⚠️ PLANNED (NOT IMPLEMENTED):
POST /auth/register → Returns 501 Not Implemented
POST /auth/login → Returns 501 Not Implemented
POST /auth/refresh → Returns 501 Not Implemented
```

### Touches
- **DB**: Would use users table (NOT IMPLEMENTED)
- **Cache**: Would use session storage (NOT IMPLEMENTED)
- **Event Bus**: Would publish events (NOT IMPLEMENTED)

## 7. Users Module ❌ PLACEHOLDER

### Purpose
- User/student/teacher CRUD operations
- User search and filtering
- Role assignment

### Key Types/Functions
- `Handler` struct in `backend/internal/modules/users/handler.go`
- **ALL METHODS RETURN "to be implemented"**: `ListUsers()`, `GetUser()`, `CreateUser()`, `UpdateUser()`, `DeleteUser()`, `ListStudents()`, `CreateStudent()`, etc.

### Important Flows
```
⚠️ PLANNED (NOT IMPLEMENTED):
POST /students → Returns {"message": "Create student - to be implemented"}
GET /students → Returns {"message": "List students - to be implemented"}
GET /students/:id → Returns {"message": "Get student - to be implemented"}
```

### Touches
- **DB**: Would use users, students, teachers tables (NOT IMPLEMENTED)
- **Cache**: Would cache data (NOT IMPLEMENTED)
- **Event Bus**: Would publish events (NOT IMPLEMENTED)

## 8. Academic Module ❌ PLACEHOLDER

### Purpose
- Academic year, class, subject management
- Grade and attendance tracking
- Report card generation

### Key Types/Functions
- `Handler` struct in `backend/internal/modules/academic/handler.go`
- **ALL METHODS RETURN "to be implemented"**: `ListClasses()`, `CreateGrade()`, `MarkAttendance()`, etc.

### Important Flows
```
⚠️ PLANNED (NOT IMPLEMENTED):
POST /grades → Returns {"message": "Create grade - to be implemented"}
POST /attendance → Returns {"message": "Mark attendance - to be implemented"}
GET /classes → Returns {"message": "List classes - to be implemented"}
```

### Touches
- **DB**: Would use academic_years, classes, subjects tables (grades/attendance tables don't exist yet)
- **Cache**: Would cache data (NOT IMPLEMENTED)
- **Event Bus**: Would publish events (NOT IMPLEMENTED)

## 9. Configuration Management

### Purpose
- Load configuration from environment and files
- Validate required settings
- Provide defaults

### Key Types/Functions
- `Config` struct in `backend/internal/infrastructure/config/config.go:11-20`
- `Load()` in `backend/internal/infrastructure/config/config.go:104-210`
- `validate()` in `backend/internal/infrastructure/config/config.go:269-294`
- `setDefaults()` in `backend/internal/infrastructure/config/config.go:212-267`

### Important Flows
```
Configuration Loading:
Load() → Viper.ReadInConfig() → Parse Env Vars →
  setDefaults() → Build Config Struct → validate() →
  Return Config or Error
```

### Touches
- **Files**: YAML config files (optional)
- **Environment**: System environment variables

## 10. Observability (Logging, Metrics, Tracing)

### Purpose
- Structured logging with Zap
- Prometheus metrics collection
- Distributed tracing with Jaeger

### Key Types/Functions
- `NewLogger()` in `backend/internal/infrastructure/observability/logger.go:14-56`
- `InitMetrics()` in `backend/internal/infrastructure/observability/metrics.go:14-72`
- `InitTracing()` in `backend/internal/infrastructure/observability/tracing.go:18-61`
- `MetricsHandler` - Prometheus HTTP handler

### Important Flows
```
Request Tracing:
Request → Create Span → Execute Handler →
  Log Events → Record Metrics → End Span → Export to Jaeger

Metrics:
Request → Increment Counter → Record Latency →
  Update Histograms → Prometheus Scrapes /metrics
```

### Touches
- **External**: Jaeger collector, Prometheus
- **Logging**: Stdout/stderr in JSON format
