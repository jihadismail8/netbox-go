#!/usr/bin/env node

// Real-browser core workflow gate implemented directly on Chrome DevTools
// Protocol. This deliberately uses only Node built-ins so the repository does
// not need a second JavaScript dependency tree or a network install in CI.

import fs from "node:fs";
import path from "node:path";

const baseURL = requiredURL("NETBOX_E2E_BASE_URL");
const username = required("NETBOX_E2E_USERNAME");
const password = required("NETBOX_E2E_PASSWORD");
const limitedUsername = required("NETBOX_E2E_LIMITED_USERNAME");
const limitedPassword = required("NETBOX_E2E_LIMITED_PASSWORD");
const knownSecrets = [password, limitedPassword];
const cdpPort = Number(required("NETBOX_E2E_CDP_PORT"));
const artifactDir = path.resolve(required("NETBOX_E2E_ARTIFACT_DIR"));
const timeoutMs = positiveInteger(
  process.env.NETBOX_E2E_TIMEOUT_MS ?? "30000",
  "NETBOX_E2E_TIMEOUT_MS",
);

fs.mkdirSync(artifactDir, { recursive: true, mode: 0o700 });

const diagnostics = {
  started_at: new Date().toISOString(),
  base_origin: baseURL.origin,
  profile_resources: [
    "site",
    "manufacturer",
    "rackrole",
    "racktype",
    "rack",
    "devicerole",
    "devicetype",
    "interfacetemplate",
    "device",
    "interface",
    "vrf",
    "prefix",
    "ipaddress",
  ],
  steps: [],
  console: [],
  network_failures: [],
};

let client;
let succeeded = false;

async function runBrowserE2E() {
  try {
    const page = await createPage(cdpPort);
    client = await CDPClient.connect(page.webSocketDebuggerUrl);
    await Promise.all([
      client.command("Page.enable"),
      client.command("Runtime.enable"),
      client.command("Network.enable"),
    ]);
    attachDiagnostics(client, diagnostics);

    await step("unauthenticated UI is redirected to login", async () => {
      await navigate("/dcim/sites/");
      await waitFor("login redirect", async () => {
        const state = await locationState();
        return state.pathname === "/login" && state.search.includes("next=");
      });
      await waitForText("Sign In");
    });

    await step("cookie session login and browser storage safety", async () => {
      await fillField("Username", username);
      await fillField("Password", password);
      await clickText("Sign In", "button");
      await waitForPath("/dcim/sites/");
      await waitForText("Sites");
      await assertCredentialFreeStorage();
    });

    const names = {
      site: "Browser E2E Site",
      siteSlug: "browser-e2e-site",
      disposableSite: "Browser E2E Disposable Site",
      disposableSiteSlug: "browser-e2e-disposable-site",
      rack: "Browser E2E Rack",
      rackRole: "Browser E2E Production Rack",
      rackRoleSlug: "browser-e2e-production-rack",
      rackType: "Browser E2E 42U Cabinet",
      rackTypeSlug: "browser-e2e-42u-cabinet",
      manufacturer: "Browser E2E Manufacturer",
      manufacturerSlug: "browser-e2e-manufacturer",
      deviceType: "Browser E2E Switch",
      deviceTypeSlug: "browser-e2e-switch",
      deviceRole: "Browser E2E Access Switch",
      deviceRoleSlug: "browser-e2e-access-switch",
      device: "browser-e2e-sw01",
      interface: "Ethernet1",
      vrf: "Browser E2E Production VRF",
      vrfRD: "64512:100",
      vrfDisplay: "Browser E2E Production VRF (64512:100)",
      prefix: "198.51.100.0/24",
      address: "198.51.100.10/24",
    };

    await step("DCIM Site creation through Vue", async () => {
      await createResourceViaUI({
        route: "/dcim/sites/",
        apiPath: "/api/dcim/sites/",
        expected: names.site,
        fields: [
          ["text", "Name", names.site],
          ["text", "Slug", names.siteSlug],
          ["select", "Status", "active"],
        ],
      });
    });

    await step("server validation is rendered in the Vue form", async () => {
      await navigate("/dcim/sites/add/");
      await waitForText("Add Site");
      await fillField("Name", "Browser E2E Duplicate");
      await fillField("Slug", names.siteSlug);
      await selectField("Status", "active");
      await clickText("Save", "button");
      await waitForText("site with this slug already exists.");
      const state = await locationState();
      assert(
        state.pathname === "/dcim/sites/add/",
        "validation failure unexpectedly left the form",
      );
    });

    await step(
      "limited-user direct edit/save receives RBAC 403 and rolls back",
      async () => {
        const siteID = await resourceID(
          "/api/dcim/sites/",
          "slug",
          names.siteSlug,
        );

        await navigate("/logout/");
        await waitForPath("/login");
        await waitForText("Sign In");
        await fillField("Username", limitedUsername);
        await fillField("Password", limitedPassword);
        await clickText("Sign In", "button");
        await waitForPath("/");
        const limitedSession = await apiJSON("/api/auth/session/");
        assert(
          limitedSession.user?.is_superuser === false,
          "limited browser principal unexpectedly has superuser status",
        );
        assert(
          JSON.stringify(limitedSession.permissions) ===
            JSON.stringify(["dcim.view_site"]),
          `limited browser principal has unexpected permissions: ${JSON.stringify(limitedSession.permissions)}`,
        );
        await assertCredentialFreeStorage();
        await navigate("/dcim/sites/");
        await waitForText(names.site);
        assert(
          !(await hasVisibleText("Add", "a, button")),
          "view-only user was offered the Site add action",
        );
        await navigate(`/dcim/sites/${siteID}/`);
        await waitForText(names.site);
        assert(
          !(await hasVisibleText("Edit", "a, button")),
          "view-only user was offered the Site edit action",
        );

        // The client-side guard is a usability boundary: a direct edit attempt
        // must render the application's 403 page for a view-only principal.
        await navigate(`/dcim/sites/${siteID}/edit/`);
        await waitForPath("/403/");
        await waitForText("Access Forbidden");

        // Independently prove the server denies the same mutation with a valid
        // limited-user session and CSRF token. This guards against a stale or
        // bypassed client permission cache without weakening the route guard.
        const mutationResult = await evaluate(
          async (id) => {
            const csrf = document.cookie
              .split(";")
              .map((part) => part.trim())
              .find((part) => part.startsWith("csrftoken="))
              ?.slice("csrftoken=".length);
            if (!csrf) return { csrf_present: false, status: 0 };
            const response = await fetch(`/api/dcim/sites/${id}/`, {
              method: "PATCH",
              credentials: "include",
              headers: {
                "Content-Type": "application/json",
                "X-CSRFToken": decodeURIComponent(csrf || ""),
              },
              body: JSON.stringify({ facility: "must-not-persist" }),
            });
            return { csrf_present: true, status: response.status };
          },
          [siteID],
        );
        assert(
          mutationResult.csrf_present,
          "limited-user session did not expose its CSRF cookie",
        );
        assert(
          mutationResult.status === 403,
          `limited-user mutation returned ${mutationResult.status}, expected 403`,
        );
        const site = await apiJSON(`/api/dcim/sites/${siteID}/`);
        assert(site.facility === "", "RBAC-denied mutation changed the Site");

        await navigate("/logout/");
        await waitForPath("/login");
        await waitForText("Sign In");
        await fillField("Username", username);
        await fillField("Password", password);
        await clickText("Sign In", "button");
        await waitForPath("/");
        await navigate("/dcim/sites/");
        await waitForText(names.site);
        await assertCredentialFreeStorage();
      },
    );

    await step(
      "DCIM manufacturer, Rack Role, and Rack Type creation through Vue",
      async () => {
        await createResourceViaUI({
          route: "/dcim/manufacturers/",
          apiPath: "/api/dcim/manufacturers/",
          expected: names.manufacturer,
          fields: [
            ["text", "Name", names.manufacturer],
            ["text", "Slug", names.manufacturerSlug],
          ],
        });
        await createResourceViaUI({
          route: "/dcim/rack-roles/",
          apiPath: "/api/dcim/rack-roles/",
          expected: names.rackRole,
          fields: [
            ["text", "Name", names.rackRole],
            ["text", "Slug", names.rackRoleSlug],
          ],
        });
        await createResourceViaUI({
          route: "/dcim/rack-types/",
          apiPath: "/api/dcim/rack-types/",
          expected: names.rackType,
          fields: [
            [
              "relation",
              "Manufacturer",
              names.manufacturer,
              names.manufacturer,
            ],
            ["text", "Model", names.rackType],
            ["text", "Slug", names.rackTypeSlug],
            ["select", "Form Factor", "4-post-cabinet"],
          ],
        });
      },
    );

    await step("DCIM Rack relationships are created through Vue", async () => {
      await createResourceViaUI({
        route: "/dcim/racks/",
        apiPath: "/api/dcim/racks/",
        expected: names.rack,
        fields: [
          ["relation", "Site", names.site, names.site],
          ["text", "Name", names.rack],
          ["relation", "Rack Type", names.rackType, names.rackType],
          ["relation", "Role", names.rackRole, names.rackRole],
          ["select", "Status", "active"],
        ],
      });
    });

    await step(
      "DCIM Device Type, role, and Interface Template creation through Vue",
      async () => {
        await createResourceViaUI({
          route: "/dcim/device-types/",
          apiPath: "/api/dcim/device-types/",
          expected: names.deviceType,
          fields: [
            [
              "relation",
              "Manufacturer",
              names.manufacturer,
              names.manufacturer,
            ],
            ["text", "Model", names.deviceType],
            ["text", "Slug", names.deviceTypeSlug],
          ],
        });
        await createResourceViaUI({
          route: "/dcim/device-roles/",
          apiPath: "/api/dcim/device-roles/",
          expected: names.deviceRole,
          fields: [
            ["text", "Name", names.deviceRole],
            ["text", "Slug", names.deviceRoleSlug],
          ],
        });
        await createResourceViaUI({
          route: "/dcim/interface-templates/",
          apiPath: "/api/dcim/interface-templates/",
          expected: names.interface,
          fields: [
            ["relation", "Device Type", names.deviceType, names.deviceType],
            ["text", "Name", names.interface],
            ["select", "Type", "1000base-t"],
          ],
        });
      },
    );

    let interfaceID;
    await step(
      "Device creation instantiates its Interface Template",
      async () => {
        await createResourceViaUI({
          route: "/dcim/devices/",
          apiPath: "/api/dcim/devices/",
          expected: names.device,
          fields: [
            ["relation", "Device Type", names.deviceType, names.deviceType],
            ["relation", "Device Role", names.deviceRole, names.deviceRole],
            ["text", "Name", names.device],
            ["relation", "Site", names.site, names.site],
            ["relation", "Rack", names.rack, names.rack],
            ["number", "Position (U)", 1],
            ["select", "Face", "front"],
            ["select", "Status", "active"],
          ],
        });
        const deviceID = await resourceID(
          "/api/dcim/devices/",
          "name",
          names.device,
        );
        const interfaces = await apiJSON(
          `/api/dcim/interfaces/?device_id=${deviceID}&q=${encodeURIComponent(names.interface)}&limit=100`,
        );
        const instantiated = interfaces.results?.find(
          (item) =>
            item.name === names.interface && item.device?.id === deviceID,
        );
        assert(
          Number.isInteger(instantiated?.id),
          "Device did not instantiate its Interface Template",
        );
        assert(
          instantiated.type?.value === "1000base-t",
          "instantiated Interface did not preserve template type",
        );
        interfaceID = instantiated.id;
        await navigate("/dcim/interfaces/");
        await waitForText(names.interface);
      },
    );

    await step(
      "IPAM VRF, Prefix, and address creation through Vue",
      async () => {
        await createResourceViaUI({
          route: "/ipam/vrfs/",
          apiPath: "/api/ipam/vrfs/",
          expected: names.vrf,
          fields: [
            ["text", "Name", names.vrf],
            ["text", "Route Distinguisher", names.vrfRD],
          ],
        });
        await createResourceViaUI({
          route: "/ipam/prefixes/",
          apiPath: "/api/ipam/prefixes/",
          expected: names.prefix,
          fields: [
            ["text", "Prefix", names.prefix],
            ["relation", "VRF", names.vrf, names.vrfDisplay],
            ["select", "Status", "active"],
          ],
        });
        await createResourceViaUI({
          route: "/ipam/ip-addresses/",
          apiPath: "/api/ipam/ip-addresses/",
          expected: names.address,
          fields: [
            ["text", "Address", names.address],
            ["relation", "VRF", names.vrf, names.vrfDisplay],
            ["select", "Status", "active"],
          ],
        });
      },
    );

    let addressID;
    await step(
      "IP address assignment and unassignment through Vue",
      async () => {
        addressID = await resourceID(
          "/api/ipam/ip-addresses/",
          "address",
          names.address,
        );
        await navigate(`/ipam/ip-addresses/${addressID}/`);
        await waitForText("Interface assignment");
        await waitForText("This IP address is not assigned to an Interface.");
        await clickText("Assign", "button");
        await selectRelation("Interface", names.interface, names.interface);
        await clickText("Save assignment", "button");
        await waitForText(`Assigned to ${names.interface}`);

        let address = await apiJSON(`/api/ipam/ip-addresses/${addressID}/`);
        assert(
          address.assigned_object?.display === names.interface,
          "assignment was not persisted",
        );
        assert(
          address.assigned_object_type === "dcim.interface",
          "assignment type was not persisted",
        );

        await clickText("Unassign", "button");
        await waitForText("Remove this IP address from its Interface?");
        await clickModalButton("Unassign IP address", "Unassign");
        await waitForText("This IP address is not assigned to an Interface.");

        address = await apiJSON(`/api/ipam/ip-addresses/${addressID}/`);
        assert(
          address.assigned_object === null,
          "unassignment did not clear assigned_object",
        );
        assert(
          address.assigned_object_type === null,
          "unassignment did not clear assigned_object_type",
        );
        assert(
          address.assigned_object_id === null,
          "unassignment did not clear assigned_object_id",
        );
      },
    );

    await step(
      "protected delete is surfaced and rolled back by Vue",
      async () => {
        const siteID = await resourceID(
          "/api/dcim/sites/",
          "slug",
          names.siteSlug,
        );
        await navigate(`/dcim/sites/${siteID}/delete/`);
        await waitForText(`Delete \"${names.site}\"?`);
        await clickText("Confirm Delete", "button");
        await waitForText(
          "Cannot delete this site because it is referenced by rack",
        );
        const site = await apiJSON(`/api/dcim/sites/${siteID}/`);
        assert(site.slug === names.siteSlug, "protected Site was deleted");
      },
    );

    await step(
      "delete cancellation preserves an object and confirmation deletes it",
      async () => {
        await createResourceViaUI({
          route: "/dcim/sites/",
          apiPath: "/api/dcim/sites/",
          expected: names.disposableSite,
          fields: [
            ["text", "Name", names.disposableSite],
            ["text", "Slug", names.disposableSiteSlug],
            ["select", "Status", "active"],
          ],
        });
        const disposableID = await resourceID(
          "/api/dcim/sites/",
          "slug",
          names.disposableSiteSlug,
        );
        await navigate(`/dcim/sites/${disposableID}/delete/`);
        await waitForText(`Delete \"${names.disposableSite}\"?`);
        await clickText("Cancel", "button");
        await waitForPath(`/dcim/sites/${disposableID}/`);
        assert(
          (await apiJSON(`/api/dcim/sites/${disposableID}/`)).slug ===
            names.disposableSiteSlug,
          "delete cancellation did not preserve the Site",
        );

        await navigate(`/dcim/sites/${disposableID}/delete/`);
        await waitForText(`Delete \"${names.disposableSite}\"?`);
        await clickText("Confirm Delete", "button");
        await waitForPath("/dcim/sites/");
        assert(
          (await apiStatus(`/api/dcim/sites/${disposableID}/`)) === 404,
          "confirmed Site delete did not persist",
        );
      },
    );

    await step(
      "Interface delete warning supports cancel and assigned-IP cascade confirmation",
      async () => {
        await navigate(`/ipam/ip-addresses/${addressID}/`);
        await waitForText("Interface assignment");
        await clickText("Assign", "button");
        await selectRelation("Interface", names.interface, names.interface);
        await clickText("Save assignment", "button");
        await waitForText(`Assigned to ${names.interface}`);

        await navigate(`/dcim/interfaces/${interfaceID}/delete/`);
        await waitForNormalizedText(
          "Deleting this Interface also deletes 1 assigned IP address.",
        );
        await clickText("Cancel", "button");
        await waitForPath(`/dcim/interfaces/${interfaceID}/`);
        assert(
          (await apiStatus(`/api/dcim/interfaces/${interfaceID}/`)) === 200,
          "cancel deleted the Interface",
        );
        assert(
          (await apiStatus(`/api/ipam/ip-addresses/${addressID}/`)) === 200,
          "cancel deleted the assigned IP",
        );

        await navigate(`/dcim/interfaces/${interfaceID}/delete/`);
        await waitForNormalizedText(
          "Deleting this Interface also deletes 1 assigned IP address.",
        );
        await clickText("Confirm Delete", "button");
        await waitForPath("/dcim/interfaces/");
        assert(
          (await apiStatus(`/api/dcim/interfaces/${interfaceID}/`)) === 404,
          "confirmed Interface delete did not persist",
        );
        assert(
          (await apiStatus(`/api/ipam/ip-addresses/${addressID}/`)) === 404,
          "assigned IP was not cascade-deleted",
        );
      },
    );

    await step("logout revokes UI access and REST access", async () => {
      await navigate("/logout/");
      await waitForPath("/login");
      await navigate(`/ipam/ip-addresses/${addressID}/`);
      await waitForPath("/login");
      const status = await evaluate(
        async (url) => (await fetch(url, { credentials: "include" })).status,
        [`/api/ipam/ip-addresses/${addressID}/`],
      );
      assert(
        status === 401,
        `logged-out REST request returned ${status}, expected 401`,
      );
      await assertCredentialFreeStorage();
    });

    succeeded = true;
    diagnostics.finished_at = new Date().toISOString();
    diagnostics.result = "passed";
    writeJSON("summary.json", diagnostics);
    process.stdout.write(
      `Browser E2E passed: ${diagnostics.steps.length} real-Chrome workflow assertions; artifacts: ${artifactDir}\n`,
    );
  } catch (error) {
    diagnostics.finished_at = new Date().toISOString();
    diagnostics.result = "failed";
    diagnostics.error = redact(
      error instanceof Error ? (error.stack ?? error.message) : String(error),
    );
    await retainFailureDiagnostics();
    writeJSON("summary.json", diagnostics);
    process.stderr.write(
      `Browser E2E failed: ${diagnostics.error}\nArtifacts: ${artifactDir}\n`,
    );
    process.exitCode = 1;
  } finally {
    if (client) client.close();
    if (succeeded) {
      // A passing run keeps only its compact, credential-free evidence file.
      for (const name of ["failure.png", "failure.html"]) {
        fs.rmSync(path.join(artifactDir, name), { force: true });
      }
    }
  }
}

async function step(name, operation) {
  const started = Date.now();
  await operation();
  diagnostics.steps.push({
    name,
    duration_ms: Date.now() - started,
    status: "passed",
  });
  process.stdout.write(`browser e2e: ${name}\n`);
}

async function createResourceViaUI({ route, apiPath, expected, fields }) {
  await navigate(`${route}add/`);
  await waitFor("resource add form", async () =>
    (await bodyText()).includes("Save"),
  );
  for (const field of fields) {
    const [type, label, value, expectedOption] = field;
    if (type === "text" || type === "number") await fillField(label, value);
    else if (type === "select") await selectField(label, value);
    else if (type === "relation")
      await selectRelation(label, value, expectedOption);
    else throw new Error(`unsupported form operation ${type}`);
  }
  await clickText("Save", "button");
  await waitForPath(route);
  await waitForText(expected);
  const response = await apiJSON(
    `${apiPath}?q=${encodeURIComponent(expected)}&limit=100`,
  );
  assert(
    response.results?.some((item) => Object.values(item).includes(expected)),
    `${expected} not persisted`,
  );
}

async function navigate(relativePath) {
  const url = new URL(relativePath, baseURL).href;
  await client.command("Page.navigate", { url });
  await waitFor(
    "document ready",
    async () => (await evaluate(() => document.readyState)) === "complete",
  );
}

async function fillField(label, value) {
  const result = await evaluate(
    (fieldLabel, nextValue) => {
      const normalize = (value) =>
        value
          .replace(/\s+/g, " ")
          .trim()
          .replace(/\s*\*$/, "");
      const labelNode = [...document.querySelectorAll("label")].find(
        (candidate) => normalize(candidate.textContent || "") === fieldLabel,
      );
      const control =
        labelNode?.parentElement?.querySelector("input, textarea");
      if (!(
        control instanceof HTMLInputElement ||
        control instanceof HTMLTextAreaElement
      ))
        return false;
      control.focus();
      const prototype =
        control instanceof HTMLTextAreaElement
          ? HTMLTextAreaElement.prototype
          : HTMLInputElement.prototype;
      Object.getOwnPropertyDescriptor(prototype, "value").set.call(
        control,
        nextValue,
      );
      control.dispatchEvent(new Event("input", { bubbles: true }));
      control.dispatchEvent(new Event("change", { bubbles: true }));
      return true;
    },
    [label, String(value)],
  );
  assert(result, `field ${label} was not found`);
}

async function selectField(label, value) {
  const result = await evaluate(
    (fieldLabel, nextValue) => {
      const normalize = (value) =>
        value
          .replace(/\s+/g, " ")
          .trim()
          .replace(/\s*\*$/, "");
      const labelNode = [...document.querySelectorAll("label")].find(
        (candidate) => normalize(candidate.textContent || "") === fieldLabel,
      );
      const control = labelNode?.parentElement?.querySelector("select");
      if (!(control instanceof HTMLSelectElement)) return false;
      const option = [...control.options].find(
        (candidate) => candidate.value === String(nextValue),
      );
      if (!option) return false;
      control.value = option.value;
      control.dispatchEvent(new Event("change", { bubbles: true }));
      return true;
    },
    [label, String(value)],
  );
  assert(result, `select ${label} did not contain ${value}`);
}

async function selectRelation(label, query, expectedOption) {
  await fillField(label, query);
  await evaluate(
    (fieldLabel) => {
      const normalize = (value) =>
        value
          .replace(/\s+/g, " ")
          .trim()
          .replace(/\s*\*$/, "");
      const labelNode = [...document.querySelectorAll("label")].find(
        (candidate) => normalize(candidate.textContent || "") === fieldLabel,
      );
      const input = labelNode?.parentElement?.querySelector("input");
      input?.dispatchEvent(new FocusEvent("focus", { bubbles: true }));
    },
    [label],
  );
  await waitFor(`relationship option ${expectedOption}`, async () =>
    evaluate(
      (expected) => {
        const normalize = (value) => value.replace(/\s+/g, " ").trim();
        return [...document.querySelectorAll("[role=listbox] button")].some(
          (button) => normalize(button.textContent || "") === expected,
        );
      },
      [expectedOption],
    ),
  );
  const selected = await evaluate(
    (expected) => {
      const normalize = (value) => value.replace(/\s+/g, " ").trim();
      const button = [
        ...document.querySelectorAll("[role=listbox] button"),
      ].find(
        (candidate) => normalize(candidate.textContent || "") === expected,
      );
      if (!(button instanceof HTMLElement)) return false;
      button.dispatchEvent(
        new MouseEvent("mousedown", {
          bubbles: true,
          cancelable: true,
          view: window,
        }),
      );
      return true;
    },
    [expectedOption],
  );
  assert(
    selected,
    `relationship option ${expectedOption} could not be selected`,
  );
  await waitFor(`relationship ${label} selection`, async () =>
    evaluate(
      (fieldLabel, expected) => {
        const normalize = (value) =>
          value
            .replace(/\s+/g, " ")
            .trim()
            .replace(/\s*\*$/, "");
        const labelNode = [...document.querySelectorAll("label")].find(
          (candidate) => normalize(candidate.textContent || "") === fieldLabel,
        );
        return (
          labelNode?.parentElement?.querySelector("input")?.value === expected
        );
      },
      [label, expectedOption],
    ),
  );
}

async function clickText(text, selector = "button, a") {
  const clicked = await evaluate(
    (expected, query) => {
      const normalize = (value) => value.replace(/\s+/g, " ").trim();
      const element = [...document.querySelectorAll(query)].find(
        (candidate) =>
          normalize(candidate.textContent || "") === expected &&
          candidate.getClientRects().length > 0,
      );
      if (!(element instanceof HTMLElement)) return false;
      element.click();
      return true;
    },
    [text, selector],
  );
  assert(clicked, `visible ${selector} with text ${text} was not found`);
}

async function hasVisibleText(text, selector = "button, a") {
  return evaluate(
    (expected, query) => {
      const normalize = (value) => value.replace(/\s+/g, " ").trim();
      return [...document.querySelectorAll(query)].some(
        (candidate) =>
          normalize(candidate.textContent || "") === expected &&
          candidate.getClientRects().length > 0,
      );
    },
    [text, selector],
  );
}

async function clickModalButton(title, text) {
  const clicked = await evaluate(
    (modalTitle, buttonText) => {
      const normalize = (value) => value.replace(/\s+/g, " ").trim();
      const heading = [...document.querySelectorAll("h3")].find(
        (candidate) => normalize(candidate.textContent || "") === modalTitle,
      );
      const modal = heading?.parentElement?.parentElement;
      const button =
        modal &&
        [...modal.querySelectorAll("button")].find(
          (candidate) => normalize(candidate.textContent || "") === buttonText,
        );
      if (!(button instanceof HTMLElement)) return false;
      button.click();
      return true;
    },
    [title, text],
  );
  assert(clicked, `modal ${title} button ${text} was not found`);
}

async function resourceID(apiPath, key, expected) {
  const query = new URLSearchParams([
    [key, String(expected)],
    ["limit", "100"],
  ]);
  const response = await apiJSON(`${apiPath}?${query}`);
  const resource = response.results?.find((item) => item[key] === expected);
  assert(
    Number.isInteger(resource?.id),
    `${key}=${expected} was not found at ${apiPath}`,
  );
  return resource.id;
}

async function apiJSON(relativePath) {
  const response = await evaluate(
    async (url) => {
      const result = await fetch(url, {
        credentials: "include",
        headers: { Accept: "application/json" },
      });
      const body = await result.json().catch(() => null);
      return { status: result.status, body };
    },
    [relativePath],
  );
  assert(
    response.status >= 200 && response.status < 300,
    `${relativePath} returned HTTP ${response.status}`,
  );
  return response.body;
}

async function apiStatus(relativePath) {
  return evaluate(
    async (url) => (await fetch(url, { credentials: "include" })).status,
    [relativePath],
  );
}

async function assertCredentialFreeStorage() {
  const storage = await evaluate(() => ({
    local: Object.fromEntries(Object.entries(localStorage)),
    session: Object.fromEntries(Object.entries(sessionStorage)),
  }));
  const allowedLocal = new Set(["netbox_sidebar", "netbox_theme"]);
  const disallowedPattern =
    /(token|pass|secret|credential|session|bearer|auth)/i;
  for (const [key, value] of Object.entries(storage.local)) {
    assert(allowedLocal.has(key), `unexpected localStorage key ${key}`);
    assert(
      !disallowedPattern.test(`${key}:${value}`),
      `credential-like localStorage entry ${key}`,
    );
  }
  assert(
    Object.keys(storage.session).length === 0,
    "sessionStorage must remain empty",
  );
}

async function waitForPath(expectedPath) {
  await waitFor(
    `path ${expectedPath}`,
    async () => (await locationState()).pathname === expectedPath,
  );
}

async function waitForText(expected) {
  await waitFor(`text ${expected}`, async () =>
    (await bodyText()).includes(expected),
  );
}

async function waitForNormalizedText(expected) {
  const normalize = (value) => value.replace(/\s+/g, " ").trim();
  await waitFor(`text ${expected}`, async () =>
    normalize(await bodyText()).includes(expected),
  );
}

async function waitFor(description, predicate, limit = timeoutMs) {
  const deadline = Date.now() + limit;
  let lastError;
  while (Date.now() < deadline) {
    try {
      if (await predicate()) return;
    } catch (error) {
      lastError = error;
    }
    await delay(100);
  }
  const state = client
    ? await locationState().catch(() => ({ pathname: "unknown", search: "" }))
    : null;
  throw new Error(
    `timed out waiting for ${description}${state ? ` at ${state.pathname}${state.search}` : ""}${
      lastError ? `: ${lastError.message}` : ""
    }`,
  );
}

async function bodyText() {
  return evaluate(() => document.body?.innerText ?? "");
}

async function locationState() {
  return evaluate(() => ({
    pathname: location.pathname,
    search: location.search,
    href: location.href,
  }));
}

async function evaluate(fn, args = []) {
  const expression = `(${fn.toString()})(...${JSON.stringify(args)})`;
  const response = await client.command("Runtime.evaluate", {
    expression,
    awaitPromise: true,
    returnByValue: true,
    userGesture: true,
  });
  if (response.exceptionDetails) {
    throw new Error(
      response.exceptionDetails.exception?.description ??
        response.exceptionDetails.text,
    );
  }
  return response.result?.value;
}

async function retainFailureDiagnostics() {
  if (!client) return;
  try {
    await evaluate(
      (secrets) => {
        for (const input of document.querySelectorAll("input[type=password]"))
          input.value = "";
        for (const secret of secrets) {
          document.body.innerHTML = document.body.innerHTML
            .split(secret)
            .join("[REDACTED]");
        }
      },
      [knownSecrets],
    );
  } catch {}
  try {
    const screenshot = await client.command("Page.captureScreenshot", {
      format: "png",
      captureBeyondViewport: true,
    });
    fs.writeFileSync(
      path.join(artifactDir, "failure.png"),
      Buffer.from(screenshot.data, "base64"),
      { mode: 0o600 },
    );
  } catch {}
  try {
    const html = await evaluate(() => document.documentElement.outerHTML);
    fs.writeFileSync(path.join(artifactDir, "failure.html"), redact(html), {
      mode: 0o600,
    });
  } catch {}
}

function attachDiagnostics(cdp, target) {
  const requests = new Map();
  cdp.on("Runtime.consoleAPICalled", (params) => {
    const values = params.args
      .map((argument) => argument.value ?? argument.description ?? "")
      .join(" ");
    target.console.push({
      type: params.type,
      text: redact(values).slice(0, 1000),
    });
  });
  cdp.on("Network.requestWillBeSent", (params) => {
    requests.set(params.requestId, {
      method: params.request.method,
      url: safeURL(params.request.url),
    });
  });
  cdp.on("Network.responseReceived", (params) => {
    if (params.response.status >= 400) {
      const request = requests.get(params.requestId) ?? {
        method: "UNKNOWN",
        url: safeURL(params.response.url),
      };
      target.network_failures.push({
        ...request,
        status: params.response.status,
      });
    }
  });
  cdp.on("Network.loadingFailed", (params) => {
    const request = requests.get(params.requestId) ?? {
      method: "UNKNOWN",
      url: "unknown",
    };
    if (!params.canceled)
      target.network_failures.push({
        ...request,
        error: redact(params.errorText),
      });
  });
}

class CDPClient {
  constructor(socket) {
    this.socket = socket;
    this.nextID = 1;
    this.pending = new Map();
    this.listeners = new Map();
    socket.addEventListener("message", (event) =>
      this.handleMessage(event.data),
    );
    socket.addEventListener("close", () =>
      this.rejectPending(new Error("Chrome DevTools connection closed")),
    );
    socket.addEventListener("error", () =>
      this.rejectPending(new Error("Chrome DevTools connection failed")),
    );
  }

  static async connect(webSocketURL) {
    const socket = new WebSocket(webSocketURL);
    await new Promise((resolve, reject) => {
      const timer = setTimeout(
        () => reject(new Error("timed out connecting to Chrome DevTools")),
        10000,
      );
      socket.addEventListener(
        "open",
        () => {
          clearTimeout(timer);
          resolve();
        },
        { once: true },
      );
      socket.addEventListener(
        "error",
        () => {
          clearTimeout(timer);
          reject(new Error("could not connect to Chrome DevTools"));
        },
        { once: true },
      );
    });
    return new CDPClient(socket);
  }

  command(method, params = {}) {
    const id = this.nextID++;
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pending.delete(id);
        reject(new Error(`Chrome DevTools command ${method} timed out`));
      }, 30000);
      this.pending.set(id, { resolve, reject, timer });
      this.socket.send(JSON.stringify({ id, method, params }));
    });
  }

  on(method, listener) {
    const current = this.listeners.get(method) ?? [];
    current.push(listener);
    this.listeners.set(method, current);
  }

  handleMessage(raw) {
    const message = JSON.parse(raw);
    if (message.id) {
      const pending = this.pending.get(message.id);
      if (!pending) return;
      clearTimeout(pending.timer);
      this.pending.delete(message.id);
      if (message.error)
        pending.reject(
          new Error(`${message.error.message} (${message.error.code})`),
        );
      else pending.resolve(message.result);
      return;
    }
    for (const listener of this.listeners.get(message.method) ?? [])
      listener(message.params);
  }

  rejectPending(error) {
    for (const pending of this.pending.values()) {
      clearTimeout(pending.timer);
      pending.reject(error);
    }
    this.pending.clear();
  }

  close() {
    this.socket.close();
  }
}

async function createPage(port) {
  const endpoint = `http://127.0.0.1:${port}/json/new?${encodeURIComponent("about:blank")}`;
  const deadline = Date.now() + 10000;
  let lastError;
  while (Date.now() < deadline) {
    try {
      const response = await fetch(endpoint, { method: "PUT" });
      if (response.ok) return response.json();
      lastError = new Error(`DevTools endpoint returned ${response.status}`);
    } catch (error) {
      lastError = error;
    }
    await delay(100);
  }
  throw new Error(
    `Chrome DevTools endpoint unavailable: ${lastError?.message ?? "unknown error"}`,
  );
}

function required(name) {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
}

function requiredURL(name) {
  const url = new URL(required(name));
  if (!["http:", "https:"].includes(url.protocol))
    throw new Error(`${name} must be HTTP(S)`);
  url.pathname = "/";
  url.search = "";
  url.hash = "";
  return url;
}

function positiveInteger(value, name) {
  const number = Number(value);
  if (!Number.isInteger(number) || number <= 0)
    throw new Error(`${name} must be a positive integer`);
  return number;
}

function safeURL(raw) {
  try {
    const url = new URL(raw);
    return `${url.origin}${url.pathname}`;
  } catch {
    return redact(raw);
  }
}

function redact(value) {
  let redacted = String(value);
  for (const secret of knownSecrets) {
    redacted = redacted.split(secret).join("[REDACTED]");
  }
  return redacted.replace(
    /(authorization|cookie|token|secret|password)(["'\s:=]+)[^\s"'<]+/gi,
    "$1$2[REDACTED]",
  );
}

function writeJSON(name, value) {
  fs.writeFileSync(
    path.join(artifactDir, name),
    `${JSON.stringify(value, null, 2)}\n`,
    { mode: 0o600 },
  );
}

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

function delay(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

await runBrowserE2E();
