# OpenSMS Architecture Documentation

## Executive Summary

OpenSMS is a production-grade, hyper-scalable school management system designed for extreme scale, handling millions of users across thousands of institutions. Built with enterprise-grade practices and 20+ years of platform engineering experience.

## Architectural Decisions

### 1. Modular Monolith with Microservices Evolution Path

**Decision**: Start with a modular monolith, evolve to microservices as needed.

**Rationale**:
- Lower operational complexity initially
- Faster development and deployment
- Clear module boundaries allow easy extraction
- Shared transactions and consistency
- Lower latency for inter-module communication

**Evolution Path**: Extract high-traffic or independent modules (e.g., notifications, file processing) into microservices when:
- Module exceeds 10K requests/second
- Independent scaling requirements emerge
- Team size justifies operational overhead

### 2. Multi-Tenancy with Row-Level Security

**Implementation**: PostgreSQL Row-Level Security (RLS) + Application-level tenant filtering

**Benefits**:
- **Security**: Database-enforced data isolation
- **Cost**: Shared infrastructure reduces per-tenant cost
- **Compliance**: Easier GDPR/FERPA compliance with data isolation
- **Scale**: Supports 10K+ tenants on single database cluster

**Sharding Strategy** (for extreme scale):
```
Tenant Sharding (100K+ tenants):
- Hash-based sharding on tenant_id
- Shard 0: Tenants 0-999
- Shard 1: Tenants 1000-1999
- Router maintains shard mapping
```

### 3. Security-First Architecture

**Zero-Trust Principles**:
- No implicit trust between components
- All requests authenticated and authorized
- Encryption at rest and in transit
- Audit logging for compliance

**Defense in Depth**:
```
Layer 1: WAF (CloudFlare) - DDoS, Bot protection
Layer 2: Load Balancer - Rate limiting, SSL termination
Layer 3: API Gateway - Authentication, Authorization
Layer 4: Application - RBAC, Input validation
Layer 5: Database - RLS, Encrypted columns
Layer 6: Network - VPC isolation, Security groups
```

**Key Security Features**:
- AES-256 encryption at rest
- TLS 1.3 in transit
- Bcrypt password hashing (cost 12)
- JWT with refresh token rotation
- RBAC with fine-grained permissions
- CSRF protection
- XSS/SQL injection prevention
- Rate limiting per user/tenant/endpoint

### 4. Event-Driven Architecture

**Event Bus**: NATS JetStream for reliability

**Benefits**:
- Decoupling between modules
- Async processing of non-critical operations
- Event sourcing capability for audit trails
- Scalable message processing

**Event Flow Example**:
```
Student Enrollment:
1. API receives enrollment request
2. Validate and save to database
3. Publish "student.enrolled" event
4. Consumers:
   - Email service: Send welcome email
   - Finance service: Generate fee invoice
   - Notification service: Alert teachers
   - Analytics service: Update metrics
```

### 5. CQRS for Read-Heavy Operations

**Pattern**: Separate read and write models for complex queries

**Implementation**:
```
Write Model (Commands):
- Academic records
- Attendance
- Grades
Direct writes to primary DB

Read Model (Queries):
- Report cards
- Analytics dashboards
- Search functionality
Read from optimized read replicas or materialized views
```

**Benefits**:
- Optimized read performance
- Scalable read replicas
- Complex aggregations pre-computed
- Reduced load on write path

### 6. Caching Strategy

**Multi-Layer Cache**:

```
L1: In-Memory (Go map/sync.Map)
- TTL: 1 minute
- Size: 1000 entries
- Use: Hot data (current user session)

L2: Redis
- TTL: 5-60 minutes
- Size: 100GB+
- Use: API responses, session data

L3: CDN (CloudFlare)
- TTL: 1 hour - 24 hours
- Use: Static assets, public content
```

**Cache Invalidation**:
- Event-driven invalidation
- TTL-based expiration
- Manual invalidation via admin API

### 7. Horizontal Scalability

**API Servers**: Stateless, auto-scale based on CPU/memory

```
Configuration:
- Min replicas: 3
- Max replicas: 100
- Scale up: CPU > 70%
- Scale down: CPU < 30%
```

**Database Scaling**:

```
PostgreSQL:
- Primary: Write traffic
- Replicas: Read traffic (lag < 100ms)
- Connection pooling: pgbouncer
- Sharding: By tenant_id for 100K+ tenants

Redis:
- Cluster mode for 100GB+ data
- Sentinel for high availability
```

**Load Balancing**:
```
Algorithm: Least connections
Health checks: HTTP GET /health every 5s
Failover: Automatic, < 1s
```

### 8. Observability Stack

**Three Pillars**:

1. **Metrics** (Prometheus + Grafana)
   - Request rate, latency, error rate
   - Database connection pool stats
   - Cache hit/miss ratio
   - Business metrics (enrollments, attendance)

2. **Logging** (Structured JSON logs + Loki)
   - All requests logged
   - Error logs with stack traces
   - Audit logs for compliance

3. **Tracing** (Jaeger + OpenTelemetry)
   - Distributed request tracing
   - Performance bottleneck identification
   - Service dependency mapping

**SLI/SLO**:
```
Target SLOs:
- Availability: 99.9% (43 min downtime/month)
- Latency P95: < 200ms
- Latency P99: < 500ms
- Error Rate: < 0.1%
```

## Database Schema Design

### Multi-Tenancy Implementation

```sql
-- All tables include tenant_id
CREATE TABLE students (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    -- other fields
);

-- Row-Level Security
ALTER TABLE students ENABLE ROW LEVEL SECURITY;

CREATE POLICY students_tenant_isolation ON students
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Application sets tenant context
SET app.current_tenant = 'tenant-uuid-here';
```

### Indexing Strategy

```sql
-- Covering indexes for common queries
CREATE INDEX idx_students_tenant_class ON students(tenant_id, current_class_id)
    INCLUDE (first_name, last_name, admission_number);

-- Partial indexes for active records
CREATE INDEX idx_active_students ON students(tenant_id, current_class_id)
    WHERE is_active = true;
```

## API Design

### RESTful Principles

```
GET    /api/v1/students       - List students (with pagination)
GET    /api/v1/students/:id   - Get student by ID
POST   /api/v1/students       - Create student
PUT    /api/v1/students/:id   - Update student
DELETE /api/v1/students/:id   - Delete student
```

### Pagination

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 1000,
    "total_pages": 50
  }
}
```

### Error Responses

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": "email: must be a valid email address"
}
```

## Frontend Architecture

### Vue.js 3 + Pinia

**State Management**:
```
stores/
├── auth.ts       - Authentication state
├── students.ts   - Student management
├── teachers.ts   - Teacher management
└── academic.ts   - Academic data
```

**Component Structure**:
```
components/
├── common/       - Reusable components
├── students/     - Student-specific components
├── teachers/     - Teacher-specific components
└── layouts/      - Layout components
```

### API Integration

```typescript
// Centralized API client with interceptors
import api from '@/services/api'

// Automatic token refresh
api.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      // Refresh token and retry
    }
  }
)
```

## Deployment Architecture

### Production Deployment

```
                    ┌─────────────┐
                    │  CloudFlare │
                    │  (CDN + WAF)│
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Load Balancer│
                    │  (HAProxy)   │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
    ┌───▼───┐         ┌───▼───┐         ┌───▼───┐
    │ API-1 │         │ API-2 │         │ API-N │
    │  Pod  │         │  Pod  │         │  Pod  │
    └───┬───┘         └───┬───┘         └───┬───┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
    ┌───▼────┐      ┌─────▼──────┐      ┌───▼────┐
    │Postgres│      │   Redis    │      │  NATS  │
    │ Primary│      │  Cluster   │      │Cluster │
    └────┬───┘      └────────────┘      └────────┘
         │
    ┌────▼────┐
    │ Replica │
    └─────────┘
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: opensms-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: opensms-api
  template:
    metadata:
      labels:
        app: opensms-api
    spec:
      containers:
      - name: api
        image: opensms/api:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Performance Optimizations

### Database Query Optimization

1. **Connection Pooling**: 25 max connections, 5 min idle
2. **Prepared Statements**: All queries use prepared statements
3. **Batch Operations**: Bulk inserts/updates where possible
4. **Index Optimization**: Covering indexes for hot queries

### Caching Strategy

```go
// Cache aside pattern
func GetStudent(id string) (*Student, error) {
    // Try cache first
    var student Student
    err := cache.Get(ctx, "student:"+id, &student)
    if err == nil {
        return &student, nil
    }

    // Cache miss, fetch from DB
    student, err = db.GetStudent(ctx, id)
    if err != nil {
        return nil, err
    }

    // Populate cache
    cache.Set(ctx, "student:"+id, student, 10*time.Minute)
    return &student, nil
}
```

### N+1 Query Prevention

```go
// Bad: N+1 queries
for _, enrollment := range enrollments {
    student := db.GetStudent(enrollment.StudentID)
}

// Good: Single query with JOIN
enrollments := db.GetEnrollmentsWithStudents()
```

## Security Considerations

### OWASP Top 10 Mitigation

1. **Injection**: Parameterized queries, input validation
2. **Broken Authentication**: JWT with refresh tokens, bcrypt hashing
3. **Sensitive Data Exposure**: TLS 1.3, AES-256 encryption
4. **XML External Entities**: JSON-only API
5. **Broken Access Control**: RBAC, RLS
6. **Security Misconfiguration**: Secure defaults, automated scanning
7. **XSS**: Output encoding, CSP headers
8. **Insecure Deserialization**: Type-safe unmarshaling
9. **Using Components with Known Vulnerabilities**: Automated dependency scanning
10. **Insufficient Logging**: Comprehensive audit logging

## Compliance & Data Privacy

### GDPR Compliance

- **Right to access**: API endpoint for data export
- **Right to deletion**: Soft delete with anonymization
- **Data portability**: Export in JSON/CSV format
- **Consent management**: Tracked in database

### FERPA Compliance

- **Access controls**: RBAC enforces who can view student records
- **Audit logging**: All accesses logged
- **Encryption**: Data at rest and in transit

## Future Enhancements

### Phase 2 Features

1. **AI-Powered Insights**
   - Predictive analytics for at-risk students
   - Automated grade predictions
   - Attendance pattern analysis

2. **Mobile Apps**
   - Native iOS/Android apps
   - Offline-first architecture
   - Real-time push notifications

3. **Advanced Reporting**
   - Custom report builder
   - Scheduled reports
   - PDF/Excel export

4. **Integration Marketplace**
   - Third-party integrations (Google Classroom, Zoom)
   - Webhook support
   - Public API for partners

### Scaling Beyond 1M Users

1. **Database Sharding**: Tenant-based horizontal sharding
2. **Multi-Region Deployment**: Global CDN, regional databases
3. **Service Extraction**: Extract high-traffic services (notifications, file processing)
4. **Caching Enhancement**: Distributed cache with consistent hashing

---

**Last Updated**: 2026-02-01
**Version**: 1.0.0
**Author**: Principal Platform Engineering Team
