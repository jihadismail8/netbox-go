# Job

> Module: `core` | Table: `core_job` | Python class: `Job` | File: `core/models/jobs.py`

**Inheritance:** `NetBoxModel`

**REST URL:** `/api/core/jobs/`

## Django Model Fields

| Field | Type | Notes |
|-------|------|-------|
| `object_type` | FK to ContentType | PROTECT |
| `object_id` | PositiveBigIntegerField | null=True |
| `name` | CharField(100) | Required |
| `job_id` | UUIDField | Required |
| `status` | CharField(50) | choices=JobStatusChoices |
| `created` | DateTimeField | auto_now_add |
| `scheduled` | DateTimeField | null=True |
| `interval` | PositiveIntegerField | null=True |
| `started` | DateTimeField | null=True |
| `completed` | DateTimeField | null=True |
| `user` | FK to User | SET_NULL, null=True |
| `data` | JSONField | default=dict |
| `error` | TextField | blank=True |

## Notes

- **Python source:** `core/models/jobs.py`
- Background job tracking (reports, scripts, etc.)

## CRUD Behavior

> See [CRUD Paradigm](../../CRUD_PARADIGM.md) for system-wide patterns.

**Categories:** A (Simple, system-managed)

### CREATE
Job records are created by the background task system. Not directly creatable via API.

### UPDATE
Not directly updatable. Status updated by job runner.

### DELETE
1. **CASCADE:** JobLogEntry.
2. Change log + event.
