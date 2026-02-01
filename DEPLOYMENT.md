# OpenSMS Deployment Guide

## Prerequisites

- Docker & Docker Compose
- Go 1.22+
- Node.js 20+
- PostgreSQL 16+
- Redis 7+
- NATS 2.10+
- MinIO or AWS S3

## Local Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/samuelayo/opensms.git
cd opensms
```

### 2. Start Infrastructure Services

```bash
# Start PostgreSQL, Redis, NATS, MinIO, Jaeger, Prometheus, Grafana
docker-compose up -d
```

### 3. Backend Setup

```bash
cd backend

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
# IMPORTANT: Change JWT secrets and encryption key in production!

# Install dependencies
go mod download

# Install development tools
make setup

# Run database migrations
make migrate-up

# Start backend server
make dev  # With hot reload
# or
make run  # Without hot reload
```

Backend will be available at `http://localhost:8080`

### 4. Frontend Setup

```bash
cd frontend/admin-portal

# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Start development server
npm run dev
```

Frontend will be available at `http://localhost:3000`

### 5. Access Services

- **API**: http://localhost:8080
- **Admin Portal**: http://localhost:3000
- **API Docs**: http://localhost:8080/swagger (TODO)
- **Jaeger UI**: http://localhost:16686
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **MinIO Console**: http://localhost:9001 (opensms/opensms_minio_password)

## Production Deployment

### Option 1: Docker Deployment

#### 1. Build Docker Images

```bash
# Backend
cd backend
docker build -t opensms/api:latest .

# Frontend
cd frontend/admin-portal
docker build -t opensms/admin-portal:latest .
```

#### 2. Push to Registry

```bash
docker tag opensms/api:latest your-registry.com/opensms/api:latest
docker push your-registry.com/opensms/api:latest

docker tag opensms/admin-portal:latest your-registry.com/opensms/admin-portal:latest
docker push your-registry.com/opensms/admin-portal:latest
```

#### 3. Deploy with Docker Compose

```bash
# Production docker-compose.yml
docker-compose -f docker-compose.prod.yml up -d
```

### Option 2: Kubernetes Deployment

#### 1. Create Namespace

```bash
kubectl create namespace opensms
```

#### 2. Create Secrets

```bash
kubectl create secret generic opensms-secrets \
  --from-literal=db-password='your-secure-password' \
  --from-literal=jwt-access-secret='your-jwt-access-secret' \
  --from-literal=jwt-refresh-secret='your-jwt-refresh-secret' \
  --from-literal=encryption-key='your-32-byte-encryption-key' \
  -n opensms
```

#### 3. Deploy PostgreSQL (Managed or StatefulSet)

```bash
# Use managed PostgreSQL (recommended)
# Or deploy using StatefulSet
kubectl apply -f infrastructure/kubernetes/postgres.yaml -n opensms
```

#### 4. Deploy Redis

```bash
kubectl apply -f infrastructure/kubernetes/redis.yaml -n opensms
```

#### 5. Deploy NATS

```bash
kubectl apply -f infrastructure/kubernetes/nats.yaml -n opensms
```

#### 6. Deploy API

```bash
kubectl apply -f infrastructure/kubernetes/api-deployment.yaml -n opensms
kubectl apply -f infrastructure/kubernetes/api-service.yaml -n opensms
```

#### 7. Deploy Ingress

```bash
kubectl apply -f infrastructure/kubernetes/ingress.yaml -n opensms
```

### Option 3: Cloud Platform Deployment

#### AWS ECS/EKS

```bash
# Use Terraform
cd infrastructure/terraform/aws
terraform init
terraform plan
terraform apply
```

#### Google Cloud Run

```bash
# Deploy API
gcloud run deploy opensms-api \
  --image gcr.io/your-project/opensms-api:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# Deploy Frontend
gcloud run deploy opensms-admin \
  --image gcr.io/your-project/opensms-admin:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## Environment Configuration

### Critical Production Settings

**Backend (.env)**:

```bash
# SECURITY: Change these values!
JWT_ACCESS_SECRET=<generate-with: openssl rand -base64 32>
JWT_REFRESH_SECRET=<generate-with: openssl rand -base64 32>
ENCRYPTION_KEY=<generate-with: openssl rand -hex 16>

# Database
DB_HOST=your-postgres-host
DB_PASSWORD=your-secure-password
DB_SSLMODE=require

# Redis
REDIS_PASSWORD=your-redis-password

# Production settings
APP_ENV=production
LOG_LEVEL=warn
ALLOWED_ORIGINS=https://yourdomain.com
```

**Frontend (.env.production)**:

```bash
VITE_API_URL=https://api.yourdomain.com/api/v1
VITE_ENABLE_ANALYTICS=true
```

## Database Migrations

### Apply Migrations

```bash
# Development
make migrate-up

# Production (from CI/CD or manually)
goose -dir migrations postgres "host=prod-db port=5432 user=opensms password=xxx dbname=opensms sslmode=require" up
```

### Rollback Migrations

```bash
make migrate-down
```

### Create New Migration

```bash
make migrate-create NAME=add_new_feature
```

## Monitoring & Observability

### Metrics

Access Prometheus at `/metrics` endpoint:

```
opensms_http_requests_total
opensms_http_request_duration_seconds
opensms_db_queries_total
opensms_cache_hits_total
opensms_cache_misses_total
opensms_active_users
opensms_students_enrolled
```

### Grafana Dashboards

Import dashboards from `infrastructure/grafana/dashboards/`:

1. API Performance Dashboard
2. Database Metrics Dashboard
3. Business Metrics Dashboard
4. Security & Audit Dashboard

### Alerting

Configure Prometheus Alertmanager:

```yaml
groups:
  - name: opensms-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(opensms_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"

      - alert: DatabaseDown
        expr: up{job="postgres"} == 0
        for: 1m
        annotations:
          summary: "Database is down"

      - alert: HighLatency
        expr: histogram_quantile(0.95, opensms_http_request_duration_seconds) > 0.5
        for: 5m
        annotations:
          summary: "API latency P95 > 500ms"
```

## Backup & Disaster Recovery

### Database Backups

```bash
# Automated daily backups
0 2 * * * pg_dump -h localhost -U opensms opensms | gzip > /backups/opensms-$(date +\%Y\%m\%d).sql.gz

# Restore from backup
gunzip < opensms-20260201.sql.gz | psql -h localhost -U opensms opensms
```

### Object Storage Backups

MinIO versioning enabled by default. Configure S3 lifecycle policies for archival.

### Configuration Backups

Store in version control (Git). Sensitive values in secret management system.

## Security Hardening

### SSL/TLS Configuration

```nginx
# Nginx SSL Configuration
ssl_protocols TLSv1.3;
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256';
ssl_prefer_server_ciphers on;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

### Firewall Rules

```bash
# Allow only necessary ports
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw enable
```

### Rate Limiting

Already implemented in application. Additional layer via Nginx:

```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
}
```

## Scaling Guidelines

### Horizontal Scaling

**API Servers**:
```bash
# Add more replicas
kubectl scale deployment opensms-api --replicas=10 -n opensms
```

**Database**:
```bash
# Add read replicas
# Configure in pgbouncer for read/write splitting
```

### Vertical Scaling

```yaml
# Increase resources
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### Auto-Scaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: opensms-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: opensms-api
  minReplicas: 3
  maxReplicas: 100
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Troubleshooting

### Common Issues

**1. Database Connection Errors**

```bash
# Check database connectivity
pg_isready -h localhost -p 5432

# Check connection pool
curl http://localhost:8080/health
```

**2. High Memory Usage**

```bash
# Check memory usage
docker stats opensms-api

# Analyze Go memory profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**3. Slow API Responses**

```bash
# Check Jaeger traces
# Navigate to http://localhost:16686

# Check database slow queries
SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;
```

## CI/CD Pipeline

### GitHub Actions Example

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run tests
        run: |
          cd backend
          go test ./...

      - name: Build Docker image
        run: |
          docker build -t opensms/api:${{ github.sha }} .

      - name: Push to registry
        run: |
          docker push opensms/api:${{ github.sha }}

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/opensms-api \
            api=opensms/api:${{ github.sha }} -n opensms
```

## Performance Benchmarks

### Expected Performance (Single Instance)

- **Throughput**: 10,000 req/s (read-heavy workload)
- **Latency P95**: < 50ms
- **Latency P99**: < 100ms
- **Database Connections**: 25 (pooled)
- **Memory Usage**: ~200MB

### Load Testing

```bash
# Install k6
brew install k6

# Run load test
k6 run load-tests/api-load-test.js
```

---

**Need Help?**
- Documentation: https://docs.opensms.io
- GitHub Issues: https://github.com/samuelayo/opensms/issues
- Community: https://community.opensms.io
