# NetBox Frontend Analysis — Pages & Components Inventory

> **Historical planning artifact — superseded.** This inventory and its counts describe an earlier snapshot and are not current implementation status. Use the canonical [status](STATUS.md), [roadmap](ROADMAP.md), [architecture](ARCHITECTURE.md), and [compatibility scope](COMPATIBILITY.md).

> **Purpose:** Complete inventory of pages, components, and features needed to replicate the NetBox UI in Vue.js. Derived from analysis of the original Python/Django templates, URL routing, and navigation structure.

---

## Table of Contents

1. [Current Vue.js Frontend State](#1-current-vuejs-frontend-state)
2. [Architecture: Generic View System](#2-architecture-generic-view-system)
3. [Page Count Summary](#3-page-count-summary)
4. [Per-Module Page Breakdown](#4-per-module-page-breakdown)
5. [Navigation Structure](#5-navigation-structure)
6. [Reusable Component Inventory](#6-reusable-component-inventory)
7. [Specialized Non-CRUD Pages](#7-specialized-non-crud-pages)
8. [Bulk Operations](#8-bulk-operations)
9. [Form Field Types](#9-form-field-types)
10. [Development Estimate](#10-development-estimate)

---

## 1. Current Vue.js Frontend State

### Tech Stack
| Component | Technology |
|---|---|
| Framework | Vue.js 3.5 (Composition API) |
| Router | Vue Router 4.6 |
| State | Pinia 3.0 |
| HTTP | Axios 1.18 |
| Icons | Lucide Icons |
| Utilities | VueUse 14.3 |
| Build | Vite 8.1 |
| Styling | TailwindCSS 3.4 + Forms + Typography |
| TypeScript | 6.0 |

### Current Files (Starter Scaffold)
```
netbox-frontend/src/
├── main.ts              # App bootstrap (Pinia + Router)
├── App.vue              # Root component (<router-view>)
├── style.css            # Global styles (Tailwind imports)
├── api/                 # API client layer
├── assets/              # Static assets
├── components/          # Reusable components
├── layouts/             # Layout wrappers
├── pages/               # Route page components
├── router/              # Vue Router configuration
└── stores/              # Pinia stores
```

The frontend is a minimal scaffold — needs to be built from scratch.

---

## 2. Architecture: Generic View System

**Critical insight:** NetBox does NOT write per-model templates. It uses a **generic class-based view system** with 12 reusable templates that serve ALL models.

### Generic Templates (the core reusable views)

| Template | Purpose | Vue.js Equivalent |
|---|---|---|
| `object_list.html` | **ALL list views** — every model uses this | `<ObjectListView>` |
| `object.html` | Base detail view | `<ObjectDetailView>` |
| `object_edit.html` | **ALL create/edit forms** | `<ObjectEditView>` |
| `object_delete.html` | Single-object delete confirmation | `<ObjectDeleteView>` |
| `bulk_edit.html` | **ALL bulk edit** operations | `<BulkEditView>` |
| `bulk_delete.html` | **ALL bulk delete** operations | `<BulkDeleteView>` |
| `bulk_import.html`` | **ALL CSV/import** operations | `<BulkImportView>` |
| `bulk_rename.html` | Bulk rename | `<BulkRenameView>` |
| `bulk_add_component.html` | Bulk-add child components | `<BulkAddComponentView>` |
| `object_children.html` | Object children display | `<ObjectChildrenView>` |
| `bulletin.html` | Bulletin/info display | `<BulletinView>` |
| `custom_field.html` | Custom field rendering | `<CustomFieldRenderer>` |

### Standard URL Paths Per Model

Each registered model gets these view paths automatically:

| View | URL Pattern | Description |
|---|---|---|
| List | `/{module}/{model}/` | Paginated table view with filters |
| Detail | `/{module}/{model}/{id}/` | Tabbed detail page |
| Add | `/{module}/{model}/add/` | Create form |
| Edit | `/{module}/{model}/{id}/edit/`` | Edit form |
| Delete | `/{module}/{model}/{id}/delete/` | Delete confirmation |
| Bulk Edit | `/{module}/{model}/edit/` | Bulk edit selected |
| Bulk Delete | `/{module}/{model}/delete/` | Bulk delete selected |
| Bulk Import | `/{module}/{model}/import/` | CSV import |
| **+ Feature Views** | (auto-registered based on mixins) | See below |

### Auto-Registered Feature Views (per model, based on mixins)

| Feature | URL Suffix | Condition |
|---|---|---|
| Contacts | `/contacts/` | Model uses `ContactsMixin` |
| Journal | `/journal/` | Model uses `JournalingMixin` |
| Change Log | `/changelog/` | Model uses `ChangeLoggingMixin` |
| Jobs | `/jobs/` | Model uses `JobsMixin` |
| Image Attachments | `/image-attachments/` | Model uses `ImageAttachmentsMixin` |
| Sync | `/sync/` | Model uses `SyncedDataMixin` |

---

## 3. Page Count Summary

### Total Templates in Python NetBox: 299 `.html` files

| Category | Count | Notes |
|---|---|---|
| Generic reusable templates | 12 | Serve ALL models |
| Per-model custom detail templates | ~35 | Extended from generic `object.html` |
| Specialized pages | ~20 | Cable trace, rack elevation, dashboard, etc. |
| Shared/inc partial templates | ~40 | Table rows, filter forms, nav, modals |
| Error/auth pages | 7 | 403, 404, 500, login, search, home, media_failure |
| Module-specific bulk templates | ~15 | Custom bulk operations |

### Vue.js Components Needed

| Component Type | Estimated Count | Based On |
|---|---|---|
| **Generic view components** | 12 | One per generic template type |
| **Shared/layout components** | ~25 | Sidebar, header, search, table, filter, pagination, modals, tabs |
| **Per-model detail overrides** | ~35 | Models with custom detail templates |
| **Specialized page components** | ~20 | Cable trace, rack elevation, dashboard, reports |
| **Form field components** | ~15 | Dynamic select, API select, tag input, custom fields |
| **Navigation components** | ~10 | Menu groups, menu items, breadcrumbs |
| **Utility components** | ~10 | Status badges, type badges, JSON viewer, code blocks |
| **Total** | **~127 components** | |

### Vue.js Pages (Route Components) Needed

| Page Type | Count | Notes |
|---|---|---|
| Login/Home/Error pages | 5 | Login, Dashboard, 403, 404, 500 |
| List pages (generic) | 1 | `<ObjectListView>` reused for all models |
| Detail pages (generic) | 1 | `<ObjectDetailView>` reused for all models |
| Detail pages (custom) | ~35 | Per-model custom detail templates |
| Add/Edit pages (generic) | 1 | `<ObjectEditView>` reused for all models |
| Bulk operation pages (generic) | 4 | Bulk edit, delete, import, rename |
| Specialized pages | ~20 | Cable trace, rack elevation, etc. |
| **Total route entries** | **~250+** | Generated from model registry |

---

## 4. Per-Module Page Breakdown

### DCIM (46 models) — Largest module

| Model | List | Detail | Edit | Delete | Bulk Edit | Bulk Delete | Bulk Import | Special |
|---|---|---|---|---|---|---|---|---|
| Site | ✓ | ✓ (custom) | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Region | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| SiteGroup | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Location | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Rack | ✓ | ✓ (custom: elevation) | ✓ | ✓ | ✓ | ✓ | ✓ | **Rack Elevation** |
| RackReservation | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| RackRole | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| RackType | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Manufacturer | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| DeviceType | ✓ | ✓ (custom) | ✓ | ✓ | ✓ | ✓ | ✓ | |
| DeviceRole | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Platform | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Device | ✓ | ✓ (custom) | ✓ | ✓ | ✓ | ✓ | ✓ | **Config Context, LLDP neighbors** |
| VirtualChassis | ✓ | ✓ (custom) | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Module | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| ModuleType | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| ModuleBay | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Bulk Add** |
| ConsolePort | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect** |
| ConsoleServerPort | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect** |
| PowerPort | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect** |
| PowerOutlet | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect** |
| Interface | ✓ | ✓ (custom) | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect, VLANs** |
| FrontPort | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect** |
| RearPort | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Trace, Connect** |
| DeviceBay | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **Bulk Add** |
| InventoryItem | ✓ | ✓ (custom: tree) | ✓ | ✓ | ✓ | ✓ | ✓ | |
| Cable | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| PowerPanel | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| PowerFeed | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | |
| + 14 more models | ✓ each | ✓ each | ✓ each | ✓ each | ✓ each | ✓ each | ✓ each | |

**DCIM URL patterns: ~300+** (46 models × ~7 views each + specialized views)

### IPAM (19 models)

| Model | Special Features |
|---|---|
| Prefix | **Hierarchy tree view**, scope, IP addresses nested |
| IPAddress | **Assignment to interface**, primary IP, bulk assign |
| VLAN | **VLAN group scope**, available VLANs view |
| VLANGroup | **VLAN availability**, scope |
| VRF | Route targets M2M |
| Aggregate | RIR grouping |
| ASN/ASNRange | Range validation |
| FHRPGroup | Protocol assignments |

**IPAM URL patterns: ~140+**

### Circuits (6 models)

| Model | Special Features |
|---|---|
| Circuit | Custom detail with terminations |
| CircuitTermination | **Cable connectivity, trace** |
| Provider | Custom detail |
| ProviderNetwork | |
| CircuitType | |
| Account | |

**Circuits URL patterns: ~45+**

### Virtualization (5 models)

| Model | Special Features |
|---|---|
| VirtualMachine | Custom detail, resources, primary IPs |
| VMInterface | **Cable connectivity**, VLAN assignment |
| Cluster | Custom detail with VMs |
| ClusterType/Group | |

**Virtualization URL patterns: ~40+**

### VPN (5 models)

| Model | Special Features |
|---|---|
| Tunnel | IPSec profile |
| TunnelTermination | Interface assignment |
| L2VPN | Terminations M2M |
| IPSecProfile | Complex crypto fields |

**VPN URL patterns: ~35+**

### Wireless (3 models)

| Model | Special Features |
|---|---|
| WirelessLAN | Scoped, group hierarchy |
| WirelessLink | Interface-to-interface connection |

**Wireless URL patterns: ~25+**

### Extras (8 models)

| Model | Special Features |
|---|---|
| ConfigContext | **Multi-object assignment** (regions, sites, tags, etc.) |
| ConfigTemplate | Template rendering preview |
| CustomField | **Type-specific rendering**, choice sets |
| Tag | Color, tagged items |
| Webhook | Content type selection, payload format |
| JournalEntry | Per-object journal |
| ImageAttachment | File upload |

**Extras URL patterns: ~60+**

### Tenancy (6 models)

Standard CRUD for Tenant, TenantGroup, Contact, ContactGroup, ContactRole, ContactAssignment.

**Tenancy URL patterns: ~45+**

### Users (3 models)

User, Group, Token management.

**Users URL patterns: ~25+**

### Core (5 models)

DataSource, DataFile, Job, ObjectChange, ObjectType.

**Core URL patterns: ~35+**

### Root Pages

| Page | URL |
|---|---|
| Dashboard/Home | `/` |
| Login | `/login/` |
| Search | `/search/` |
| 403 Forbidden | `/403/` |
| 404 Not Found | `/404/` |
| 500 Server Error | `/500/` |

---

## 5. Navigation Structure

### Sidebar Menu (Bootstrap navbar-vertical, collapsible)

The sidebar contains grouped menu items. Based on analysis:

```
📦 Organization
  ├── Sites
  ├── Regions
  ├── Site Groups
  ├── Locations
  ├── Tenants
  ├── Tenant Groups
  ├── Contacts
  ├── Contact Groups
  └── Contact Roles

📦 DCIM
  ├── Devices
  ├── Device Types
  ├── Manufacturers
  ├── Device Roles
  ├── Platforms
  ├── Virtual Chassis
  ├── Modules
  ├── Module Types
  ├── Module Bays
  ├── Console Ports
  ├── Console Server Ports
  ├── Power Ports
  ├── Power Outlets
  ├── Interfaces
  ├── Front Ports
  ├── Rear Ports
  ├── Device Bays
  ├── Inventory Items
  ├── Cables
  ├── Racks
  ├── Rack Reservations
  ├── Rack Roles
  ├── Rack Types
  ├── Power Panels
  ├── Power Feeds
  └── Virtual Device Contexts

📦 IPAM
  ├── VRFs
  ├── Route Targets
  ├── RIRs
  ├── Aggregates
  ├── Prefixes
  ├── IP Ranges
  ├── IP Addresses
  ├── VLANs
  ├── VLAN Groups
  ├── ASN Ranges
  ├── ASNs
  ├── FHRP Groups
  ├── Services
  └── Service Templates

📦 Circuits
  ├── Circuits
  ├── Circuit Types
  ├── Providers
  ├── Provider Networks
  └── Accounts

📦 Virtualization
  ├── Virtual Machines
  ├── VM Interfaces
  ├── Clusters
  ├── Cluster Types
  └── Cluster Groups

📦 VPN
  ├── Tunnels
  ├── Tunnel Terminations
  ├── IPSec Profiles
  ├── L2VPNs
  └── L2VPN Terminations

📦 Wireless
  ├── Wireless LANs
  ├── Wireless LAN Groups
  └── Wireless Links

📦 Other
  ├── Tags
  ├── Config Contexts
  ├── Config Templates
  ├── Custom Fields
  ├── Custom Field Choice Sets
  ├── Webhooks
  ├── Journal Entries

📦 Admin
  ├── Users
  ├── Groups
  ├── Tokens
  ├── Data Sources
  ├── Jobs
  ├── Change Log
  └── Object Types
```

### Top Bar
- Global search form (GET `/search?q=`)
- User menu (profile, admin, logout)
- Light/dark mode toggle

---

## 6. Reusable Component Inventory

### Layout Components (8)

| Component | Purpose |
|---|---|
| `AppLayout.vue` | Main layout: sidebar + header + content |
| `SidebarNav.vue` | Collapsible vertical navigation menu |
| `SidebarMenuGroup.vue` | Expandable menu group (e.g., "DCIM") |
| `SidebarMenuItem.vue` | Individual menu link |
| `TopBar.vue` | Top bar: search + user menu |
| `SearchBar.vue` | Global search input |
| `UserMenu.vue` | User dropdown (profile, logout) |
| `PageHeader.vue` | Page title + breadcrumb + action buttons |

### Table Components (6)

| Component | Purpose |
|---|---|
| `DataTable.vue` | Generic data table with sorting |
| `TableColumn.vue` | Column definition wrapper |
| `TableRow.vue` | Row rendering with selection |
| `BulkSelectCheckbox.vue` | Select-all / select-row checkboxes |
| `TablePagination.vue` | Page number navigation |
| `ExportButton.vue` | CSV/JSON export trigger |

### Filter Components (5)

| Component | Purpose |
|---|---|
| `FilterPanel.vue` | Slide-out filter form |
| `FilterField.vue` | Individual filter input |
| `DynamicSelectFilter.vue` | API-backed multi-select filter |
| `BooleanFilter.vue` | True/false/null tri-state |
| `DateRangeFilter.vue` | Date range picker |

### Form Components (10)

| Component | Purpose |
|---|---|
| `DynamicForm.vue` | Renders form fields from schema |
| `DynamicField.vue` | Routes to correct field type |
| `ApiSelectField.vue` | API-backed select (searchable dropdown) |
| `TagInputField.vue` | Tag creation/selection |
| `CustomFieldRenderer.vue` | Custom fields (text, int, bool, date, select, JSON, URL) |
| `StatusSelect.vue` | Status badge dropdown |
| `CsvImportField.vue` | CSV paste or file upload |
| `SlugField.vue` | Auto-slug from name |
| `MarkdownEditor.vue` | Markdown textarea with preview |
| `ImageUpload.vue` | Image attachment upload |

### Detail Components (8)

| Component | Purpose |
|---|---|
| `ObjectDetailLayout.vue` | Tabbed detail page container |
| `DetailTab.vue` | Individual tab panel |
| `PropertiesTable.vue` | Key-value attribute display |
| `StatusBadge.vue` | Colored status pill |
| `TypeBadge.vue` | Model type badge |
| `TagsDisplay.vue` | Tag list display |
| `ContactsPanel.vue` | Contact assignments panel |
| `JournalPanel.vue` | Journal entries panel |

### Modal Components (4)

| Component | Purpose |
|---|---|
| `ConfirmModal.vue` | Generic confirmation dialog |
| `BulkActionModal.vue` | Bulk operation confirmation |
| `ErrorModal.vue` | Error detail display |
| `JsonViewer.vue` | JSON tree viewer |

### Utility Components (5)

| Component | Purpose |
|---|---|
| `BreadcrumbNav.vue` | Hierarchical breadcrumbs |
| `LoadingSpinner.vue` | Loading indicator |
| `EmptyState.vue` | "No results found" display |
| `CopyToClipboard.vue` | Click-to-copy utility |
| `PaginationControls.vue` | Prev/next + page size |

**Total reusable components: ~46**

---

## 7. Specialized Non-CRUD Pages

These are the most complex UI components requiring custom implementation:

### Cable Trace Visualization (1 page, 1 component)
- **Page:** Cable trace from any termination endpoint
- **Visualization:** SVG rendering of cable path (near_ends → cable → far_ends)
- **Features:** Split path detection, total length calculation (meters/feet), multi-path selection
- **Vue.js:** `<CableTraceView>` with SVG rendering or D3.js graph

### Rack Elevation Diagram (1 page, 1 component)
- **Page:** Visual rack unit display (front + rear)
- **Visualization:** Vertical grid of rack units with device placement
- **Features:** Color coding by device role, reservation highlighting, u1/u2 formatting
- **Vue.js:** `<RackElevationView>` with CSS grid or SVG

### Dashboard / Home (1 page)
- **Widgets:** Configurable dashboard panels (object counts, recent changes, etc.)
- **Vue.js:** `<DashboardView>` with draggable widgets

### Global Search (1 page)
- **Features:** Full-text search across all models, categorized results
- **Vue.js:** `<SearchResultsView>`

### Config Context Rendering (1 page)
- **Features:** Render Jinja2/template against object data
- **Vue.js:** `<ConfigContextRender>` (requires API-side rendering)

### Reports & Custom Scripts (2 pages)
- **Page:** Report list + execution + results
- **Page:** Custom script list + execution form
- **Vue.js:** `<ReportsView>`, `<ScriptsView>`

### Data Source Sync (1 page)
- **Features:** Trigger sync, view sync status, browse synced files
- **Vue.js:** `<DataSourceSyncView>`

### Background Jobs Queue (1 page)
- **Features:** Job list with status, log entries, retry/cancel
- **Vue.js:** `<JobsView>`

### Change Log (1 page)
- **Features:** Object change history with diff view (prechange vs postchange)
- **Vue.js:** `<ChangeLogView>` with diff display

### GraphQL Explorer (1 page)
- **Features:** Interactive GraphQL query editor
- **Vue.js:** Embed GraphiQL or custom query builder

### API Browser (1 page)
- **Features:** Swagger/OpenAPI interactive docs
- **Vue.js:** Embed Swagger UI or custom API browser

### Inventory Item Tree (1 component)
- **Features:** Hierarchical tree view of inventory items
- **Vue.js:** `<InventoryItemTree>` recursive component

### Prefix Hierarchy Tree (1 component)
- **Features:** Collapsible tree of nested prefixes
- **Vue.js:** `<PrefixTreeView>` recursive component

**Total specialized pages: ~13 pages, ~15 components**

---

## 8. Bulk Operations

NetBox has powerful bulk operations that all use generic templates:

| Operation | Description | Vue.js Component |
|---|---|---|
| **Bulk Edit** | Select multiple objects → apply field changes to all | `<BulkEditView>` |
| **Bulk Delete** | Select multiple → confirm deletion | `<BulkDeleteView>` |
| **Bulk Import** | CSV paste or file upload → create objects | `<BulkImportView>` |
| **Bulk Rename** | Find-and-replace pattern on selected objects | `<BulkRenameView>` |
| **Bulk Add Components** | Add child components to multiple parent devices | `<BulkAddComponentView>` |
| **Bulk Assign** | Assign VLANs/tags to multiple objects | Reuses BulkEdit |

### CSV Import Format
Each model has a `to_csv()` method and CSV import form with:
- Header row auto-detection
- Field mapping
- Foreign key resolution by name/ID
- Validation errors per row

---

## 9. Form Field Types

NetBox forms use these field types that need Vue.js equivalents:

| Field Type | Django Widget | Vue.js Component |
|---|---|---|
| CharField | text input | `<input type="text">` |
| SlugField | text input (auto-slug) | `<SlugField>` |
| IntegerField | number input | `<input type="number">` |
| BooleanField | checkbox/tristate | `<BooleanFilter>` |
| ChoiceField | select dropdown | `<select>` |
| ModelChoiceField | **API-backed select** | `<ApiSelectField>` (searchable) |
| ModelMultipleChoiceField | **API-backed multi-select** | `<ApiSelectField :multiple="true">` |
| TagField | tag input | `<TagInputField>` |
| DatePicker | date input | `<input type="date">` |
| DateTimePicker | datetime input | `<input type="datetime-local">` |
| JSONField | JSON editor | `<JsonEditor>` |
| MarkdownField | markdown editor | `<MarkdownEditor>` |
| CSVDataField | CSV text area | `<CsvImportField>` |
| DynamicModelChoiceField | API select with add button | `<ApiSelectField :creatable="true">` |
| ContentTypeChoiceField | content type select | `<ContentTypeSelect>` |

---

## 10. Development Estimate

### Summary

| Metric | Count |
|---|---|
| **Total models to serve** | 104 |
| **Standard views per model** | 7 (list, detail, add, edit, delete, bulk_edit, bulk_delete) |
| **Total standard route entries** | ~728 (104 × 7) |
| **Bulk import routes** | ~80 |
| **Specialized routes** | ~40 |
| **Feature tab routes** (contacts, journal, changelog) | ~300 |
| **Total route entries** | **~1,150** |
| **Generic Vue components needed** | 12 |
| **Shared/layout components** | 46 |
| **Specialized page components** | 15 |
| **Per-model detail overrides** | ~35 |
| **Total Vue components** | **~108** |

### Development Phases

#### Phase 1: Foundation (Weeks 1-3)
- Layout components (sidebar, header, search)
- Generic ObjectListView (data table, filters, pagination, bulk select)
- Generic ObjectDetailView (tabbed layout, properties table)
- Generic ObjectEditView (dynamic form rendering)
- Router configuration with auto-generation from model registry
- Pinia stores (auth, UI state, notifications)

#### Phase 2: Core Models (Weeks 4-7)
- DCIM: Site, Region, Location, Rack, Device, DeviceType, Manufacturer
- IPAM: VRF, Prefix, IPAddress, VLAN, VLANGroup, Aggregate
- Tenancy: Tenant, Contact
- Circuits: Circuit, Provider
- Basic CRUD for all above

#### Phase 3: Components & Connections (Weeks 8-10)
- DCIM: All device components (Interface, ConsolePort, PowerPort, etc.)
- Cable connection forms
- Cable trace visualization
- Rack elevation diagrams
- Bulk operations (edit, delete, import, rename, add components)

#### Phase 4: Remaining Models (Weeks 11-14)
- All remaining DCIM, IPAM, Circuits, Virtualization, VPN, Wireless models
- Extras (ConfigContext, CustomField, Tag, Webhook)
- Users & Core (User, Group, Token, Jobs, ChangeLog)

#### Phase 5: Specialized Features (Weeks 15-18)
- Dashboard with widgets
- Global search
- Reports & scripts
- Data source sync
- GraphQL explorer
- API browser
- Change log diff viewer
- Prefix hierarchy tree
- Inventory item tree

---

## Appendix: Auto-Generated Route Strategy

Since NetBox uses a registry-based approach, the Vue.js frontend should also auto-generate routes:

```typescript
// src/router/auto-routes.ts
const MODEL_REGISTRY = [
  { module: 'dcim', model: 'site', name: 'Site', ... },
  { module: 'dcim', model: 'device', name: 'Device', ... },
  // ... 104 models
];

export function generateRoutes(): RouteRecordRaw[] {
  return MODEL_REGISTRY.flatMap(m => [
    { path: `/${m.module}/${m.model}/`, component: ObjectListView, meta: { model: m } },
    { path: `/${m.module}/${m.model}/add/`, component: ObjectEditView, meta: { model: m } },
    { path: `/${m.module}/${m.model}/:id/`, component: ObjectDetailView, meta: { model: m } },
    { path: `/${m.module}/${m.model}/:id/edit/`, component: ObjectEditView, meta: { model: m } },
    { path: `/${m.module}/${m.model}/:id/delete/`, component: ObjectDeleteView, meta: { model: m } },
    { path: `/${m.module}/${m.model}/edit/`, component: BulkEditView, meta: { model: m } },
    { path: `/${m.module}/${m.model}/delete/`, component: BulkDeleteView, meta: { model: m } },
  ]);
}
```

This means **~1,150 routes are generated from just 7 generic components + 104 model definitions**.
