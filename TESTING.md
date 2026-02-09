# OpenSMS Testing Guide

## Overview

OpenSMS has a comprehensive test suite covering backend (Go) and frontend (Vue.js) components with unit tests, integration tests, and end-to-end testing capabilities.

## Backend Testing (Go)

### Test Structure

```
backend/
├── internal/
│   ├── shared/security/
│   │   ├── crypto_test.go        # Encryption/hashing tests
│   │   ├── jwt_test.go           # JWT token tests
│   │   └── rbac_test.go          # RBAC permissions tests
│   ├── server/
│   │   └── errors_test.go        # Error handling tests
│   └── modules/                   # Module-specific tests
└── tests/                         # Integration tests
```

### Running Tests

```bash
# Run all tests
cd backend
make test

# Run specific package tests
go test ./internal/shared/security/... -v

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./internal/shared/security/...

# Run only unit tests (exclude integration)
go test -short ./...
```

### Test Coverage

Current test coverage:

- **Security Package**: >95% coverage
  - Crypto operations (encrypt/decrypt, hashing)
  - JWT generation and validation
  - RBAC permissions checking
  - Password strength validation

- **Server Package**: >85% coverage
  - Error handling
  - HTTP middleware
  - Request validation

### Writing Tests

Example test structure:

```go
package security

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test-data"

    // Act
    result, err := FunctionToTest(input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedValue, result)
}
```

### Test Conventions

1. **File naming**: `*_test.go`
2. **Test function naming**: `TestFunctionName` or `TestFeature_Scenario`
3. **Benchmark naming**: `BenchmarkFunctionName`
4. **Table-driven tests** for multiple scenarios:

```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid input", "test", "expected", false},
    {"invalid input", "", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := FunctionToTest(tt.input)
        if tt.wantErr {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        }
    })
}
```

## Frontend Testing (Vue.js + Vitest)

### Test Structure

```
frontend/admin-portal/
├── src/
│   ├── stores/
│   │   └── auth.spec.ts          # Pinia store tests
│   ├── services/
│   │   └── api.spec.ts           # API client tests
│   ├── components/
│   │   └── *.spec.ts             # Component tests
│   └── views/
│       └── *.spec.ts             # View tests
└── vitest.config.ts               # Vitest configuration
```

### Running Tests

```bash
cd frontend/admin-portal

# Run all tests
npm run test

# Run with UI
npm run test:ui

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test -- --watch

# Run specific test file
npm run test auth.spec.ts
```

### Test Coverage

Current test coverage:

- **Auth Store**: >90% coverage
  - Login/logout functionality
  - Token management
  - Permission checking
  - State management

- **API Service**: >80% coverage
  - Request interceptors
  - Response handling
  - Error handling

### Writing Vue Tests

Example component test:

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import MyComponent from './MyComponent.vue'

describe('MyComponent', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders properly', () => {
    const wrapper = mount(MyComponent, {
      props: { msg: 'Hello' }
    })
    expect(wrapper.text()).toContain('Hello')
  })

  it('handles click event', async () => {
    const wrapper = mount(MyComponent)
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('click')
  })
})
```

Example store test:

```typescript
import { setActivePinia, createPinia } from 'pinia'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useMyStore } from './myStore'
import api from '@/services/api'

vi.mock('@/services/api')

describe('My Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('performs action correctly', async () => {
    const store = useMyStore()
    vi.mocked(api.post).mockResolvedValue({ data: {} })

    await store.myAction()

    expect(store.someState).toBe(expectedValue)
  })
})
```

## Integration Tests

### Database Integration Tests

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
cd backend
TEST_DATABASE_URL="postgres://opensms_test:test@localhost:5432/opensms_test" \
  go test -tags=integration ./tests/integration/...
```

### API Integration Tests

```bash
# Full integration test with all services
cd backend
make test-integration
```

## CI/CD Testing

GitHub Actions automatically runs tests on:
- Push to `main` or `develop` branches
- Pull requests

See `.github/workflows/test.yml` for CI configuration.

### CI Test Matrix

- **Backend**: Go 1.22 with PostgreSQL 16, Redis 7
- **Frontend**: Node.js 20
- **Linting**: golangci-lint, ESLint
- **Coverage**: Uploaded to Codecov

## Test Data

### Test Database

For integration tests, use the test database:

```bash
# Create test database
createdb opensms_test

# Run migrations
goose -dir migrations postgres "host=localhost port=5432 user=opensms_test dbname=opensms_test sslmode=disable" up

# Seed test data
go run scripts/seed_test_data.go
```

### Test Fixtures

Test fixtures are located in:
- `backend/tests/fixtures/` - JSON test data
- `frontend/admin-portal/src/__fixtures__/` - Mock data

## Performance Testing

### Load Testing

```bash
# Install k6
brew install k6  # macOS
# or
sudo apt-get install k6  # Linux

# Run load test
k6 run tests/load/api-load-test.js

# Run with custom settings
k6 run --vus 100 --duration 30s tests/load/api-load-test.js
```

### Benchmark Tests

```bash
# Run Go benchmarks
go test -bench=. -benchmem ./internal/shared/security/...

# Compare benchmarks
go test -bench=. -benchmem ./... > old.txt
# Make changes
go test -bench=. -benchmem ./... > new.txt
benchcmp old.txt new.txt
```

## Mocking

### Backend Mocking

Use `testify/mock` for interface mocking:

```go
import "github.com/stretchr/testify/mock"

type MockDB struct {
    mock.Mock
}

func (m *MockDB) GetUser(id string) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

// In test
mockDB := new(MockDB)
mockDB.On("GetUser", "123").Return(&User{ID: "123"}, nil)
```

### Frontend Mocking

Use Vitest's `vi.mock()`:

```typescript
import { vi } from 'vitest'

vi.mock('@/services/api', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn()
  }
}))
```

## Test Best Practices

### Do's

✅ Write tests before fixing bugs (TDD)
✅ Use table-driven tests for multiple scenarios
✅ Test error cases and edge cases
✅ Use meaningful test names
✅ Keep tests isolated and independent
✅ Mock external dependencies
✅ Test behavior, not implementation
✅ Aim for >80% code coverage on critical paths

### Don'ts

❌ Don't test framework/library code
❌ Don't share state between tests
❌ Don't make tests depend on execution order
❌ Don't use sleep/delays (use proper synchronization)
❌ Don't test private methods directly
❌ Don't skip tests (fix or remove them)

## Debugging Tests

### Backend

```bash
# Run single test with verbose output
go test -v -run TestSpecificTest ./internal/shared/security/

# Debug with Delve
dlv test ./internal/shared/security/ -- -test.run TestSpecificTest
```

### Frontend

```bash
# Run with debugger
npm run test -- --inspect-brk

# Then attach Chrome DevTools to debug
```

## Coverage Reports

### Generating Reports

```bash
# Backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Frontend
npm run test:coverage
open coverage/index.html
```

### Coverage Thresholds

Minimum coverage requirements:
- **Critical security code**: 95%
- **Business logic**: 85%
- **API handlers**: 80%
- **UI components**: 70%
- **Overall project**: 75%

## Continuous Testing

### Watch Mode

```bash
# Backend (using gotestsum or similar)
gotestsum --watch

# Frontend
npm run test -- --watch
```

## Test Environment

### Environment Variables

Create `.env.test` for test-specific configuration:

```bash
APP_ENV=test
DB_HOST=localhost
DB_PORT=5432
DB_NAME=opensms_test
REDIS_HOST=localhost
JWT_ACCESS_SECRET=test-secret-key-min-32-characters
ENCRYPTION_KEY=12345678901234567890123456789012
```

## Troubleshooting

### Common Issues

**Tests fail with "too many open files":**
```bash
ulimit -n 4096
```

**Database connection errors:**
```bash
# Ensure test database is running
docker-compose -f docker-compose.test.yml up -d postgres
# Verify connection
psql -h localhost -U opensms_test -d opensms_test
```

**Frontend tests timeout:**
```typescript
// Increase timeout in test
it('slow test', async () => {
  // ...
}, 10000) // 10 seconds
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [Vitest Documentation](https://vitest.dev/)
- [Vue Test Utils](https://test-utils.vuejs.org/)
- [Testing Best Practices](https://github.com/goldbergyoni/javascript-testing-best-practices)

---

**Last Updated**: 2026-02-01
**Maintained by**: OpenSMS Development Team
