# NetBox → Go+Vue.js Rewrite: Comprehensive Test Plan

> **Historical planning artifact — superseded.** The coverage inventory, priorities, and checklists below are retained for context and are not current test status. Use the canonical [testing strategy](TESTING.md), [status](STATUS.md), and [roadmap](ROADMAP.md).

> **Goal**: A test-driven roadmap for rewriting Python NetBox into Go (Gin+GORM backend)
> with Vue.js frontend, exposing both REST HTTP API and gRPC interfaces.

---

## Table of Contents

1. [Testing Strategy Overview](#1-testing-strategy-overview)
2. [Test Pyramid & Tooling](#2-test-pyramid--tooling)
3. [Current Coverage Inventory](#3-current-coverage-inventory)
4. [Backend Test Plan (Go)](#4-backend-test-plan-go)
5. [API Compatibility Test Plan (DRF Parity)](#5-api-compatibility-test-plan-drf-parity)
6. [gRPC Interface Test Plan](#6-grpc-interface-test-plan)
7. [Frontend Test Plan (Vue.js)](#7-frontend-test-plan-vuejs)
8. [Integration / E2E Test Plan](#8-integration--e2e-test-plan)
9. [Performance & Load Testing](#9-performance--load-testing)
10. [Migration Parity Testing](#10-migration-parity-testing)
11. [CI/CD Pipeline Integration](#11-cicd-pipeline-integration)
12. [Test Data & Fixtures](#12-test-data--fixtures)
13. [Execution Matrix & Priorities](#13-execution-matrix--priorities)

---

## 1. Testing Strategy Overview

### Principles

- **Test-Driven Rewrite**: Every Python NetBox feature being ported gets a test *first*,
  then the Go implementation.
- **Parity Verification**: Side-by-side comparison of Python NetBox API responses vs. Go API
  responses using the official NetBox OpenAPI schema (`netbox/contrib/openapi.json`) as the
  contract.
- **Layered Testing**: Unit → Integration → E2E, with clear boundaries.
- **Mock External Dependencies**: Database (sqlmock), Redis (miniredis), gRPC clients, HTTP
  clients — all mocked in unit tests; real in integration tests.

### Rewrite Phases (Test-Gated)

| Phase | Scope | Test Gate |
|-------|-------|-----------|
| **P0** | Models + DAO CRUD (✅ done) | DAO unit tests pass for all 146 models |
| **P1** | DRF-compatible REST API (✅ in progress) | Router tests + API parity tests |
| **P2** | Business logic & validation | Domain logic unit tests |
| **P3** | gRPC interface | Proto conformance tests |
| **P4** | Vue.js frontend | Component + E2E tests |
| **P5** | Feature parity sign-off | Full regression vs. Python NetBox |

---

## 2. Test Pyramid & Tooling

```
           /\
          /E2E\          ← Playwright (frontend), docker-compose integration
         /------\
        /  API   \        ← HTTP integration tests (Go test + testcontainers)
       / Parity   \
      /------------\
     / Integration  \     ← Real Postgres + Redis (testcontainers)
    /----------------\
   /   Unit Tests     \   ← Go testing, sqlmock, miniredis, Vitest
  /--------------------\
```

### Backend Tooling

| Tool | Purpose | Status |
|------|---------|--------|
| `testing` (stdlib) | Go unit test runner | ✅ In use |
| `github.com/DATA-DOG/go-sqlmock` | Mock database for DAO tests | ✅ In use |
| `github.com/stretchr/testify` | Assertions | ✅ In use |
| `github.com/redis/redismock/v8` | Mock Redis for cache tests | 🔲 Add |
| `github.com/testcontainers/testcontainers-go` | Real Postgres/Redis integration | 🔲 Add |
| `github.com/gin-gonic/gin` (TestMode) | HTTP handler testing | ✅ In use |
| `google.golang.org/grpc/test/bufconn` | In-memory gRPC testing | 🔲 Add (P3) |
| `github.com/benchttp/runner` | HTTP performance benchmarks | 🔲 Add (P5) |

### Frontend Tooling

| Tool | Purpose | Status |
|------|---------|--------|
| `vitest` | Unit/component test runner | 🔲 Add |
| `@vue/test-utils` | Vue component mounting | 🔲 Add |
| `@testing-library/vue` | DOM interaction testing | 🔲 Add |
| `msw` (Mock Service Worker) | API mocking | 🔲 Add |
| `playwright` | E2E browser testing | 🔲 Add |
| `@playwright/test` | E2E test framework | 🔲 Add |

---

## 3. Current Coverage Inventory

### What Exists (✅)

| Layer | Files | Coverage |
|-------|-------|----------|
| **DAO unit tests** | 146 files in `internal/dao/` | Create, Update, Delete, GetByID, GetByColumns, GetByCondition, GetByIDs, GetByLastID, Tx variants |
| **Cache unit tests** | 146 files in `internal/cache/` | Set, Get, Delete, cache miss |
| **Service client tests** | ~120 files in `internal/service/` | gRPC client method coverage |
| **Router tests** | `netbox_drf_test.go` | Search, filter, pagination, ordering, normalize body, FK resolution |
| **Handler auth tests** | `auth_test.go` | Auth middleware |

### What's Missing (🔲)

| Area | Priority | Description |
|------|----------|-------------|
| **API parity tests** | P1 | HTTP-level tests comparing Go API output to Python NetBox schema |
| **Business logic tests** | P2 | Validation rules, computed fields, model methods |
| **gRPC server tests** | P3 | Server-side handler tests |
| **Frontend component tests** | P4 | Vue component mounting and interaction |
| **E2E tests** | P4-P5 | Full browser-based user flows |
| **Migration tests** | P5 | Database schema parity with Django migrations |

---

## 4. Backend Test Plan (Go)

### 4.1 DAO Layer — COMPLETE ✅

All 146 models have comprehensive DAO tests using `sqlmock`. Pattern:

```go
// File: internal/dao/dcimSite_test.go (representative)
func Test_dcimSiteDao_Create(t *testing.T) {
    d := newDcimSiteDao()
    defer d.Close()
    testData := d.TestData.(*model.DcimSite)

    d.SQLMock.ExpectBegin()
    d.SQLMock.ExpectExec("INSERT INTO .*").
        WithArgs(d.GetAnyArgs(testData)...).
        WillReturnResult(sqlmock.NewResult(1, 1))
    d.SQLMock.ExpectCommit()

    err := d.IDao.(DcimSiteDao).Create(d.Ctx, testData)
    assert.NoError(t, err)
}
```

**Test methods per model**: `Create`, `DeleteByID`, `UpdateByID`, `GetByID`,
`GetByColumns`, `GetByCondition`, `GetByIDs`, `GetByLastID`, `CreateByTx`,
`DeleteByTx`, `UpdateByTx`.

### 4.2 Cache Layer — COMPLETE ✅

All 146 models have cache tests covering Set/Get/Delete with Redis mock.

### 4.3 Service Layer (gRPC Clients) — COMPLETE ✅

~120 service client tests covering all gRPC client methods.

### 4.4 Router / Handler Layer — IN PROGRESS

**Existing tests** (`netbox_drf_test.go`):

| Test | Covers |
|------|--------|
| `TestBuildSearchFilter` | ILIKE search construction |
| `TestEscapeSQL` | SQL injection prevention |
| `TestParsePagination` | Limit/offset parsing, defaults, bounds |
| `TestParseOrdering` | Sort field whitelist, direction |
| `TestParseIDList` | Comma-separated ID parsing |
| `TestNormalizeRequestBody` | Nested FK object flattening, null handling |
| `TestBuildPaginationURLs` | Next/previous URL generation |
| `TestNormalizeListResults` | Slice type conversion |
| `TestRegisterModelEndpoint` | Registry population |
| `TestSitesRegistry` | End-to-end registry config verification |
| `TestSchemeFromRequest` | Protocol detection (X-Forwarded-Proto) |

**Tests to add** (P1):

| Test | File | Description |
|------|------|-------------|
| `TestDRFList_EmptyResult` | `netbox_drf_integration_test.go` | GET list returns `{count:0, results:[]}` |
| `TestDRFList_WithPagination` | same | Verify `limit`, `offset`, `next`, `previous` |
| `TestDRFList_WithSearch` | same | `?q=keyword` triggers ILIKE search |
| `TestDRFList_WithFilter` | same | `?status=active` filters correctly |
| `TestDRFList_WithOrdering` | same | `?ordering=-name` sorts descending |
| `TestDRFRetrieve` | same | GET single object by ID |
| `TestDRFCreate` | same | POST creates object, returns 201 |
| `TestDRFCreate_WithNestedFK` | same | POST with nested `{"region":{"id":5}}` flattens to `region_id` |
| `TestDRFUpdate_PUT` | same | PUT replaces object |
| `TestDRFUpdate_PATCH` | same | PATCH partial update |
| `TestDRFDelete` | same | DELETE removes object, returns 204 |
| `TestDRFBulkDelete` | same | POST to `/delete/` with ID list |
| `TestDRF_NestedFKResolution` | same | Verify FK table resolution (`region` → `dcim_region`) |

### 4.5 Business Logic Tests (P2)

Python NetBox has domain logic that must be ported and tested:

| Module | Test File | Key Logic |
|--------|-----------|-----------|
| DCIM | `dcim_business_test.go` | Cable path computation, interface ordering, rack position validation |
| IPAM | `ipam_business_test.go` | Prefix containment, IP address assignment, VLAN uniqueness |
| Circuits | `circuits_business_test.go` | Circuit termination validation |
| VPN | `vpn_business_test.go` | Tunnel termination, IPsec policy validation |

**Reference**: `netbox/netbox/dcim/tests/test_models.py`, `test_cablepaths.py`, etc.

---

## 5. API Compatibility Test Plan (DRF Parity)

This is the **critical** test layer — it verifies the Go API produces output compatible
with the Python NetBox REST API.

### 5.1 Contract Source

Use `netbox/contrib/openapi.json` as the API contract.

### 5.2 Parity Test Structure

```
netbox-backend/test/parity/
├── parity_test.go           # Test harness: starts Go server, runs all parity checks
├── fixtures/                # JSON request/response fixtures from Python NetBox
│   ├── dcim/
│   │   ├── sites_list.json
│   │   ├── sites_create.json
│   │   └── sites_retrieve.json
│   ├── ipam/
│   └── ...
├── comparator.go            # Deep comparison logic (field-by-field)
└── snapshot.go              # Snapshot testing for response shapes
```

### 5.3 Parity Test Cases

For **each** of the 115 registered endpoints:

| Test Case | Method | Path | Assertion |
|-----------|--------|------|-----------|
| List shape | GET | `/api/{module}/{resource}/` | Response has `count`, `next`, `previous`, `results[]` |
| List pagination | GET | `?limit=10&offset=20` | `next`/`previous` URLs correct |
| List search | GET | `?q=test` | Search applied to configured columns |
| List filter | GET | `?{field}=value` | Filter applied |
| List ordering | GET | `?ordering=name` | Results sorted |
| Retrieve | GET | `/{id}/` | Single object with all fields |
| Create | POST | `/` | 201 status, object returned with `id` |
| Create FK | POST | `/` | Nested FK objects accepted |
| Update | PUT | `/{id}/` | 200, fields updated |
| Partial update | PATCH | `/{id}/` | 200, only sent fields updated |
| Delete | DELETE | `/{id}/` | 204 No Content |
| Bulk delete | POST | `/delete/` | 204 or bulk response |

### 5.4 Field-Level Parity Rules

```go
// Comparator rules for Go vs Python response parity:
// 1. All Python response fields must exist in Go response
// 2. Nested FK objects: {id, url, display, name, slug} sub-objects
// 3. Timestamps: ISO 8601 format (created, last_updated)
// 4. Computed fields: url (absolute URL), display (human-readable)
// 5. Nullable fields: null vs missing key
```

---

## 6. gRPC Interface Test Plan

### 6.1 Proto Conformance

Each gRPC service method must have:

| Test Type | Description |
|-----------|-------------|
| Happy path | Valid request → expected response |
| Not found | Non-existent ID → `codes.NotFound` |
| Invalid argument | Missing required field → `codes.InvalidArgument` |
| Pagination | Large result sets paged correctly |
| Auth | Missing/invalid token → `codes.Unauthenticated` |

### 6.2 Test Structure

```go
// Using bufconn for in-memory gRPC testing (no port allocation)
func setupGRPCServer(t *testing.T) (*grpc.ClientConn, func()) {
    lis := bufconn.Listen(1024 * 1024)
    srv := grpc.NewServer()
    pb.RegisterDcimSiteServiceServer(srv, &server{})
    go srv.Serve(lis)

    conn, _ := grpc.DialContext(context.Background(),
        "bufnet",
        grpc.WithContextDialer(bufconn.Dialer(lis)),
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    return conn, srv.Stop
}
```

### 6.3 Service Coverage

All ~120 generated gRPC services need server-side tests. Template:

```go
func TestSiteService_Create(t *testing.T)          // happy path
func TestSiteService_Create_Validation(t *testing.T) // invalid input
func TestSiteService_GetByID(t *testing.T)
func TestSiteService_GetByID_NotFound(t *testing.T)
func TestSiteService_List(t *testing.T)
func TestSiteService_Update(t *testing.T)
func TestSiteService_Delete(t *testing.T)
```

---

## 7. Frontend Test Plan (Vue.js)

### 7.1 Component Unit Tests (Vitest + @vue/test-utils)

| Component | Test File | Key Assertions |
|-----------|-----------|----------------|
| `DataTable.vue` | `DataTable.spec.ts` | Renders rows, columns, sorting, pagination |
| `FilterPanel.vue` | `FilterPanel.spec.ts` | Filter application, clear, URL sync |
| `ObjectListView.vue` | `ObjectListView.spec.ts` | Fetches data, renders table, handles loading/error |
| `ScriptsView.vue` | `ScriptsView.spec.ts` | Script list, execute, results display |

### 7.2 Store/API Client Tests

| Module | Test | Assertions |
|--------|------|------------|
| API client (`api/`) | `api.spec.ts` | Request building, auth headers, error handling |
| Router models | `model-registry.spec.ts` | Registry completeness (all models registered) |
| Types | `types.spec.ts` | Type guards, serialization |

### 7.3 Test Setup

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    coverage: { provider: 'v8', reporter: ['text', 'lcov'] }
  }
})
```

### 7.4 MSW for API Mocking

```typescript
// tests/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/dcim/sites/', () => {
    return HttpResponse.json({
      count: 2,
      next: null,
      previous: null,
      results: [
        { id: 1, name: 'Site A', slug: 'site-a' },
        { id: 2, name: 'Site B', slug: 'site-b' },
      ]
    })
  }),
]
```

---

## 8. Integration / E2E Test Plan

### 8.1 Integration Tests (testcontainers)

```
netbox-backend/test/integration/
├── integration_test.go       # Test harness with real Postgres + Redis
├── crud_flow_test.go          # Full CRUD lifecycle per model
├── fk_resolution_test.go      # FK nested object resolution with real DB
└── pagination_test.go         # Pagination with real data volumes
```

**Setup**:

```go
//go:build integration

func TestMain(m *testing.M) {
    // Start Postgres container
    pgContainer, _ = postgres.Run(ctx, "postgres:16-alpine")
    // Start Redis container
    redisContainer, _ = redis.Run(ctx, "redis:7-alpine")
    // Run migrations
    database.AutoMigrate(db)
    // Seed test data
    os.Exit(m.Run())
}
```

### 8.2 E2E Tests (Playwright)

| Flow | Test | Steps |
|------|------|-------|
| Site management | `site-crud.spec.ts` | Login → Create site → View → Edit → Delete |
| Device management | `device-crud.spec.ts` | Create device type → Create device → Assign to site |
| IPAM | `ipam-flow.spec.ts` | Create VRF → Create prefix → Create IP address |
| Search & filter | `search-filter.spec.ts` | Search across models → Apply filters → Verify results |
| Multi-tenancy | `tenancy.spec.ts` | Create tenant → Assign to objects → Filter by tenant |

---

## 9. Performance & Load Testing

### 9.1 Benchmarks

```go
// File: internal/routers/benchmark_test.go
func BenchmarkDRFList(b *testing.B) {
    // Benchmark list endpoint with 10k records
}

func BenchmarkFKResolution(b *testing.B) {
    // Benchmark nested FK object resolution
}
```

### 9.2 Load Testing (k6)

```javascript
// tests/load/list_sites.js
import http from 'k6/http'
export const options = { vus: 50, duration: '30s' }
export default function () {
    http.get('http://localhost:8080/api/dcim/sites/?limit=50', {
        headers: { Authorization: `Token ${__ENV.TOKEN}` }
    })
}
```

**Targets** (vs. Python NetBox baseline):

| Metric | Target | Python Baseline |
|--------|--------|-----------------|
| List (50 items) p99 | < 50ms | ~200ms |
| Retrieve p99 | < 10ms | ~50ms |
| Create p99 | < 30ms | ~100ms |
| Concurrent users | 500+ | ~100 |

---

## 10. Migration Parity Testing

### 10.1 Schema Parity

Verify Go GORM auto-migration produces schema identical to Django migrations.

```go
//go:build migration
func TestSchemaParity(t *testing.T) {
    // 1. Run Django migrations on DB A
    // 2. Run GORM AutoMigrate on DB B
    // 3. Compare information_schema.columns
}
```

### 10.2 Data Migration Tests

| Test | Description |
|------|-------------|
| Column types | `varchar(N)`, `integer`, `boolean`, `inet`, `cidr` match |
| Constraints | NOT NULL, UNIQUE, CHECK match |
| Foreign keys | All FKs present, correct ON DELETE behavior |
| Indexes | All Django indexes replicated |

---

## 11. CI/CD Pipeline Integration

### Jenkinsfile Stages

```groovy
pipeline {
    agent any
    stages {
        stage('Backend Unit Tests') {
            steps {
                sh 'cd netbox-backend && go test ./internal/... -coverprofile=coverage.out'
                sh 'cd netbox-backend && go tool cover -func=coverage.out'
            }
        }
        stage('Backend Build') {
            steps { sh 'cd netbox-backend && go build ./...' }
        }
        stage('Generate') {
            steps {
                sh 'node scripts/generate_drf_registry.mjs'
                sh 'cd netbox-backend && go build ./...'  // verify generated code compiles
            }
        }
        stage('Frontend Build') {
            steps { sh 'cd netbox-frontend && npm ci && npm run build' }
        }
        stage('Frontend Tests') {
            steps { sh 'cd netbox-frontend && npm run test:unit' }
        }
        stage('Integration Tests') {
            steps { sh 'cd netbox-backend && go test -tags=integration ./test/integration/...' }
        }
        stage('API Parity') {
            steps { sh 'cd netbox-backend && go test -tags=parity ./test/parity/...' }
        }
        stage('E2E Tests') {
            steps { sh 'cd netbox-frontend && npx playwright test' }
        }
        stage('Docker Build') {
            steps { sh 'docker compose build' }
        }
    }
}
```

### Coverage Gates

| Layer | Minimum Coverage | Current |
|-------|-----------------|---------|
| DAO | 90% | ~95% ✅ |
| Cache | 85% | ~90% ✅ |
| Router/Handler | 80% | ~40% 🔲 |
| Service | 85% | ~85% ✅ |
| Frontend components | 70% | 0% 🔲 |

---

## 12. Test Data & Fixtures

### 12.1 Fixture Generation

```bash
# Export fixtures from Python NetBox for parity testing
python manage.py dumpdata dcim.Site --natural-foreign > fixtures/dcim_sites.json
python manage.py dumpdata ipam.Prefix > fixtures/ipam_prefixes.json
```

### 12.2 Go Test Factories

```go
// internal/testutil/factory.go
package testutil

func NewSite() *model.DcimSite {
    return &model.DcimSite{
        Name:   "Test Site " + random.String(6),
        Slug:   "test-site-" + random.String(6),
        Status: "active",
    }
}

func NewDevice() *model.DcimDevice { ... }
func NewPrefix() *model.IpamPrefix { ... }
```

### 12.3 Database Seeder

```go
// test/integration/seeder.go
func SeedTestData(t *testing.T, db *gorm.DB) {
    sites := []model.DcimSite{ *testutil.NewSite(), *testutil.NewSite() }
    db.Create(&sites)
    // ... seed related data
}
```

---

## 13. Execution Matrix & Priorities

### Phase 1: Core API Parity (Weeks 1-3)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 1.1 | Add HTTP integration test harness (gin TestMode) | 🔴 High | 1d | 🔲 |
| 1.2 | Write `TestDRFList_*` tests for Sites (reference model) | 🔴 High | 1d | 🔲 |
| 1.3 | Write `TestDRFRetrieve` / `Create` / `Update` / `Delete` for Sites | 🔴 High | 2d | 🔲 |
| 1.4 | Parameterize DRF tests to cover all 115 endpoints | 🔴 High | 2d | 🔲 |
| 1.5 | Add FK resolution integration test | 🟡 Med | 1d | 🔲 |
| 1.6 | Set up testcontainers for integration tests | 🟡 Med | 2d | 🔲 |

### Phase 2: Business Logic (Weeks 3-5)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 2.1 | Port DCIM validation rules + tests | 🔴 High | 5d | 🔲 |
| 2.2 | Port IPAM logic (prefix containment, IP assignment) + tests | 🔴 High | 5d | 🔲 |
| 2.3 | Port Circuits/VPN/Wireless logic + tests | 🟡 Med | 3d | 🔲 |

### Phase 3: gRPC Server Tests (Weeks 5-6)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 3.1 | Add bufconn test harness | 🟡 Med | 1d | 🔲 |
| 3.2 | Write server-side tests for top 20 services | 🟡 Med | 3d | 🔲 |
| 3.3 | Add proto conformance tests | 🟢 Low | 2d | 🔲 |

### Phase 4: Frontend Tests (Weeks 6-8)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 4.1 | Set up Vitest + @vue/test-utils + msw | 🟡 Med | 1d | 🔲 |
| 4.2 | Component tests for DataTable, FilterPanel, ObjectListView | 🔴 High | 3d | 🔲 |
| 4.3 | Set up Playwright E2E | 🟡 Med | 1d | 🔲 |
| 4.4 | E2E flows: Site CRUD, Device CRUD, IPAM | 🟡 Med | 3d | 🔲 |

### Phase 5: Parity Sign-off (Weeks 8-10)

| # | Task | Priority | Effort | Status |
|---|------|----------|--------|--------|
| 5.1 | Snapshot-based API parity tests (all 115 endpoints) | 🔴 High | 5d | 🔲 |
| 5.2 | Schema migration parity tests | 🟡 Med | 3d | 🔲 |
| 5.3 | Load testing + benchmarking | 🟡 Med | 2d | 🔲 |
| 5.4 | Full regression run vs. Python NetBox | 🔴 High | 2d | 🔲 |

---

## Appendix A: Test Commands Cheat Sheet

```bash
# Backend unit tests (fast, no external deps)
cd netbox-backend && go test ./internal/... -v

# Backend with coverage
cd netbox-backend && go test ./internal/... -coverprofile=coverage.out -covermode=atomic
cd netbox-backend && go tool cover -html=coverage.out -o coverage.html

# Specific package
cd netbox-backend && go test ./internal/routers/ -v -run TestSitesRegistry

# Integration tests (requires Docker for testcontainers)
cd netbox-backend && go test -tags=integration ./test/integration/... -v

# API parity tests
cd netbox-backend && go test -tags=parity ./test/parity/... -v

# Regenerate DRF registry (after adding/modifying models)
node scripts/generate_drf_registry.mjs

# Frontend tests
cd netbox-frontend && npm run test:unit

# Frontend E2E
cd netbox-frontend && npx playwright test

# Full pipeline
make test-all
```

## Appendix B: Python NetBox Test Reference

The Python NetBox test suite provides the specification for expected behavior:

| Python Test File | Go Equivalent | Priority |
|------------------|---------------|----------|
| `dcim/tests/test_models.py` | `dcim_models_test.go` | P2 |
| `dcim/tests/test_api.py` | `dcim_api_parity_test.go` | P1 |
| `dcim/tests/test_filtersets.py` | `dcim_filters_test.go` | P1 |
| `dcim/tests/test_cablepaths.py` | `dcim_cablepaths_test.go` | P2 |
| `dcim/tests/test_forms.py` | `dcim_validation_test.go` | P2 |
| `ipam/tests/test_*.py` | `ipam_*_test.go` | P1-P2 |
| `circuits/tests/test_*.py` | `circuits_*_test.go` | P2 |
| `vpn/tests/test_*.py` | `vpn_*_test.go` | P2 |

Each Python test class/method should have a corresponding Go test that verifies
the same behavior, adapted to Go idioms.
