# VLANTranslationPolicy

> Module: `ipam` | Table: `ipam_vlantranslationpolicy` | Python class: `VLANTranslationPolicy` | File: `ipam/models/vlans.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/vlan-translation-policies/`

## Implementation Status

- [ ] Go model (`internal/model/ipamVlantranslationpolicy.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] Proto generated code
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Error codes defined
- [ ] REST URL matches NetBox convention
- [ ] Vue.js views

## Django Model Fields

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | CharField(100) | Required |

## Referenced By

- [VLANTranslationRule](./ipamVlantranslationrule.md) via `policy` (FK)
- [Interface](./../dcim/dcimInterface.md) via `vlan_translation_policy` (FK)

## Notes

- **Python source:** `ipam/models/vlans.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple Organizational)

### CREATE
Standard flow. No side effects.

### UPDATE
Standard flow. No side effects.

### DELETE
No downstream effects.
