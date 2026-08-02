# Business Logic Reference — Python NetBox → Go Port

> **Reference analysis, not an accepted or complete specification.** These notes were extracted from the post-4.4.6 NetBox snapshot at commit `fbb948d30e79ce657fac62994a22aca72c1770a9`; they may omit behavior and do not establish what the current Go code implements. Verify implementation decisions against the canonical [architecture](../ARCHITECTURE.md) and [compatibility contract](../COMPATIBILITY.md).

> These documents extract non-trivial business rules from Python/Django NetBox
> that may need to be reproduced in the Go backend. They are a discovery aid for
> validation logic, computed fields, and custom endpoints; the pinned source and
> executable compatibility tests remain authoritative.

---

## Document Index

| Document | Description | Key Models |
|----------|-------------|------------|
| **[dcim-models.md](dcim-models.md)** | DCIM model fields, relationships, clean() rules, and computed properties | Device, DeviceType, Interface, Rack, Cable, Site, Region, Location, Manufacturer, ConsolePort, PowerPort, FrontPort, RearPort, etc. |
| **[ipam-models.md](ipam-models.md)** | IPAM model fields, IP/CIDR validation, overlap detection, utilization calculations | Prefix, IPAddress, IPRange, VLAN, VRF, Aggregate, ASN, FHRPGroup, Service |
| **[api-patterns.md](api-patterns.md)** | REST API serialization formats, response shapes, custom @action endpoints, bulk operations, count annotations | All modules — response format reference for handlers |
| **[filtersets.md](filtersets.md)** | Dynamic lookup expression system (__ic, __n, __gte), per-model filter field catalog | Site, Device, Rack, Prefix, IPAddress, VLAN, Circuit, VM, VMInterface |

---

## How to Use These Documents

### When implementing a Go model (GORM struct)
1. Check **dcim-models.md** or **ipam-models.md** for the model's field list and constraints
2. Implement `clean()` rules as a `Validate()` method on the model or in the service layer
3. Implement computed properties as Go methods (e.g., `func (p Prefix) GetUtilization() float64`)
4. Add `BeforeSave` / `BeforeCreate` / `BeforeUpdate` GORM hooks for save() behaviors (e.g., CIDR normalization)

### When implementing a REST handler
1. Check **api-patterns.md** for the expected JSON response format
2. Implement nested serializers for FK fields (id, url, display, name)
3. Add count annotations via GORM subqueries or preload counts
4. Implement custom action endpoints (available-ips, trace, elevation, etc.) as separate routes

### When implementing filtering
1. Check **filtersets.md** for the model's supported filter fields
2. Use the dynamic lookup expression parser to auto-generate __ic, __n, __gte variants
3. Implement custom method filters as separate handler functions
4. Support comma-separated multi-value → SQL IN clause

---

## Cross-Cutting Concerns

### Model Inheritance Hierarchy (Python → Go)

```
NetBoxModel (abstract)
  ├─ created (DateTimeField, auto_now_add)
  ├─ last_updated (DateTimeField, auto_now)
  ├─ tags (JSONField)
  ├─ custom_field_data (JSONField)
  │
  ├─ PrimaryModel(NetBoxModel)
  │    ├─ description (CharField, blank)
  │    └─ comments (TextField, blank)
  │
  └─ OrganizationalModel(NetBoxModel)
       └─ slug (SlugField)
```

**Go mapping:** Embed a `BaseModel` struct in every GORM model:

```go
type BaseModel struct {
    ID              uint       `gorm:"primaryKey"`
    Created         time.Time  `gorm:"autoCreateTime"`
    LastUpdated     time.Time  `gorm:"autoUpdateTime"`
    Tags            datatypes.JSON
    CustomFieldData datatypes.JSON
}

type PrimaryModel struct {
    BaseModel
    Description string
    Comments    string
}

type OrganizationalModel struct {
    BaseModel
    Slug string `gorm:"uniqueIndex"`
}
```

### Validation Pattern (Python clean() → Go Validate())

Python's `clean()` method runs field-level and cross-field validation before save.
In Go, implement this as a `Validate()` method called from the service layer:

```go
func (p *Prefix) Validate(db *gorm.DB) error {
    // No /0 mask
    if p.PrefixLen == 0 {
        return errors.New("cannot create prefix with /0 mask")
    }
    // Reject host-bit input at the public boundary; retain canonical storage.
    // Overlap detection
    // Unique enforcement
    return nil
}
```

### PostgreSQL-Specific Features Requiring Go Equivalents

| Python/Django Feature | Go Equivalent |
|----------------------|---------------|
| `IPNetworkField` (netaddr) | `netip.Prefix` (stdlib) + GORM serializer |
| `IPAddressField` (netaddr) | `netip.Addr` (stdlib) + GORM serializer |
| GIST index with `inet_ops` | Raw SQL migration `CREATE INDEX ... USING gist (...)` |
| `Host()` function | `netip.Addr.Next()` or custom SQL `HOST()` |
| `ArrayField(IntegerField)` | `datatypes.JSON` or `pq.Int64Array` |
| `GenericForeignKey` | Polymorphic association: `ObjectType string` + `ObjectID uint` + resolver |
| `F()` expressions for ordering | Explicit ORDER BY with NULLS FIRST |
| `django.contrib.postgres.search` | PostgreSQL full-text search via `gorm.io/gorm/fulltext` or raw SQL |
