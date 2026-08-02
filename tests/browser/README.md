# Core browser E2E gate

`make browser-e2e` starts the disposable root Compose stack, creates an
ephemeral administrator plus a view-only user through the local administration
CLI, grants the latter only `dcim.view_site`, and drives the compiled Vue
application in installed headless Google Chrome. CLI authentication and new
passwords are supplied on standard input and never appear in CLI arguments or
command output. The browser driver receives the two ephemeral login credentials
through its private process environment, never through Chrome arguments, and
redacts them from every retained diagnostic.

The driver speaks Chrome DevTools Protocol using Node.js built-ins. It does not
download Playwright, a browser, or any other package. The gate covers:

- unauthenticated redirect, cookie-session login/logout, and HTTP 401 behavior;
- the absence of credentials in `localStorage` and `sessionStorage`;
- Vue-driven creation of every `core-workflow-v1` resource, including Rack
  Role/Type relationships and Device instantiation of an Interface Template;
- Vue-driven VRF, Prefix, and IP address creation;
- Interface assignment and unassignment of an IP address;
- surfaced uniqueness validation and rollback;
- real limited-user RBAC: Site visibility, the UI 403 route guard, a
  CSRF-valid server-side mutation denial, and unchanged persisted state;
- delete cancellation and confirmation, including the Interface warning and
  cascade behavior for an assigned IP address.

On failure the command reports a private temporary artifact directory holding
a screenshot, sanitized DOM, Chrome log, and a summary containing only request
method/URL/status metadata. Request bodies, headers, cookies, and credentials
are never recorded.

To run the driver against an already-running development instance:

```bash
NETBOX_E2E_BASE_URL=http://127.0.0.1:8080 \
NETBOX_E2E_USERNAME=browser-e2e \
NETBOX_E2E_PASSWORD='use-an-ephemeral-password' \
NETBOX_E2E_LIMITED_USERNAME=browser-e2e-viewer \
NETBOX_E2E_LIMITED_PASSWORD='use-a-second-ephemeral-password' \
./tests/browser/run.sh
```

The two users must already exist for a standalone run, and the limited user
must have only `dcim.view_site` for this profile scenario.
