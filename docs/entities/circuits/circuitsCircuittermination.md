# CircuitTermination

> Module: `circuits` | Table: `circuits_circuittermination` | Python class: `CircuitTermination` | File: `circuits/models/circuits.py`

**Inheritance:** `PrimaryModel <- CabledObjectModel <- PathEndpoint`

**REST URL:** `/api/circuits/circuit-terminations/`

## Implementation Status

- [ ] Go model (`internal/model/circuitsCircuittermination.go`)
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
| `circuit` | [Circuit](./circuitsCircuit.md) | `CASCADE` | No | `terminations` |
| `site` | [Site](./../dcim/dcimSite.md) | `PROTECT` | Yes | `circuit_terminations` |
| `provider_network` | [ProviderNetwork](./circuitsProvidernetwork.md) | `PROTECT` | Yes | `circuit_terminations` |
| `port_speed` | PositiveIntegerField | — | Yes | (Kbps) |
| `upstream_speed` | PositiveIntegerField | — | Yes | (Kbps) |
| `xconnect_id` | CharField(100) | — | Yes | blank=True |
| `pp_info` | CharField(100) | — | Yes | blank=True |
| `description` | CharField(200) | — | Yes | blank=True |
| `cable` | [Cable](./../dcim/dcimCable.md) | `SET_NULL` | Yes | `+` |
| `_path` | [CablePath](./../dcim/dcimCablepath.md) | `SET_NULL` | Yes | — |

### Regular Fields

| Field | Type | Notes |
|-------|------|-------|
| `term_side` | CharField(1) | choices=CircuitTerminationSideChoices (A or Z); required |

## Notes

- **Python source:** `circuits/models/circuits.py`
- Either `site` or `provider_network` must be set (mutually exclusive)

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** D (Cable-Connected), C (Cache Consumer)

### CREATE
1. `clean()` validates `circuit` set, termination point (site or provider_network).
2. Cache (`_site`) populated.
3. Save.

### UPDATE
Standard flow.

### DELETE
1. **Cable disconnect:** If connected, cable termination removed, path recomputed.
2. Change log + event.
