# NetBox Go — Frontend Implementation Plan

> **Historical planning artifact — superseded.** The checklist below is retained for context; its boxes and dependency suggestions do not report current implementation status or accepted scope. Use the canonical [status](STATUS.md), [roadmap](ROADMAP.md), [architecture](ARCHITECTURE.md), and [compatibility scope](COMPATIBILITY.md).

> **Purpose:** Living checklist for the Vue.js frontend implementation. Check off items as they are completed. Each phase builds on the previous one.

**Legend:** `- [ ]` = TODO · `- [x]` = Done · `- [~]` = In Progress / Blocked

---

## Phase 0: Project Setup & Infrastructure

- [ ] Configure Vite with path aliases (`@/` → `src/`)
- [ ] Configure TypeScript strict mode in `tsconfig.app.json`
- [ ] Set up TailwindCSS theme to match NetBox color palette
- [ ] Install additional dependencies:
  - [ ] `@vueuse/core` (already installed)
  - [ ] `vue-router` (already installed)
  - [ ] `pinia` (already installed)
  - [ ] `axios` (already installed)
  - [ ] `@lucide/vue` (already installed)
  - [ ] `qs` (query string serialization for API filters)
  - [ ] `chart.js` + `vue-chartjs` (dashboard widgets)
  - [ ] `d3` or `vis-network` (cable trace / topology)
  - [ ] `graphiql` (GraphQL explorer)
  - [ ] `swagger-ui-dist` (API browser)
  - [ ] `jsdiff` or `diff2html` (change log diff)
  - [ ] `papaparse` (CSV import/export)
  - [ ] `marked` or `markdown-it` (markdown rendering)
- [ ] Create `.env` with `VITE_API_BASE_URL`
- [ ] Create `src/config.ts` with API base URL and app settings
- [ ] Set up ESLint + Prettier rules
- [ ] Create folder structure:
  ```
  src/
  ├── api/           # API client modules per module
  ├── components/    # Reusable components
  │   ├── layout/    # Layout components
  │   ├── table/     # Table components
  │   ├── filter/    # Filter components
  │   ├── form/      # Form field components
  │   ├── detail/    # Detail view components
  │   ├── modal/     # Modal/dialog components
  │   └── shared/    # Utility/shared components
  ├── composables/   # Vue composables (hooks)
  ├── layouts/       # Page layouts
  ├── pages/         # Route page components
  │   ├── generic/   # Generic CRUD views
  │   └── special/   # Specialized non-CRUD views
  ├── router/        # Router config + auto-routes
  ├── stores/        # Pinia stores
  ├── types/         # TypeScript types
  └── utils/         # Utility functions
  ```

---

## Phase 1: Foundation — Layout & Core Infrastructure (Weeks 1-3)

### 1.1 API Client Layer

- [ ] `src/api/client.ts` — Axios instance with:
  - [ ] Base URL from config
  - [ ] Request interceptor (auth token injection)
  - [ ] Response interceptor (error normalization)
  - [ ] Pagination header parsing (`X-Total-Count`, `Link`)
- [ ] `src/api/types.ts` — Shared API response types:
  - [ ] `PaginatedResponse<T>`
  - [ ] `ApiResponse<T>`
  - [ ] `ApiError`
  - [ ] `ChoiceOption` (for select fields)
- [ ] `src/api/modules/` — One file per API module:
  - [ ] `dcim.ts` (Site, Device, Rack, Interface, etc.)
  - [ ] `ipam.ts` (VRF, Prefix, IPAddress, VLAN, etc.)
  - [ ] `circuits.ts`
  - [ ] `tenancy.ts`
  - [ ] `virtualization.ts`
  - [ ] `vpn.ts`
  - [ ] `wireless.ts`
  - [ ] `extras.ts`
  - [ ] `users.ts`
  - [ ] `core.ts`
- [ ] `src/composables/useApi.ts` — Generic CRUD composable:
  - [ ] `useList(model, filters)` — paginated list fetching
  - [ ] `useDetail(model, id)` — single object fetch
  - [ ] `useCreate(model, data)` — create
  - [ ] `useUpdate(model, id, data)` — update
  - [ ] `useDelete(model, id)` — delete
  - [ ] `useBulkEdit(model, ids, data)` — bulk edit
  - [ ] `useBulkDelete(model, ids)` — bulk delete
  - [ ] `useBulkImport(model, csvData)` — CSV import

### 1.2 Pinia Stores

- [ ] `src/stores/auth.ts`:
  - [ ] Current user state
  - [ ] Login/logout actions
  - [ ] Token management
  - [ ] Permission checking helpers
- [ ] `src/stores/ui.ts`:
  - [ ] Sidebar collapsed/expanded state
  - [ ] Dark/light mode toggle
  - [ ] Active breadcrumb state
- [ ] `src/stores/notifications.ts`:
  - [ ] Toast notification queue
  - [ ] Success/error/info actions
- [ ] `src/stores/registry.ts`:
  - [ ] Model registry (all 104 models with metadata)
  - [ ] Per-model config: display name, fields, filters, tabs
- [ ] `src/stores/table.ts`:
  - [ ] Per-model selected rows (for bulk operations)
  - [ ] Per-model table preferences (columns, sort, page size)

### 1.3 Layout Components

- [ ] `src/layouts/DefaultLayout.vue` — Main app shell:
  - [ ] Sidebar (left)
  - [ ] Top bar
  - [ ] Content router-view
  - [ ] Toast notification container
- [ ] `src/layouts/AuthLayout.vue` — Login page wrapper (no sidebar)
- [ ] `src/components/layout/SidebarNav.vue`:
  - [ ] Collapsible menu groups
  - [ ] Active route highlighting
  - [ ] Pin/unpin sidebar
  - [ ] Mobile responsive (off-canvas)
- [ ] `src/components/layout/SidebarMenuGroup.vue`:
  - [ ] Expand/collapse animation
  - [ ] Group icon
  - [ ] Child items list
- [ ] `src/components/layout/SidebarMenuItem.vue`:
  - [ ] Router-link navigation
  - [ ] Active state
  - [ ] Badge count (optional)
- [ ] `src/components/layout/TopBar.vue`:
  - [ ] Global search input
  - [ ] User dropdown menu
  - [ ] Dark/light toggle button
  - [ ] Mobile menu toggle button
- [ ] `src/components/layout/SearchBar.vue`:
  - [ ] Debounced search input
  - [ ] Navigates to `/search?q=`
  - [ ] Keyboard shortcut (Ctrl+K)
- [ ] `src/components/layout/UserMenu.vue`:
  - [ ] User avatar + name
  - [ ] Profile link
  - [ ] Admin link
  - [ ] Logout action
- [ ] `src/components/layout/ThemeToggle.vue`:
  - [ ] Toggle dark/light class on `<html>`
  - [ ] Persist preference in localStorage

### 1.4 Page Header Component

- [ ] `src/components/layout/PageHeader.vue`:
  - [ ] Page title (from route meta)
  - [ ] Breadcrumb navigation
  - [ ] Action buttons (Add, Import, Export)
  - [ ] Tab bar (for detail pages)

### 1.5 Navigation Menu Configuration

- [ ] `src/config/navigation.ts` — Full sidebar menu definition:
  - [ ] Organization group (Sites, Regions, Site Groups, Locations, Tenants, Contacts)
  - [ ] DCIM group (26+ items)
  - [ ] IPAM group (14+ items)
  - [ ] Circuits group (5 items)
  - [ ] Virtualization group (5 items)
  - [ ] VPN group (5 items)
  - [ ] Wireless group (3 items)
  - [ ] Other group (Tags, Config Contexts, Custom Fields, Webhooks)
  - [ ] Admin group (Users, Groups, Jobs, Change Log)

### 1.6 Router Setup

- [ ] `src/router/index.ts` — Main router:
  - [ ] Auto-generated model routes from registry
  - [ ] Static routes (login, dashboard, search, error pages)
  - [ ] Navigation guards (auth check)
  - [ ] Scroll behavior reset
- [ ] `src/router/auto-routes.ts` — Route generator:
  - [ ] Generate list/detail/add/edit/delete/bulk routes per model
  - [ ] Generate feature tab routes (contacts, journal, changelog)
  - [ ] Specialized routes (cable trace, rack elevation, etc.)
- [ ] `src/router/model-registry.ts` — Model definitions:
  - [ ] All 104 models with module, name, display_name, api_path
  - [ ] Per-model field schemas (for forms and tables)
  - [ ] Per-model filter definitions
  - [ ] Per-model tab definitions

---

## Phase 2: Generic CRUD Views (Weeks 2-4)

### 2.1 ObjectListView (used by ALL 104 models)

- [ ] `src/pages/generic/ObjectListView.vue`:
  - [ ] Page header with title + action buttons (Add, Import, Export)
  - [ ] Filter panel (toggle button + slide-out panel)
  - [ ] Data table with:
    - [ ] Column sorting (click header)
    - [ ] Bulk selection checkboxes
    - [ ] Select-all across pages
    - [ ] Row click → detail page
  - [ ] Pagination controls (page numbers, page size selector)
  - [ ] Empty state ("No results found")
  - [ ] Loading spinner
  - [ ] URL query string sync (filters + pagination are bookmarkable)
  - [ ] Bulk action bar (appears when rows selected):
    - [ ] "Edit Selected" button → BulkEditView
    - [ ] "Delete Selected" button → BulkDeleteView
    - [ ] Selected count display
    - [ ] Clear selection

### 2.2 Data Table Components

- [ ] `src/components/table/DataTable.vue`:
  - [ ] Accept columns config + data
  - [ ] Sort state management
  - [ ] Slot for custom cell rendering
  - [ ] Responsive (horizontal scroll on mobile)
- [ ] `src/components/table/TableColumn.vue`:
  - [ ] Column definition (key, label, sortable, width)
  - [ ] Custom template slot
- [ ] `src/components/table/BulkSelectBar.vue`:
  - [ ] Sticky bar at bottom
  - [ ] Selected count
  - [ ] Action buttons
  - [ ] Clear button
- [ ] `src/components/table/PaginationControls.vue`:
  - [ ] Previous/Next buttons
  - [ ] Page number buttons (with ellipsis for many pages)
  - [ ] Page size selector (25, 50, 100, 250)
  - [ ] Total count display
- [ ] `src/components/table/ExportButton.vue`:
  - [ ] Export to CSV
  - [ ] Export current page vs all results

### 2.3 Filter Components

- [ ] `src/components/filter/FilterPanel.vue`:
  - [ ] Slide-out panel from right
  - [ ] Render filter fields from model schema
  - [ ] Apply/Clear buttons
  - [ ] Sync with URL query params
- [ ] `src/components/filter/FilterField.vue`:
  - [ ] Route to correct input type based on filter config
- [ ] `src/components/filter/TextFilter.vue` — Text search input
- [ ] `src/components/filter/SelectFilter.vue` — Static dropdown
- [ ] `src/components/filter/ApiSelectFilter.vue`:
  - [ ] API-backed searchable dropdown
  - [ ] Multi-select support
  - [ ] Debounced search
  - [ ] Pagination of results
- [ ] `src/components/filter/BooleanFilter.vue` — Tristate (Yes/No/Any)
- [ ] `src/components/filter/DateRangeFilter.vue` — Date range picker
- [ ] `src/components/filter/IntegerRangeFilter.vue` — Min/max number inputs

### 2.4 ObjectDetailView (used by ALL models)

- [ ] `src/pages/generic/ObjectDetailView.vue`:
  - [ ] Page header with object name + breadcrumb
  - [ ] Action buttons (Edit, Delete, Clone, Add Component)
  - [ ] Tab bar:
    - [ ] Main tab (properties table)
    - [ ] Contacts tab (if ContactsMixin)
    - [ ] Journal tab (if JournalingMixin)
    - [ ] Change Log tab (if ChangeLoggingMixin)
    - [ ] Image Attachments tab (if ImageAttachmentsMixin)
    - [ ] Advanced tab (custom fields, tags)
  - [ ] Properties table (key-value display)
  - [ ] Description/comments section
  - [ ] Tags display
  - [ ] Custom fields panel
- [ ] `src/components/detail/PropertiesTable.vue`:
  - [ ] Two-column key-value layout
  - [ ] Link rendering for FK fields
  - [ ] Boolean → check/cross icon
  - [ ] Date formatting
- [ ] `src/components/detail/DetailTabBar.vue`:
  - [ ] Tab switching
  - [ ] Badge counts on tabs (e.g., contact count)
- [ ] `src/components/detail/StatusBadge.vue`:
  - [ ] Colored pill based on status
  - [ ] Configurable color map per model
- [ ] `src/components/detail/TagsDisplay.vue`:
  - [ ] Colored tag pills
  - [ ] Click to filter

### 2.5 ObjectEditView (used by ALL models)

- [ ] `src/pages/generic/ObjectEditView.vue`:
  - [ ] Page header with title (Add/Edit)
  - [ ] Dynamic form rendered from model schema
  - [ ] Save/Cancel buttons
  - [ ] Validation error display (per-field)
  - [ ] Unsaved changes warning (beforeunload)
  - [ ] Clone support (pre-fill from existing)
- [ ] `src/components/form/DynamicForm.vue`:
  - [ ] Accept field schema + v-model
  - [ ] Group fields into sections
  - [ ] Submit/cancel slot
- [ ] `src/components/form/DynamicField.vue`:
  - [ ] Route to correct field component by type
  - [ ] Label + required indicator
  - [ ] Help text
  - [ ] Error message display

### 2.6 Form Field Components

- [ ] `src/components/form/TextField.vue` — Standard text input
- [ ] `src/components/form/SlugField.vue` — Auto-slug from name (with manual override)
- [ ] `src/components/form/NumberField.vue` — Number input with min/max
- [ ] `src/components/form/BooleanField.vue` — Checkbox/tristate
- [ ] `src/components/form/SelectField.vue` — Static dropdown
- [ ] `src/components/form/ApiSelectField.vue`:
  - [ ] API-backed searchable dropdown
  - [ ] Debounced search
  - [ ] Multi-select support
  - [ ] Create-on-the-fly button (for tags, etc.)
  - [ ] Display nested object name
- [ ] `src/components/form/TagInputField.vue`:
  - [ ] Type to search/create tags
  - [ ] Selected tags as removable pills
  - [ ] Color display
- [ ] `src/components/form/MarkdownField.vue`:
  - [ ] Textarea with preview toggle
  - [ ] Markdown rendering
  - [ ] Cheat sheet toggle
- [ ] `src/components/form/JsonField.vue`:
  - [ ] JSON text area
  - [ ] Syntax validation
  - [ ] Pretty-print toggle
- [ ] `src/components/form/DateField.vue` — Date picker
- [ ] `src/components/form/DateTimeField.vue` — Datetime picker
- [ ] `src/components/form/CustomFieldRenderer.vue`:
  - [ ] Render custom fields by type (text, integer, boolean, date, select, JSON, URL, long text, object)
  - [ ] Choice set support
  - [ ] Required/optional handling
- [ ] `src/components/form/CsvImportField.vue`:
  - [ ] CSV text area paste
  - [ ] File upload support
  - [ ] Column mapping preview
  - [ ] Validation per row

### 2.7 ObjectDeleteView

- [ ] `src/pages/generic/ObjectDeleteView.vue`:
  - [ ] Confirmation modal with object name
  - [ ] Warning about cascading deletes
  - [ ] Confirm/Cancel buttons
- [ ] `src/components/modal/ConfirmModal.vue`:
  - [ ] Reusable confirmation dialog
  - [ ] Customizable title, message, button text

### 2.8 Bulk Operation Views

- [ ] `src/pages/generic/BulkEditView.vue`:
  - [ ] Show selected objects count
  - [ ] Form with fields to update (checkbox to enable each field)
  - [ ] "Set null" option per field
  - [ ] Add/remove tags
  - [ ] Confirm → PATCH all selected
  - [ ] Per-object error report
- [ ] `src/pages/generic/BulkDeleteView.vue`:
  - [ ] Show selected objects list
  - [ ] Warning about cascading deletes
  - [ ] Confirm → DELETE all selected
  - [ ] Per-object error report
- [ ] `src/pages/generic/BulkImportView.vue`:
  - [ ] CSV text area / file upload
  - [ ] Required fields hint
  - [ ] Preview parsed data
  - [ ] Submit → POST batch
  - [ ] Per-row error report
- [ ] `src/pages/generic/BulkRenameView.vue`:
  - [ ] Find-and-replace pattern
  - [ ] Regular expression option
  - [ ] Preview changes
- [ ] `src/pages/generic/BulkAddComponentView.vue`:
  - [ ] Select component type (interface, console port, etc.)
  - [ ] Pattern-based name generation (e.g., "eth[0-9]")
  - [ ] Select target devices
  - [ ] Submit → create components

---

## Phase 3: Feature Panels & Shared Components (Weeks 3-5)

### 3.1 Contacts Panel

- [x] `src/components/detail/ContactsPanel.vue`:
  - [ ] List contact assignments for this object
  - [ ] Add contact (select contact + role)
  - [ ] Edit contact assignment
  - [ ] Remove contact assignment

### 3.2 Journal Panel

- [x] `src/components/detail/JournalPanel.vue`:
  - [ ] List journal entries (chronological)
  - [ ] Add new entry (kind selection: info, success, warning, danger)
  - [ ] Markdown editor for entry
  - [ ] Edit/delete own entries

### 3.3 Change Log Panel

- [x] `src/components/detail/ChangeLogPanel.vue`:
  - [ ] List ObjectChange records for this object
  - [ ] Show action (create/update/delete), user, timestamp
  - [ ] Diff view (prechange vs postchange)
- [x] `src/components/shared/DiffViewer.vue`:
  - [ ] JSON diff rendering
  - [ ] Added/removed/changed field highlighting

### 3.4 Image Attachments Panel

- [x] `src/components/detail/ImageAttachmentsPanel.vue`:
  - [ ] Grid of attached images
  - [ ] Upload new image (file picker)
  - [ ] Image preview modal
  - [ ] Delete attachment

### 3.5 Shared Utility Components

- [ ] `src/components/shared/BreadcrumbNav.vue` — Hierarchical breadcrumbs
- [ ] `src/components/shared/LoadingSpinner.vue` — Spinner/skeleton variants
- [ ] `src/components/shared/EmptyState.vue` — "No results" with icon + optional CTA
- [ ] `src/components/shared/CopyToClipboard.vue` — Click-to-copy text
- [ ] `src/components/shared/Modal.vue` — Base modal (backdrop, close button, slot)
- [ ] `src/components/shared/ToastContainer.vue` — Toast notification stack
- [ ] `src/components/shared/ToastItem.vue` — Individual toast
- [ ] `src/components/shared/CodeBlock.vue` — Syntax-highlighted code display
- [ ] `src/components/shared/JsonViewer.vue` — Collapsible JSON tree
- [ ] `src/components/shared/PillBadge.vue` — Small colored badge
- [ ] `src/components/shared/CountBadge.vue` — Numeric badge (for tabs, menu items)
- [ ] `src/components/shared/ExternalLink.vue` — Opens in new tab
- [ ] `src/components/shared/RelativeTime.vue` — "5 minutes ago" time display
- [ ] `src/components/shared/ConfirmButton.vue` — Button with confirmation step

---

## Phase 4: DCIM Module — Core Models (Weeks 5-8)

### 4.1 Sites & Locations

- [ ] **Site**
  - [ ] List view config (columns: name, status, region, tenant, facility, device count)
  - [ ] Detail view (custom: map coordinates, address, stats)
  - [ ] Form config (name, slug, status, region, group, tenant, facility, ASN, time zone, etc.)
  - [ ] Filter config (name, status, region, group, tenant)
- [ ] **Region**
  - [ ] List config + tree view toggle (MPTT hierarchy)
  - [ ] Detail view (custom: sub-regions, sites)
- [ ] **SiteGroup**
  - [ ] List config + tree view toggle
- [ ] **Location**
  - [ ] List config
  - [ ] Detail view (custom: parent location tree, racks)

### 4.2 Racks

- [ ] **Rack**
  - [ ] List view config
  - [ ] Detail view (**custom: rack elevation diagram**)
    - [ ] `src/pages/special/RackElevationView.vue`
    - [ ] Front/rear elevation display
    - [ ] Device placement coloring (by role)
    - [ ] Reservation highlighting
    - [ ] Click unit → device detail
    - [ ] Unit numbering (bottom-up)
  - [ ] Form config (site, location, name, facility_id, type, width, height, status, role, tenant, serial, etc.)
- [ ] **RackReservation** — List + detail + form
- [ ] **RackRole** — Standard CRUD
- [ ] **RackType** — Standard CRUD

### 4.3 Device Types & Manufacturers

- [ ] **Manufacturer** — Standard CRUD + device_type_count counter
- [ ] **DeviceType**
  - [ ] List config (columns: model, manufacturer, part_number, instance_count)
  - [ ] Detail view (**custom: component templates**):
    - [ ] Console port templates tab
    - [ ] Console server port templates tab
    - [ ] Power port templates tab
    - [ ] Power outlet templates tab
    - [ ] Interface templates tab
    - [ ] Front port templates tab
    - [ ] Rear port templates tab
    - [ ] Device bay templates tab
    - [ ] Module bay templates tab
    - [ ] Inventory item templates tab
  - [ ] Form config (manufacturer, model, slug, part_number, u_height, is_full_depth, etc.)
- [ ] **DeviceRole** — Standard CRUD
- [ ] **Platform** — Standard CRUD

### 4.4 Devices

- [ ] **Device**
  - [ ] List config (columns: name, status, tenant, site, location, rack, position, device_type, role, primary_ip)
  - [ ] Detail view (**custom, most complex**):
    - [ ] Main tab: properties + status
    - [ ] Console ports tab (with cable status)
    - [ ] Console server ports tab
    - [ ] Power ports tab
    - [ ] Power outlets tab
    - [ ] Interfaces tab (with IP addresses, VLANs, cable status)
    - [ ] Front ports tab
    - [ ] Rear ports tab
    - [ ] Device bays tab (installed devices)
    - [ ] Module bays tab (installed modules)
    - [ ] Inventory items tab (**tree view**)
    - [ ] Config context tab
    - [ ] Services tab
    - [ ] Images tab
    - [ ] LLDP neighbors (if available)
    - [ ] Virtual chassis membership
  - [ ] Form config (name, device_type, role, tenant, platform, site, location, rack, position, face, status, serial, primary_ip4, primary_ip6, etc.)
- [ ] **VirtualChassis**
  - [ ] Detail view (custom: member devices list with vc_position/vc_priority)

### 4.5 Modules

- [ ] **Module** — Standard CRUD (link to module_bay + module_type)
- [ ] **ModuleType** — Standard CRUD
- [ ] **ModuleTypeProfile** — Standard CRUD
- [ ] **ModuleBay** — Standard CRUD

### 4.6 Device Components (each with trace + connect)

For each component type:
- [ ] **ConsolePort**
  - [ ] List config
  - [ ] Detail view with cable connection info
  - [ ] **Trace view** (`src/pages/special/CableTraceView.vue`)
  - [ ] **Connect form** (select cable / create cable)
- [ ] **ConsoleServerPort** — Same as ConsolePort
- [ ] **PowerPort** — Same + power outlet feed assignment
- [ ] **PowerOutlet** — Same + power port link
- [ ] **Interface** (most complex component):
  - [ ] List config
  - [ ] Detail view (custom: IP addresses, VLANs, MAC addresses, cable, wireless, LAG)
  - [ ] Form (type, speed, duplex, MTU, MAC, description, mode, untagged VLAN, tagged VLANs, VRF, etc.)
  - [ ] **Trace view**
  - [ ] **Connect form**
  - [ ] **IP address assignment** sub-form
  - [ ] **VLAN assignment** sub-form
- [ ] **FrontPort** — Standard + trace + connect
- [ ] **RearPort** — Standard + trace + connect
- [ ] **DeviceBay** — Standard CRUD + installed device selection
- [ ] **InventoryItem**
  - [ ] Detail view (**custom: hierarchical tree**)
  - [ ] `src/components/detail/InventoryItemTree.vue` (recursive component)
- [ ] **InventoryItemRole** — Standard CRUD
- [ ] **MACAddress** — Standard CRUD
- [ ] **VirtualDeviceContext** — Standard CRUD

### 4.7 Cables & Power

- [ ] **Cable**
  - [ ] List config (columns: id, termination_a, termination_b, type, status, length)
  - [ ] Detail view (custom: shows both endpoints with device/context)
  - [ ] Form (termination_a, termination_b, type, status, length, length_unit, label, color, description, tags)
- [ ] **CableTermination** — System-managed (read-only display)
- [ ] **CablePath** — System-managed (read-only display)
- [ ] **PowerPanel** — Standard CRUD + power feeds list
- [ ] **PowerFeed**
  - [ ] Detail view (custom: connected power port, power panel info)
  - [ ] Trace view (like cable trace)

### 4.8 Cable Trace Visualization

- [ ] `src/pages/special/CableTraceView.vue`:
  - [ ] Fetch trace data from API (`/{module}/{model}/{id}/trace/`)
  - [ ] **SVG rendering** of path:
    - [ ] Near end device/port boxes
    - [ ] Cable connections (lines with color)
    - [ ] Far end device/port boxes
    - [ ] Intermediate devices (patch panels, etc.)
  - [ ] Split path indicator + selector
  - [ ] Total length display (meters + feet)
  - [ ] "Trace completed" / "Path split!" status
  - [ ] Legend

### 4.9 DCIM Component Templates

For each template type (same standard CRUD, used in DeviceType detail):
- [ ] ConsolePortTemplate
- [ ] ConsoleServerPortTemplate
- [ ] PowerPortTemplate
- [ ] PowerOutletTemplate
- [ ] InterfaceTemplate
- [ ] FrontPortTemplate
- [ ] RearPortTemplate
- [ ] DeviceBayTemplate
- [ ] ModuleBayTemplate
- [ ] InventoryItemTemplate

---

## Phase 5: IPAM Module (Weeks 8-10)

### 5.1 VRFs & Route Targets

- [ ] **VRF** — Standard CRUD + route targets + prefix/IP counters
- [ ] **RouteTarget** — Standard CRUD

### 5.2 Prefixes & Aggregates

- [ ] **Prefix**
  - [ ] List config (columns: prefix, status, vrf, tenant, site, vlan, role)
  - [ ] Detail view (**custom: hierarchy tree**):
    - [ ] `src/pages/special/PrefixTreeView.vue` (collapsible nested tree)
    - [ ] Child prefixes list
    - [ ] IP addresses under this prefix
    - [ ] Available prefixes calculator
    - [ ] IP utilization bar
  - [ ] Form (prefix, vrf, tenant, status, role, scope, is_pool, description, tags, custom fields)
  - [ ] Filter (prefix, vrf, status, role, scope, within, contains)
- [ ] **Aggregate** — Standard CRUD (prefix, rir, tenant, date_added, description)
- [ ] **IPRange** — Standard CRUD
- [ ] **RIR** — Standard CRUD

### 5.3 IP Addresses

- [ ] **IPAddress**
  - [ ] List config (columns: address, status, vrf, tenant, assigned_object, role, dns_name)
  - [ ] Detail view (custom: assignment info, NAT inside/outside)
  - [ ] Form (address, vrf, tenant, status, role, dns_name, description, assigned_object, tags)
  - [ ] **Assignment sub-form** (select interface/VM interface)
  - [ ] **Bulk assign** (assign IPs to interfaces)

### 5.4 VLANs

- [ ] **VLAN**
  - [ ] List config (columns: vid, name, status, group, site, tenant, role, description)
  - [ ] Detail view (custom: prefixes, interfaces tagged/untagged)
  - [ ] Form (vid, name, status, group, site, tenant, role, description, tags, custom fields)
- [ ] **VLANGroup**
  - [ ] Detail view (custom: VLAN availability grid, scope)
  - [ ] VLAN range generator
- [ ] **VLANTranslationPolicy** — Standard CRUD
- [ ] **VLANTranslationRule** — Standard CRUD

### 5.5 ASNs

- [ ] **ASN** — Standard CRUD + site association
- [ ] **ASNRange** — Standard CRUD + ASN generator

### 5.6 Services

- [ ] **Service** — Standard CRUD (name, protocol, ports, device/VM assignment)
- [ ] **ServiceTemplate** — Standard CRUD
- [ ] **FHRPGroup** — Standard CRUD + protocol/auth
- [ ] **FHRPGroupAssignment** — Standard CRUD

---

## Phase 6: Remaining Modules (Weeks 10-13)

### 6.1 Tenancy

- [ ] **Tenant** — Standard CRUD + group + counters
- [ ] **TenantGroup** — Standard CRUD + MPTT tree
- [ ] **Contact** — Standard CRUD + assignments
- [ ] **ContactGroup** — Standard CRUD + MPTT tree
- [ ] **ContactRole** — Standard CRUD
- [ ] **ContactAssignment** — Standard CRUD (link contact + object + role)

### 6.2 Circuits

- [ ] **Circuit**
  - [ ] List config (cid, provider, type, status, tenant, termination_a, termination_z)
  - [ ] Detail view (custom: terminations, connections)
  - [ ] Form (cid, provider, type, status, tenant, install_date, commit_rate, description, terminations)
- [ ] **CircuitTermination** — Standard CRUD + cable trace + connect
- [ ] **CircuitType** — Standard CRUD
- [ ] **Provider** — Standard CRUD (custom detail: circuits, accounts, networks)
- [ ] **ProviderNetwork** — Standard CRUD
- [ ] **Account** — Standard CRUD

### 6.3 Virtualization

- [ ] **VirtualMachine**
  - [ ] List config (name, status, cluster, role, tenant, vcpus, memory, disk)
  - [ ] Detail view (custom: resources, interfaces, services, primary IPs)
  - [ ] Form (name, status, site, cluster, device, role, tenant, platform, vcpus, memory, disk, serial, primary_ip4, primary_ip6, etc.)
- [ ] **VMInterface** — Standard CRUD + IP assignment + VLAN assignment + cable
- [ ] **Cluster** — Standard CRUD + VM list
- [ ] **ClusterType** — Standard CRUD
- [ ] **ClusterGroup** — Standard CRUD

### 6.4 VPN

- [ ] **Tunnel** — Standard CRUD (status, type, ipsec_profile, encapsulation, etc.)
- [ ] **TunnelTermination** — Standard CRUD (interface/VM interface assignment)
- [ ] **IPSecProfile** — Standard CRUD (complex crypto fields)
- [ ] **L2VPN** — Standard CRUD + terminations
- [ ] **L2VPNTermination** — Standard CRUD (interface/VM interface assignment)

### 6.5 Wireless

- [ ] **WirelessLAN** — Standard CRUD (ssid, scope, group, status, description, vlan)
- [ ] **WirelessLANGroup** — Standard CRUD + MPTT tree
- [ ] **WirelessLink**
  - [ ] List + detail (interface_a, interface_b, status, ssid)
  - [ ] Form (status, interface_a, interface_b, ssid, description, auth_type, auth_cipher, auth_psk)

### 6.6 Extras

- [ ] **ConfigContext**
  - [ ] Detail view (custom: weight, regions, sites, groups, roles, platforms, clusters, tenants, tags, data JSON)
  - [ ] Form (name, weight, description, regions/sites/etc. multi-selects, data JSON editor)
  - [ ] Config context rendering preview
- [ ] **ConfigTemplate** — Standard CRUD + template syntax validation
- [ ] **CustomField**
  - [ ] List config (name, type, required, content types)
  - [ ] Form (name, label, type, group_name, description, required, default, filter_logic, choices, weight)
- [ ] **CustomFieldChoiceSet** — Standard CRUD (choices JSON/CSV editor)
- [ ] **Tag** — Standard CRUD (name, slug, color)
- [ ] **Webhook** — Standard CRUD (content types, payload URL, HTTP method, body template, etc.)
- [ ] **JournalEntry** — Standard CRUD (kind, assigned_object, created_by, comment)
- [ ] **ImageAttachment** — Standard CRUD (upload form)

### 6.7 Users

- [ ] **User** — Standard CRUD (username, email, first_name, last_name, is_active, is_staff, groups)
- [ ] **Group** — Standard CRUD (name, users)
- [ ] **Token** — Standard CRUD (key, user, write_enabled, expires)

### 6.8 Core

- [ ] **DataSource** — Standard CRUD + sync trigger + data file browser
- [ ] **DataFile** — Read-only display (source, path, size, hash, updated)
- [ ] **Job** — Read-only display (status, created, completed, data, log entries)
- [ ] **ObjectChange** — Read-only display + diff viewer
- [ ] **ObjectType** — Read-only display

---

## Phase 7: Specialized Pages (Weeks 13-16)

### 7.1 Authentication Pages

- [ ] `src/pages/auth/LoginPage.vue`:
  - [ ] Username/password form
  - [ ] LDAP/SSO redirect support
  - [ ] Error handling
  - [ ] Redirect after login
- [ ] `src/pages/auth/LogoutAction.vue` — Token clear + redirect

### 7.2 Dashboard

- [ ] `src/pages/special/DashboardView.vue`:
  - [ ] Configurable widget grid
  - [ ] `src/components/dashboard/ObjectCountWidget.vue` — Count by model type
  - [ ] `src/components/dashboard/RecentChangesWidget.vue` — Recent ObjectChange list
  - [ ] `src/components/dashboard/ChartWidget.vue` — Pie/bar chart (Chart.js)
  - [ ] `src/components/dashboard/CustomLinkWidget.vue` — User-defined links
  - [ ] Widget add/remove/rearrange controls
  - [ ] Persist layout to user preferences

### 7.3 Global Search

- [ ] `src/pages/special/SearchResultsView.vue`:
  - [ ] Search input with debouncing
  - [ ] Categorized results by model type
  - [ ] Result count per category
  - [ ] Click result → detail page
  - [ ] Faceted filter (filter by model type)

### 7.4 Reports & Scripts

- [ ] `src/pages/special/ReportsView.vue`:
  - [ ] Report module list
  - [ ] Report detail (test methods, description)
  - [ ] Run report button
  - [ ] Results table (pass/fail/log per test)
- [ ] `src/pages/special/ScriptsView.vue`:
  - [ ] Script module list
  - [ ] Script execution form (dynamic variables)
  - [ ] Execution results/log

### 7.5 Data Source Sync

- [ ] `src/pages/special/DataSourceSyncView.vue`:
  - [ ] Sync trigger button
  - [ ] Sync status display
  - [ ] Data file browser (tree view of synced files)
  - [ ] File content preview

### 7.6 Background Jobs

- [ ] `src/pages/special/JobsView.vue`:
  - [ ] Job queue table (name, status, created, completed, user)
  - [ ] Filter by status (pending, running, completed, failed, errored)
  - [ ] Job detail (data, log entries, error trace)
  - [ ] Retry/cancel actions (if supported)

### 7.7 GraphQL Explorer

- [ ] `src/pages/special/GraphiQLView.vue`:
  - [ ] Embed GraphiQL component
  - [ ] Configure GraphQL endpoint
  - [ ] Auth token injection
  - [ ] Schema explorer sidebar

### 7.8 API Browser

- [ ] `src/pages/special/ApiDocsView.vue`:
  - [ ] Embed Swagger UI / OpenAPI viewer
  - [ ] Configure spec URL
  - [ ] Try-it-out with auth

### 7.9 Error Pages

- [ ] `src/pages/errors/ForbiddenView.vue` (403)
- [ ] `src/pages/errors/NotFoundView.vue` (404)
- [ ] `src/pages/errors/ServerErrorView.vue` (500)

---

## Phase 8: Polish & Optimization (Weeks 16-18)

### 8.1 Performance

- [ ] Implement route-level code splitting (lazy imports)
- [ ] Add API response caching (stale-while-revalidate pattern)
- [ ] Virtual scrolling for large tables (e.g., IP address lists with 10k+ items)
- [ ] Debounce filter inputs
- [ ] Optimize bundle size (analyze with `vite-bundle-visualizer`)

### 8.2 UX Polish

- [ ] Loading skeletons for all data fetches
- [ ] Smooth page transitions
- [ ] Keyboard shortcuts (Ctrl+K for search, Esc to close modals)
- [ ] Responsive design audit (mobile/tablet layouts)
- [ ] Dark mode refinement (all components tested)
- [ ] Accessibility audit (ARIA labels, keyboard nav, focus traps)
- [ ] Empty states with helpful CTAs
- [ ] Error states with retry actions

### 8.3 Security

- [ ] CSRF token handling
- [ ] XSS prevention (v-html only for trusted content)
- [ ] Session timeout handling
- [ ] RBAC enforcement in UI (hide/disable buttons user can't use)

### 8.4 Testing

- [ ] Unit tests for composables (`useApi`, `useTable`, etc.)
- [ ] Component tests for generic views (ObjectListView, ObjectEditView)
- [ ] E2E tests for critical flows (login, create device, cable trace)
- [ ] Visual regression tests for layout components

---

## Tracking Metrics

| Metric | Target | Current |
|---|---|---|
| Generic components built | 12 | 0 |
| Shared/layout components built | 46 | 0 |
| Per-model detail overrides built | ~35 | 0 |
| Specialized page components built | 15 | 0 |
| **Total components** | **~108** | **0** |
| Routes configured | ~1,150 | 0 |
| Model registry entries | 104 | 0 |

---

## Notes

- **Do NOT create per-model Vue components for list/edit/delete views.** Use the generic views + model registry config.
- **DO create per-model detail overrides** only when the detail page needs custom tabs or layouts.
- The model registry (`src/router/model-registry.ts`) is the single source of truth for field schemas, filters, columns, and tabs.
- All API calls go through the generic `useApi` composable — no direct Axios calls in components.
- Forms are rendered dynamically from the field schema in the registry.
