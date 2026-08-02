---
status: accepted
---

# Use unified authentication and authorization across REST, gRPC, and the browser

All Managed Object and domain operations require an authenticated principal by default. Narrow health, readiness, login, and CSRF entry points may be public but cannot expose or mutate Managed Objects. The Vue application will use a secure `HttpOnly`, `SameSite` session cookie; REST automation will support NetBox-compatible API-token authentication; and gRPC callers will send bearer credentials in request metadata. Every credential type resolves to the same user, group membership, token restrictions, and object-level permission evaluation in the shared application layer, avoiding the security and behavioral drift of transport-specific authorization.

## Consequences

- Anonymous mutations are always forbidden. Optional anonymous read-only access is deferred and, if introduced, must be explicitly enabled.
- Browser credentials must not be stored in `localStorage`; state-changing cookie-authenticated requests require CSRF protection.
- A clean standalone database needs a Go-owned one-time administrator
  bootstrap, protected CLI password reset, password lifecycle, and session
  create/revoke/rotation flow; none may require Django. The local CLI may also
  create a non-superuser and grant a global model permission by username, but
  those operations must authenticate an existing active superuser and accept
  passwords only through protected stdin.
- Token management uses an authenticated REST extension with paired identity RPCs and one-time secret disclosure. The baseline anonymous username/password token-provision action is deliberately unsupported and remains an explicit compatibility gap.
- User and authentication serializers must never expose password hashes, session secrets, or reusable token material outside a documented creation response.
- Authentication verifies identity, while the shared authorization service separately enforces view, add, change, and delete permissions plus object constraints.
- REST and gRPC handlers may parse different credentials but may not implement separate permission rules.
- Local bootstrap and identity administration are transport-scope operations,
  not Managed Object capabilities. They do not require cookie-shaped or RPC
  equivalents and must never become anonymous network endpoints.
- Demo auto-login, dummy tokens, and browser-stored credentials are prohibited in the supported runtime; test doubles must remain explicit test-only fixtures.
