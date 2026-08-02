# VLANTranslationRule

> Module: `ipam` | Table: `ipam_vlantranslationrule` | Python class: `VLANTranslationRule` | File: `ipam/models/vlans.py`

**Inheritance:** `PrimaryModel`

**REST URL:** `/api/ipam/vlan-translation-rules/`

## Implementation Status

- [ ] Go model (`internal/model/ipamVlantranslationrule.go`)
- [ ] GORM mapping verified
- [ ] Proto definition
- [ ] DAO layer
- [ ] Service layer
- [ ] Handler layer
- [ ] HTTP routes registered
- [ ] Vue.js views

## Django Model Fields

### Foreign Keys

| Field | Related Model | on_delete | null | related_name |
|-------|---------------|-----------|------|--------------|
| `policy` | [VLANTranslationPolicy](./ipamVlantranslationpolicy.md) | `CASCADE` | No | `rules` |
| `local_vid` | [VLAN](./ipamVlan.md) | `PROTECT` | No | — |
| `remote_vid` | PositiveIntegerField | — | — | Required; min=1, max=4094 |

## Notes

- **Python source:** `ipam/models/vlans.py`

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple)

### CREATE
Standard flow. Links VLANTranslationPolicy to VLAN.

### UPDATE
Standard flow.

### DELETE
No downstream effects.
