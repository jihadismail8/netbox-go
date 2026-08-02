# API Implementation Plan: Frontend Parity with Python NetBox

> **Historical planning artifact — superseded.** The inventory, gap counts, status labels, and checklists below are not current. Use the canonical [status](STATUS.md), [roadmap](ROADMAP.md), and [compatibility contract](COMPATIBILITY.md).

> **Goal:** Bridge all gaps between the Go backend's CRUD-only APIs and the Python NetBox's rich REST API so the Vue.js frontend can fully mimic the original implementation.

## Executive Summary

The Go backend currently has **131 model-level CRUD services** (9 endpoints each via sponge framework gRPC + HTTP gateway). However, the Python NetBox API provides significantly more functionality that the frontend needs:

| Category | Python NetBox | Go Backend (Current) | Gap |
|----------|--------------|----------------------|-----|
| Model CRUD | ✅ 131 models | ✅ 131 models | ✅ Done |
| Bulk operations | ✅ All models | ❌ Missing | 🔴 Critical |
| DRF response envelope | ✅ count/next/previous/results | ❌ Raw array | 🔴 Critical |
| Page-based pagination | ✅ ?page=&page_size= | ❌ Cursor-based (lastID) | 🔴 Critical |
| Trailing slash URLs | ✅ /api/dcim/sites/ | ❌ /api/dcim/sites | 🔴 Critical |
| PATCH semantics | ✅ Partial update | ❌ PUT full replace | 🟡 High |
| Token auth | ✅ Token-based | ❌ JWT Bearer | 🟡 High |
| Custom actions (trace, paths, elevation, render-config, etc.) | ✅ 15+ actions | ❌ Missing | 🟡 High |
| Available IPs/Prefixes/ASNs | ✅ 3 custom views | ❌ Missing | 🟡 High |
| System endpoints (status, search, schema) | ✅ | ❌ Missing | 🟡 Medium |
| Nested serializers (depth, brief mode) | ✅ | ❌ Flat only | 🟡 Medium |
| FilterSet query params | ✅ Rich filtering | ❌ Basic only | 🟡 Medium |

---

## Phase 1: Core Infrastructure (Critical — must do first)

These changes affect ALL 131 models and must be completed before any per-model work.

### 1.1 HTTP Middleware Layer

**File:** `netbox-backend/internal/routers/middleware.go` (new)

Create Gin middleware that:

1. **Trailing Slash Normalization** — Redirect `/api/dcim/sites/` → match routes with or without trailing slash using `gin.RedirectTrailingSlash` and `gin.RedirectFixedPath`.

2. **DRF Response Envelope Wrapper** — Intercept list responses and wrap them:
```json
// BEFORE (current):
{ "dcimSites": [ {...}, {...} ] }

// AFTER (DRF-compatible):
{
  "count": 150,
  "next": "/api/dcim/sites/?page=2&page_size=50",
  "previous": null,
  "results": [ {...}, {...} ]
}
```

3. **Page-Based Pagination Handler** — Parse `?page=1&page_size=50&limit=&offset=` query params. Map to the existing `ListByLastID` or add a new `ListByPage` RPC.

4. **CORS Middleware** — Allow the frontend dev server origin.

### 1.2 Authentication Bridge

**File:** `netbox-backend/internal/routers/auth.go` (new)

Implement token-based authentication compatible with the frontend:
- Accept `Authorization: Token <token>` header
- Look up token in `users_token` table
- Attach user info to request context
- Fall back to session-based auth for browser access

Endpoints needed:
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/auth/login/` | Username/password → `{token, user}` |
| `POST` | `/api/auth/logout/` | Invalidate token |
| `GET` | `/api/auth/me/` | Get current user profile |
| `POST` | `/api/users/tokens/` | Create new API token |

### 1.3 Bulk Operations

**Files:** Modify all 131 proto files + add generic HTTP handlers

Two approaches:

#### Approach A: Generic HTTP Handlers (Recommended — less proto churn)

Add two generic HTTP endpoints per model route group via Gin router middleware:

| Method | Path Pattern | Body | Behavior |
|--------|-------------|------|----------|
| `DELETE` | `/api/{module}/{model}/` | `{"pk": [1,2,3]}` | Bulk delete by IDs |
| `PATCH` | `/api/{module}/{model}/` | `{"pk": [1,2,3], ...fields}` | Bulk partial update |
| `POST` | `/api/{module}/{model}/` | `[{...}, {...}]` | Bulk create (array of objects) |

This intercepts at the Gin router level and calls the existing `DeleteByIDs` / `UpdateByIDs` DAO methods.

#### Approach B: Proto-level Bulk RPCs (More work, more "correct")

Add to each proto:
```protobuf
rpc BulkCreate(BulkCreateRequest) returns (BulkCreateReply) {
  option (google.api.http) = {
    post: "/api/{base}/bulk_create/"
    body: "*"
  };
}
rpc BulkUpdate(BulkUpdateRequest) returns (BulkUpdateReply) {
  option (google.api.http) = {
    patch: "/api/{base}/bulk_update/"
    body: "*"
  };
}
rpc BulkDelete(BulkDeleteRequest) returns (BulkDeleteReply) {
  option (google.api.http) = {
    delete: "/api/{base}/bulk_delete/"
    body: "*"
  };
}
```

### 1.4 PATCH Semantics

**File:** Modify each proto's `UpdateByID` annotation from `put` to `patch`.

OR: Add middleware that treats all PUT requests as PATCH (partial updates). This is simpler and less disruptive.

### 1.5 List Endpoint Compatibility

The frontend expects:
```
GET /api/dcim/sites/?page=1&page_size=50&q=...&region_id=1
```

The backend currently supports:
```
GET /api/dcim/sites?lastID=0&limit=50
POST /api/dcim/sites/list  (with Params body)
```

**Solution:** Add a compatibility handler at `GET /api/{module}/{model}` that:
1. Parses `page`, `page_size`, `q`, and filter params from query string
2. Calls the existing DAO with offset/limit (converted from page)
3. Returns DRF-envelope response

---

## Phase 2: System-Level Endpoints

### 2.1 Status Endpoint

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/api/status/` | `{netbox-version, python-version→go-version, django-version→"", hostname, installed_apps, plugins, rq-workers-running: 0}` |

### 2.2 API Root

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/api/` | Links to all module roots: `{circuits: "/api/circuits/", core: "/api/core/", dcim: "/api/dcim/", ...}` |

### 2.3 Per-Module Root

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/api/dcim/` | Links to all DCIM model endpoints |
| `GET` | `/api/ipam/` | Links to all IPAM model endpoints |
| ... | ... | ... |

### 2.4 Search Endpoint (Optional — Phase 3)

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/api/search/?q=...` | Cross-model search results |

### 2.5 OpenAPI Schema

| Method | Path | Response |
|--------|------|----------|
| `GET` | `/api/schema/` | OpenAPI 3.1 JSON (use swag/gin-swagger) |

---

## Phase 3: Custom Actions (Non-CRUD Endpoints)

These are the Python NetBox `@action` endpoints and custom views that the frontend needs:

### 3.1 DCIM Custom Actions

| Model | Action | Method | Path | Description |
|-------|--------|--------|------|-------------|
| Interface, ConsolePort, ConsoleServerPort, PowerPort, PowerOutlet, PowerFeed | `trace` | `GET` | `/api/dcim/{model}/{id}/trace/` | Trace cable path (returns segments) |
| FrontPort, RearPort | `paths` | `GET` | `/api/dcim/{model}/{id}/paths/` | Return cable paths through port |
| Rack | `elevation` | `GET` | `/api/dcim/racks/{id}/elevation/` | Rack elevation (JSON or SVG) |
| Device (via RenderConfigMixin) | `render-config` | `POST` | `/api/dcim/devices/{id}/render-config/` | Render device config from template |
| ConnectedDevice (custom ViewSet) | `list` | `GET` | `/api/dcim/connected-device/?peer_device=&peer_interface=` | Find connected device |

**Implementation files:**
- `netbox-backend/internal/handler/dcim_custom.go` — trace, paths, elevation handlers
- `netbox-backend/internal/service/cable_trace.go` — cable path tracing logic
- `netbox-backend/internal/service/rack_elevation.go` — rack elevation calculation

### 3.2 IPAM Custom Actions (Available Addresses)

| Model | Action | Method | Path | Description |
|-------|--------|--------|------|-------------|
| Prefix | `available-ips` | `GET` | `/api/ipam/prefixes/{id}/available-ips/` | List available IPs in prefix |
| Prefix | `available-ips` | `POST` | `/api/ipam/prefixes/{id}/available-ips/` | Allocate IP from prefix |
| Prefix | `available-prefixes` | `GET` | `/api/ipam/prefixes/{id}/available-prefixes/` | List child prefix space |
| Prefix | `available-prefixes` | `POST` | `/api/ipam/prefixes/{id}/available-prefixes/` | Allocate child prefix |
| IPRange | `available-ips` | `GET` | `/api/ipam/ip-ranges/{id}/available-ips/` | List available IPs in range |
| IPRange | `available-ips` | `POST` | `/api/ipam/ip-ranges/{id}/available-ips/` | Allocate IP from range |
| ASNRange | `available-asns` | `GET` | `/api/ipam/asn-ranges/{id}/available-asns/` | List available ASNs |
| ASNRange | `available-asns` | `POST` | `/api/ipam/asn-ranges/{id}/available-asns/` | Allocate ASN |

**Implementation files:**
- `netbox-backend/internal/handler/ipam_custom.go` — available IPs/prefixes/ASNs handlers
- `netbox-backend/internal/service/ipam_available.go` — IP/prefix/ASN allocation logic

### 3.3 Extras Custom Actions

| Model | Action | Method | Path | Description |
|-------|--------|--------|------|-------------|
| ConfigContext | (brief mode) | `GET` | `/api/extras/config-contexts/?brief=true` | Return nested serializer |
| ConfigTemplate (via RenderConfigMixin) | `render-config` | `POST` | `/api/extras/config-templates/{id}/render-config/` | Render config from template |
| CustomField | `choices` | (implicit) | — | Return choice values for select fields |
| Webhook | (CRUD only) | — | — | Standard CRUD, no custom actions |

### 3.4 Core Custom Actions

| Model | Action | Method | Path | Description |
|-------|--------|--------|------|-------------|
| DataSource | `sync` | `POST` | `/api/core/data-sources/{id}/sync/` | Trigger data source sync |
| Job | (list/detail only) | — | — | Read-only job status |

---

## Phase 4: Serializer Compatibility

### 4.1 Nested Serializers (Brief Mode)

Python NetBox uses `?brief=true` to return nested/compact serializers (e.g., Site returns `{id, name, slug}` instead of all 30+ fields).

**Implementation:**
- Add `brief` query param support to list/detail handlers
- When `brief=true`, project only `id`, `name`, `display`, `slug` fields
- This reduces payload size significantly for dropdowns/select fields

### 4.2 Include/Exclude Fields

Support `?include=id,name,status` and `?exclude=config_context` query params.

### 4.3 Computed Fields

Python NetBox serializers include computed fields like:
- `*_count` fields (e.g., `site_count` on Region, `device_count` on SiteGroup)
- `display` field (human-readable name)
- `connected_endpoints` on interfaces
- `_path` on cable-connected components

**Implementation:** Add these to response DTOs via GORM hooks or post-query processing.

### 4.4 Tags Support

Python NetBox uses a `TagSerializerField` that returns `[{id, name, slug, color}]`. The Go models have tag junction tables but the serialization format needs to match.

---

## Phase 5: FilterSet Compatibility

Python NetBox uses django-filter FilterSets with rich query param support:

### 5.1 Standard Filters (all models)

| Filter | Example | Description |
|--------|---------|-------------|
| `q` | `?q=router` | Full-text search |
| `id` | `?id=1&id=2` | Filter by IDs |
| `id__in` | `?id__in=1,2,3` | Filter by ID list |
| `id__gte` | `?id__gte=100` | Greater than or equal |
| `created__gte` | `?created__gte=2024-01-01` | Date filter |
| `last_updated__gte` | — | Date filter |
| `tag` | `?tag=production` | Filter by tag |
| `limit` | `?limit=100` | Result limit |
| `offset` | `?offset=50` | Result offset |

### 5.2 Model-Specific Filters

Each model has 5-30 specific filter fields. Examples:

**Site:** `region`, `region_id`, `group`, `group_id`, `status`, `facility`, `asn`, `latitude`, `longitude`, `contact`, `contact_email`, `tenant`, `tenant_id`, `tag`

**Device:** `region`, `region_id`, `site`, `site_id`, `location`, `location_id`, `rack_id`, `rack`, `device_type`, `device_type_id`, `role`, `role_id`, `manufacturer`, `manufacturer_id`, `platform`, `platform_id`, `status`, `mac_address`, `serial`, `has_primary_ip`, `virtual_chassis_id`, `position`, `asset_tag`, `tenant`, `tenant_id`, `tag`

**Prefix:** `family`, `prefix`, `within`, `within_include`, `contains`, `mask_length`, `mask_length__gte`, `mask_length__lte`, `vrf`, `vrf_id`, `region`, `region_id`, `site`, `site_id`, `vlan`, `vlan_id`, `status`, `role`, `role_id`, `tag`, `rir`, `rir_id`, `aggregate`

**Implementation:** Extend the DAO query-building to accept these filter params from query strings. Use a generic filter-building middleware that maps `field__lookup` patterns to GORM queries.

---

## Phase 6: Remaining Models CRUD Gap

Some Python models have routes but the Go backend may not expose them via HTTP yet. Verify these have working HTTP routes:

### Models to Verify/Add HTTP Routes

| Module | Model | Python Route | Status |
|--------|-------|-------------|--------|
| core | ObjectType | `/api/core/object-types/` | ⚠️ Check |
| core | DataSource sync | `/api/core/data-sources/{id}/sync/` | ❌ Missing |
| extras | CustomField choices | implicit in serializer | ❌ Missing |
| users | ObjectPermission | `/api/users/object-permissions/` | ⚠️ Check |
| dcim | ConnectedDevice | `/api/dcim/connected-device/` | ❌ Missing |

---

## Implementation Order & Priority

### Sprint 1: Unblock Frontend (Critical Path)
1. ✅ ~~Fix AutoMigrate crash~~ (done)
2. **Trailing slash middleware** (1 day)
3. **DRF response envelope wrapper** (1 day)
4. **Page-based pagination handler** (1 day)
5. **PATCH semantics** (0.5 day)
6. **Token auth bridge** (1 day)

### Sprint 2: Bulk Operations & System Endpoints
7. **Bulk create/update/delete** (2 days)
8. **API root + status endpoint** (0.5 day)
9. **Per-module root views** (0.5 day)

### Sprint 3: Custom Actions (DCIM)
10. **Cable trace endpoint** (2 days)
11. **Cable paths endpoint** (1 day)
12. **Rack elevation endpoint** (2 days)
13. **Connected device locator** (0.5 day)
14. **Render config** (1 day)

### Sprint 4: Custom Actions (IPAM)
15. **Available IPs from prefix** (1 day)
16. **Available prefixes from prefix** (1 day)
17. **Available IPs from IP range** (1 day)
18. **Available ASNs from ASN range** (1 day)

### Sprint 5: Serializer & Filter Compatibility
19. **Brief mode serializer** (1 day)
20. **Computed count fields** (1 day)
21. **Display field generation** (0.5 day)
22. **Generic FilterSet builder** (2 days)
23. **Tags serializer format** (0.5 day)

### Sprint 6: Core/Extras Actions
24. **DataSource sync** (1 day)
25. **Job status polling** (0.5 day)
26. **Custom field choices** (1 day)
27. **Search endpoint** (2 days)

### Sprint 7: Polish & Testing
28. **OpenAPI schema generation** (1 day)
29. **Integration tests for all endpoints** (3 days)
30. **Frontend E2E testing** (2 days)

**Total estimated effort: ~30 development days**

---

## Testing Strategy

### Unit Tests (Go)
- Each custom handler tested independently
- DAO layer tests for filter combinations
- Middleware tests for envelope wrapping and trailing slashes

### Integration Tests (Go + PostgreSQL)
- API contract tests that match Python NetBox OpenAPI schema
- Compare response shapes between Python and Go for each endpoint
- Bulk operation tests (create 100 items, update 50, delete 25)

### Frontend E2E Tests
- Playwright/Vitest tests against running Go backend
- Test every page in the model registry (126 models)
- Test special pages (login, dashboard, search, reports, scripts)

### Contract Testing
- Generate API schema from Python NetBox `openapi.json`
- Validate Go backend responses against that schema
- Ensure field names, types, and nesting match

---

## File Structure (New Files)

```
netbox-backend/internal/
├── routers/
│   ├── middleware.go          # Trailing slash, CORS, envelope wrapper
│   ├── auth.go                # Token auth middleware + login/logout
│   ├── system.go              # /api/status/, /api/ root
│   ├── bulk_operations.go     # Generic bulk create/update/delete
│   └── compatibility.go       # DRF-compatible list handler
├── handler/
│   ├── dcim_custom.go         # trace, paths, elevation, connected-device
│   ├── ipam_custom.go         # available-ips, available-prefixes, available-asns
│   ├── extras_custom.go       # render-config, custom-field-choices
│   ├── core_custom.go         # data-source sync
│   └── system.go              # status, API root
├── service/
│   ├── cable_trace.go         # Cable path tracing algorithm
│   ├── rack_elevation.go      # Rack elevation calculation
│   ├── ipam_available.go      # IP/prefix/ASN allocation
│   ├── config_render.go       # Jinja-like template rendering
│   └── filterset.go           # Generic filter builder
└── dao/
    └── filterset.go           # GORM query builder from filter params
```

---

## API Compatibility Checklist

Use this checklist to track parity for each module:

- [ ] **Authentication**: Token login/logout/me endpoints work
- [ ] **List**: `GET /api/{module}/{model}/?page=1&page_size=50` returns DRF envelope
- [ ] **Detail**: `GET /api/{module}/{model}/{id}/` returns full object with nested relations
- [ ] **Create**: `POST /api/{module}/{model}/` creates single object
- [ ] **Bulk Create**: `POST /api/{module}/{model}/` with array creates multiple
- [ ] **Update**: `PATCH /api/{module}/{model}/{id}/` partial update
- [ ] **Bulk Update**: `PATCH /api/{module}/{model}/` with `{pk: [...], ...fields}`
- [ ] **Delete**: `DELETE /api/{module}/{model}/{id}/` deletes single
- [ ] **Bulk Delete**: `DELETE /api/{module}/{model}/` with `{pk: [...]}`
- [ ] **Filter**: Query params (`?region_id=1&status=active`) filter correctly
- [ ] **Search**: `?q=keyword` performs text search
- [ ] **Brief**: `?brief=true` returns compact serializer
- [ ] **Tags**: Tags returned as `[{id, name, slug, color}]`
- [ ] **Custom Actions**: Model-specific actions (trace, elevation, etc.) work
- [ ] **Count Fields**: Computed `*_count` fields included where expected

---

*This plan is based on analysis of:*
- *Python NetBox API (all `api/urls.py` + `api/views.py` files)*
- *Go backend (131 model services, routers, handlers)*
- *Frontend model registry (126 models across 10 modules)*
- *Existing API gap analysis in `docs/apis/API_GAPS.md`*
