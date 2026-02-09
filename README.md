# OpenSMS - Open School Management System

A production-grade, hyper-scalable school management system designed for primary through university level education institutions.

## Architecture

Built using **Clean Architecture** principles with a **Modular Monolith** pattern that can evolve to microservices.

### Tech Stack

**Backend:**
- Go 1.22+ (Clean Architecture, DDD)
- PostgreSQL 16+ (Primary database with multi-tenancy)
- Redis 7+ (Caching, sessions, rate limiting)
- NATS (Event bus for async communication)
- MinIO/S3 (Object storage)

**Frontend:**
- Vue.js 3 (Composition API)
- Pinia (State management)
- TypeScript
- Vite (Build tool)
- TailwindCSS (Styling)

**Infrastructure:**
- Docker & Docker Compose
- Kubernetes (Production)
- Terraform (IaC)
- GitHub Actions (CI/CD)

**Observability:**
- Prometheus (Metrics)
- Grafana (Dashboards)
- Loki (Logging)
- Jaeger (Distributed tracing)
- OpenTelemetry (Instrumentation)

### Key Features

- **Multi-tenancy**: Isolated data per institution with row-level security
- **Extreme Scalability**: Horizontal scaling, database sharding, read replicas
- **Security-First**: Zero-trust, encryption at rest/transit, RBAC, audit logging
- **Event-Driven**: Async processing via NATS for decoupling
- **CQRS**: Separate read/write models for complex queries
- **High Availability**: 99.99% uptime SLA with automated failover
- **Observability**: Full stack monitoring, tracing, and alerting
- **API-First**: RESTful APIs with OpenAPI/Swagger documentation

## Project Structure

```
opensms/
├── backend/                    # Go backend (Modular Monolith)
│   ├── cmd/                   # Application entrypoints
│   ├── internal/              # Private application code
│   │   ├── modules/          # Business domains (bounded contexts)
│   │   ├── shared/           # Shared kernel
│   │   └── infrastructure/   # Infrastructure layer
│   ├── pkg/                  # Public libraries
│   ├── migrations/           # Database migrations
│   └── tests/                # Integration tests
├── frontend/                  # Vue.js frontends
│   ├── admin-portal/         # Admin dashboard
│   ├── teacher-portal/       # Teacher dashboard
│   ├── student-portal/       # Student dashboard
│   ├── parent-portal/        # Parent dashboard
│   └── shared/               # Shared components
├── infrastructure/            # Infrastructure as Code
│   ├── docker/               # Docker configs
│   ├── kubernetes/           # K8s manifests
│   └── terraform/            # Terraform modules
├── docs/                     # Documentation
└── scripts/                  # Automation scripts
```

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16+
- Redis 7+

### Development Setup

```bash
# Clone repository
git clone https://github.com/samuelayo/opensms.git
cd opensms

# Start infrastructure services
docker-compose up -d postgres redis nats minio

# Run backend
cd backend
go mod download
make migrate-up
make run

# Run frontend (separate terminal)
cd frontend/admin-portal
npm install
npm run dev
```

## Security Features

- **Authentication**: JWT + Refresh Token rotation
- **Authorization**: Role-Based Access Control (RBAC) with fine-grained permissions
- **Encryption**: AES-256 at rest, TLS 1.3 in transit
- **Multi-tenancy**: Row-level security, tenant isolation
- **Rate Limiting**: Per-user, per-tenant, per-endpoint
- **Audit Logging**: Immutable audit trail for compliance
- **Secrets Management**: HashiCorp Vault integration
- **Security Headers**: HSTS, CSP, X-Frame-Options, etc.
- **Input Validation**: Comprehensive request validation
- **SQL Injection Prevention**: Parameterized queries only
- **XSS Prevention**: Output encoding, CSP
- **CSRF Protection**: Token-based

## Performance & Scale

- **Horizontal Scaling**: Stateless API servers behind load balancer
- **Database Sharding**: By tenant for extreme scale
- **Read Replicas**: Async replication for read-heavy workloads
- **Caching Strategy**: Multi-layer (L1: In-memory, L2: Redis, L3: CDN)
- **CDN**: Static assets and public content
- **Connection Pooling**: Optimized database connections
- **Query Optimization**: Indexed queries, query plans
- **Async Processing**: Background jobs via NATS
- **Rate Limiting**: Protect against abuse

## Compliance

- GDPR compliant (data privacy, right to be forgotten)
- FERPA compliant (student record privacy)
- SOC 2 Type II ready
- ISO 27001 ready
- Audit logging for compliance

## License

GNU General Public License v2.0

## Contributors

Built with enterprise-grade practices for extreme scale.
