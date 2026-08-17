# NetBox Go frontend

Vue 3 and TypeScript operator interface for the standalone Go rewrite. The
normal runtime exposes only the `core-workflow-v1` capability profile: ten DCIM
resources, three IPAM resources, browser-session authentication, and the
canonical OpenAPI browser.

## Local development

The pinned toolchain is Node 24.18.0 and npm 11.16.0.

```bash
npm ci
npm run dev
```

Vite listens on `http://localhost:3000` and proxies `/api` to the Go backend at
`http://localhost:8080`. Set `VITE_API_BASE_URL` only when the API is hosted at
a different origin. For that cross-origin deployment, the backend's
`NETBOX_CORS_ALLOWED_ORIGINS` must include the frontend page origin in its
comma-separated list of exact HTTP(S) origins. The setting defaults to empty,
which grants no cross-origin access; wildcard and regular-expression origins
are rejected.

The browser authenticates with the Go-owned HttpOnly session cookie and a CSRF
cookie/header pair. Passwords, session credentials, API tokens, and permissions
are never stored in `localStorage`.

## Runtime boundaries

- `src/features/core/manifest.ts` is the closed runtime capability manifest.
  Routes, navigation, dashboard cards, relationship lookups, and resource API
  paths all resolve through it.
- `src/router/models/core-profile.ts` contains presentation metadata for those
  13 resources. Legacy model catalogues are not registered at runtime.
- `src/features/*/api.ts` modules own wire DTOs and exact snake-case REST
  mapping. Pages and components do not build endpoint or polymorphic assignment
  payloads.
- IP assignment is presented as an Interface selection. The IPAM feature mapper
  alone translates it to the baseline `assigned_object_type` and
  `assigned_object_id` pair, including explicit-null unassignment.
- Forms enforce profile semantics that are easy to misrepresent generically:
  Interface ownership is immutable after creation, Rack Type-owned dimensions
  become read-only, and Device Rack choices are scoped to the selected Site.
- Deferred bulk, GraphQL, scripts, reports, rack-elevation, cable, allocation,
  tag, and custom-field surfaces are absent from the supported route tree.

The API browser loads `/api/schema/` from the running Go service. It does not
derive a contract from the Vue registry.

## Quality gate

```bash
npm run check
```

This runs the toolchain guard, formatting, lint, application and test
typechecks, coverage tests, and a production build. Component tests cover the
manifest surface, asynchronous edit hydration, typed relationship values,
permission decisions, browser-session state, IP assignment/unassignment, and
the Interface deletion cascade warning. The profile tests also pin all 206
baseline Interface types and every displayed server-supported ordering field.
