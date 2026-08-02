# NetBox Go Frontend — Completion Plan

> **Historical planning artifact — superseded.** The completeness percentage, counts, findings, and tasks below describe an earlier snapshot and are not current. Use the canonical [status](STATUS.md), [roadmap](ROADMAP.md), [architecture](ARCHITECTURE.md), and [compatibility scope](COMPATIBILITY.md).

> **Purpose:** Detailed, prioritized plan to take the Vue.js frontend from ~92% to production-ready.
> Derived from a careful review of the original Python NetBox v4.4.0 implementation and the current
> Go+Vue.js codebase state.

**Generated:** 2026-06-27
**Current completeness:** ~92% (100/100 models registered, 90 components, 13 specialized pages, 41 tests)

---

## Executive Summary

The frontend is functionally complete — all 100 models render via generic views, navigation works,
filters/bulk-ops/CSV import-export work, specialized pages (dashboard, rack elevation, prefix tree,
cable trace, VLAN availability, search, jobs, data sources, reports, scripts, API docs, GraphQL) all
exist. The remaining work is **hardening + completeness**: replacing fallback form fields with proper
typed inputs, adding the 3 missing bulk views, expanding test coverage, and setting up lint tooling.

---

## Current Inventory (verified)

### Model Registry — `src/router/models/*.ts` (100 models)
| Module | Count | File |
|---|---|---|
| DCIM | 41 | `dcim.ts` |
| IPAM | 18 | `ipam.ts` |
| Tenancy | 6 | `tenancy.ts` |
| Circuits | 6 | `circuits.ts` |
| Virtualization | 5 | `virtualization.ts` |
| VPN | 5 | `vpn.ts` |
| Wireless | 3 | `wireless.ts` |
| Extras | 8 | `extras.ts` |
| Users | 3 | `users.ts` |
| Core | 5 | `core.ts` |

### Components (90 total)
- **`components/form/`** (9): TextField, SlugField, SelectField, BooleanField, TextareaField, ApiSelectField, TagInputField, DynamicField, DynamicForm
- **`components/filter/`** (8): FilterPanel, FilterField, TextFilter, SelectFilter, ApiSelectFilter, BooleanFilter, DateRangeFilter, IntegerRangeFilter
- **`components/detail/`** (16): PropertiesTable, DetailTabBar, StatusBadge, TagsDisplay, ContactsPanel, JournalPanel, ChangeLogPanel, ImageAttachmentsPanel, AdvancedTab, CustomFieldsPanel, ComponentListTab, InventoryItemTree, InventoryNode
- **`components/shared/`** (many): LoadingSpinner, EmptyState, Modal, ConfirmModal, ToastContainer, StatusBadge, PaginationControls, CodeBlock, ConfirmButton, CopyToClipboard
- **`components/table/`** (3): DataTable, BulkSelectBar, ExportButton
- **`components/layout/`** (8): SidebarNav, SidebarMenuGroup, SidebarMenuItem, TopBar, SearchBar, UserMenu, ThemeToggle, PageHeader
- **`components/dashboard/`** (3): ObjectCountWidget, RecentChangesWidget, ChartWidget
- **`components/modal/`**: Modal.vue

### Pages (23 total)
- **`pages/generic/`** (6): ObjectListView, ObjectDetailView, ObjectEditView, BulkEditView, BulkDeleteView, BulkImportView
- **`pages/special/`** (13): DashboardView, SearchResultsView, RackElevationView, PrefixTreeView, PrefixNode, CableTraceView, VLANAvailabilityGrid, JobsView, DataSourceSyncView, ReportsView, ScriptsView, ApiDocsView, GraphiQLView
- **`pages/auth/`** (2): LoginPage, LogoutAction
- **`pages/errors/`** (3): NotFoundView, ForbiddenView, ServerErrorView

### Tests (3 files, 41 cases)
- `utils/csv.test.ts` — CSV parse/export
- `stores/cache.test.ts` — cache store
- `utils/contentType.test.ts` — content-type builder


---

## GAP ANALYSIS vs. `FRONTEND_PLAN.md`

### Gap 1: Form Field Components (HIGH — blocks proper data entry for all 100 models)

The `DynamicField.vue` currently routes 7 field types to **fallback** components:
```ts
number:   TextField     // ← plain text, no min/max/step validation
markdown: TextareaField // ← no preview, no markdown rendering
json:     TextareaField // ← no syntax validation, no pretty-print
date:     TextField     // ← no date picker
datetime: TextField     // ← no datetime picker
csv:      TextareaField // ← handled by BulkImportView
```

**Missing dedicated components:**
- [x] `NumberField.vue` — numeric input with min/max/step, integer/decimal
- [x] `MarkdownField.vue` — textarea with live preview + markdown rendering
- [x] `JsonField.vue` — textarea with JSON validation + pretty-print toggle
- [x] `DateField.vue` — native date picker (`<input type="date">`)
- [x] `DateTimeField.vue` — native datetime picker (`<input type="datetime-local">`)
- [x] `CustomFieldRenderer.vue` — render custom fields by type (Phase 2.6)
- [x] `CsvImportField.vue` — column mapping + validation (Phase 2.6)

### Gap 2: Missing Bulk Views (MEDIUM — completeness)

- [x] `ObjectDeleteView.vue` — single-object delete confirmation (Phase 2.7). Delete link on detail
  pages currently routes to `/delete/` which returns 404.
- [x] `BulkRenameView.vue` — find-and-replace rename (Phase 2.8)
- [x] `BulkAddComponentView.vue` — pattern-based component creation (Phase 2.8)

### Gap 3: Test Coverage (MEDIUM — quality)

- [ ] Unit tests for composables (`useTable`, `useApi`)
- [ ] Component tests for generic views (ObjectListView, ObjectEditView)
- [ ] Unit tests for new form field components

### Gap 4: Tooling & Quality (LOW — developer experience)

- [ ] ESLint config with `eslint-plugin-vue`
- [ ] Prettier config

### Gap 5: Polish (LOW — UX)

- [ ] Route-level code splitting (bundle > 500kB warning)
- [ ] Loading skeletons for data fetches
- [ ] Keyboard shortcuts (Ctrl+K search, Esc modals)

---

## IMPLEMENTATION PLAN

### Phase A: Form Field Components (Priority 1)

> **Why first:** Every model's create/edit form depends on these. Currently `number` fields accept
> non-numeric input, `date` fields are free-text, and `json`/`markdown` have no validation/preview.

**Deliverables:** 7 new components + updated `DynamicField.vue` routing + tests

#### A1. `NumberField.vue`
- `<input type="number">` with `min`, `max`, `step` props
- Emits `number | null` (not string)
- Validates against min/max
- Used for: VID, U-height, position, MTU, speed, etc.

#### A2. `DateField.vue`
- `<input type="date">` (native browser picker)
- Emits ISO date string `YYYY-MM-DD` or `null`
- Used for: install_date, date_added, etc.

#### A3. `DateTimeField.vue`
- `<input type="datetime-local">` (native picker)
- Emits ISO datetime string or `null`
- Used for: token expires, last_updated, etc.

#### A4. `JsonField.vue`
- Textarea with JSON syntax validation (real-time, debounced)
- Pretty-print toggle button (minify vs 2-space indent)
- Error message on invalid JSON
- Emits parsed object (not string) on valid input
- Used for: config context `data`, webhook `body_template`, custom field JSON

#### A5. `MarkdownField.vue`
- Textarea + "Write/Preview" toggle tabs
- Preview renders markdown via existing `utils/markdown.ts`
- Cheat-sheet popover (common syntax reference)
- Emits raw markdown string
- Used for: comments, journal entries, descriptions

#### A6. `CustomFieldRenderer.vue`
- Accepts a custom field definition + value
- Routes by type: text, longtext (textarea), integer (number), decimal (number step=0.01),
  boolean (checkbox), date (date picker), datetime (datetime picker), url (text type=url),
  json (JsonField), select/multiselect (SelectField), object/objects (ApiSelectField)
- Used by: `CustomFieldsPanel.vue` and `ObjectEditView` advanced section

#### A7. `CsvImportField.vue`
- Textarea for CSV paste OR file drop zone
- Column mapping UI (CSV header to model field key)
- Per-row validation preview
- Emits parsed+validated rows array
- Used by: `BulkImportView.vue`

#### A8. Update `DynamicField.vue`
- Add proper routing for all field types (remove TextField fallbacks)

### Phase B: Missing Views (Priority 2)

#### B1. `ObjectDeleteView.vue`
- Confirmation page for single-object delete
- Shows object name/display, properties summary
- DELETE call to success toast to redirect to list
- Route: `/:module/:model/:id/delete/`

#### B2. `BulkRenameView.vue`
- Find-and-replace UI: "Find" + "Replace with" inputs
- Regex toggle + case-sensitivity toggle
- Live preview table (current name to new name)
- Confirm to PATCH batch with new names

#### B3. `BulkAddComponentView.vue`
- Component type selector (interface, console port, etc.)
- Name pattern input (e.g., `eth[0-3]` generates eth0, eth1, eth2, eth3)
- Target device multi-select
- Confirm to batch POST

### Phase C: Tests (Priority 3)

- [x] `NumberField.test.ts`, `DateField.test.ts`, `JsonField.test.ts`, `MarkdownField.test.ts`
- [x] `useTable.test.ts`, `useApi.test.ts`
- [x] `ObjectListView.test.ts`, `ObjectEditView.test.ts`

### Phase D: Tooling (Priority 4)

- [ ] ESLint + Prettier configs

---

## IMPLEMENTATION ORDER

1. **Phase A** (form fields) — 7 components + DynamicField update — highest impact
2. **Phase B** (missing views) — 3 views + router update
3. **Phase C** (tests) — lock in the new components
4. **Phase D** (tooling) — developer experience

---

## ACCEPTANCE CRITERIA

- [x] `npm run build` passes with zero errors
- [x] `npm test` passes with 68 test cases (up from 41)
- [x] All field types in `FormFieldDef` have dedicated components (no TextField fallbacks)
- [x] Delete button on detail pages works (ObjectDeleteView route)
