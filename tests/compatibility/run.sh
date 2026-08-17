#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="$ROOT/tests/compatibility/compose.yaml"
ORACLE_SOURCE_MATERIALIZER="$ROOT/tests/compatibility/materialize_oracle_source.sh"
PINNED_SHA="fbb948d30e79ce657fac62994a22aca72c1770a9"
PINNED_RELEASE="4.4.6"
PROFILE_BASELINE="netbox-v4.4.6-post7"
PROJECT="netbox-compat-${UID}-$$"
BUILD_DIR="$(mktemp -d "/tmp/${PROJECT}-build.XXXXXX")"
ARTIFACT_DIR="${NETBOX_COMPAT_ARTIFACT_DIR:-/tmp/${PROJECT}-artifacts}"
GO_PID=""

mkdir -p "$ARTIFACT_DIR"

free_port() {
  node -e 'const net=require("node:net"); const server=net.createServer(); server.listen(0,"127.0.0.1",()=>{console.log(server.address().port); server.close();});'
}

export NETBOX_ORACLE_HTTP_PORT="${NETBOX_ORACLE_HTTP_PORT:-$(free_port)}"
export NETBOX_GO_POSTGRES_PORT="${NETBOX_GO_POSTGRES_PORT:-$(free_port)}"
NETBOX_GO_HTTP_PORT="${NETBOX_GO_HTTP_PORT:-$(free_port)}"
NETBOX_GO_GRPC_PORT="${NETBOX_GO_GRPC_PORT:-$(free_port)}"
export COMPOSE_PROJECT_NAME="$PROJECT"

cleanup() {
  local status=$?
  trap - EXIT
  set +e
  if [[ -n "$GO_PID" ]]; then
    kill "$GO_PID" >/dev/null 2>&1
    wait "$GO_PID" >/dev/null 2>&1
  fi
  if (( status != 0 )); then
    docker compose -f "$COMPOSE_FILE" logs --no-color >"$ARTIFACT_DIR/compose.log" 2>&1
  fi
  docker compose -f "$COMPOSE_FILE" down --volumes --remove-orphans >/dev/null 2>&1
  rm -rf "$BUILD_DIR"
  if (( status != 0 )); then
    printf 'compatibility oracle failed; diagnostics: %s\n' "$ARTIFACT_DIR" >&2
  else
    printf 'compatibility oracle artifacts: %s\n' "$ARTIFACT_DIR"
  fi
  exit "$status"
}
trap cleanup EXIT

for command in docker git go node curl; do
  if ! command -v "$command" >/dev/null 2>&1; then
    printf 'required command is unavailable: %s\n' "$command" >&2
    exit 1
  fi
done

"$ORACLE_SOURCE_MATERIALIZER" \
  "$ROOT/netbox" \
  "$PINNED_SHA" \
  netbox \
  "$BUILD_DIR/oracle"
export NETBOX_ORACLE_SOURCE_DIR="$BUILD_DIR/oracle/netbox"
if ! grep -qx "version: \"$PINNED_RELEASE\"" "$NETBOX_ORACLE_SOURCE_DIR/release.yaml"; then
  printf 'oracle release metadata is not NetBox %s\n' "$PINNED_RELEASE" >&2
  exit 1
fi
node -e '
  const fs = require("node:fs");
  const profile = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
  if (profile.compatibility_baseline !== process.argv[2]) {
    throw new Error(`profile baseline mismatch: ${profile.compatibility_baseline}`);
  }
' "$ROOT/contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml" "$PROFILE_BASELINE"

node "$ROOT/tests/compatibility/comparator_self_test.mjs"

for image in postgres:16-alpine redis:7-alpine netboxcommunity/netbox:v4.4-3.4.1; do
  if ! docker image inspect "$image" >/dev/null 2>&1; then
    printf 'required cached image is unavailable (the job never pulls): %s\n' "$image" >&2
    exit 1
  fi
done

printf 'starting pinned NetBox oracle and disposable PostgreSQL services\n'
docker compose -f "$COMPOSE_FILE" up -d --wait --wait-timeout 300 oracle go-postgres

ORACLE_CONFIG="$({
  docker compose -f "$COMPOSE_FILE" exec -T oracle \
    /opt/netbox/venv/bin/python /opt/netbox/netbox/manage.py shell -c \
    'import json; from django.conf import settings; print("COMPAT_CONFIG=" + json.dumps({"version": settings.VERSION, "login_required": settings.LOGIN_REQUIRED, "allow_token_retrieval": settings.ALLOW_TOKEN_RETRIEVAL, "maintenance_mode": settings.MAINTENANCE_MODE, "enforce_global_unique": settings.ENFORCE_GLOBAL_UNIQUE, "time_zone": settings.TIME_ZONE}, sort_keys=True))'
} 2>&1)"
printf '%s\n' "$ORACLE_CONFIG" >"$ARTIFACT_DIR/oracle-config.txt"
NETBOX_EFFECTIVE_ORACLE_CONFIG="$ORACLE_CONFIG" node -e '
  const line = process.env.NETBOX_EFFECTIVE_ORACLE_CONFIG.split(/\r?\n/).find((entry) => entry.startsWith("COMPAT_CONFIG="));
  if (!line) throw new Error("oracle did not report its effective configuration");
  const actual = JSON.parse(line.slice("COMPAT_CONFIG=".length));
  const expected = {
    allow_token_retrieval: false,
    enforce_global_unique: true,
    login_required: true,
    maintenance_mode: false,
    time_zone: "UTC",
    version: "4.4.6",
  };
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`effective oracle configuration mismatch: ${JSON.stringify(actual)}`);
  }
'

printf 'building standalone Go server and administrator CLI from the workspace\n'
(
  cd "$ROOT/netbox-backend"
  env CGO_ENABLED=0 GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
    go build -trimpath -o "$BUILD_DIR/netbox-go" ./cmd/netbox_go
  env CGO_ENABLED=0 GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
    go build -trimpath -o "$BUILD_DIR/netbox-go-admin" ./cmd/netbox_go_admin
)

GO_DSN="netbox_compat:netbox_compat@127.0.0.1:${NETBOX_GO_POSTGRES_PORT}/netbox_compat?sslmode=disable"
GO_USERNAME="compat-admin"
GO_PASSWORD="Compatibility-Only-2026!"

printf 'bootstrapping a local Go administrator through the offline CLI\n'
printf '%s\n%s\n' "$GO_PASSWORD" "$GO_PASSWORD" | env \
  NETBOX_DATABASE_DSN="$GO_DSN" \
  NETBOX_APP_ENV=test \
  "$BUILD_DIR/netbox-go-admin" bootstrap \
    --config "$ROOT/netbox-backend/configs/netbox_go.yml" \
    --username "$GO_USERNAME" \
    --email compat-admin@example.test \
    >"$ARTIFACT_DIR/go-admin.log" 2>&1

env \
  NETBOX_DATABASE_DSN="$GO_DSN" \
  NETBOX_APP_ENV=test \
  NETBOX_HTTP_PORT="$NETBOX_GO_HTTP_PORT" \
  NETBOX_GRPC_PORT="$NETBOX_GO_GRPC_PORT" \
  "$BUILD_DIR/netbox-go" -c "$ROOT/netbox-backend/configs/netbox_go.yml" \
  >"$ARTIFACT_DIR/go-server.log" 2>&1 &
GO_PID=$!

ready=false
for _ in $(seq 1 180); do
  if curl --fail --silent --show-error "http://127.0.0.1:${NETBOX_GO_HTTP_PORT}/health" >/dev/null 2>&1; then
    ready=true
    break
  fi
  if ! kill -0 "$GO_PID" >/dev/null 2>&1; then
    printf 'Go service exited before becoming ready\n' >&2
    tail -n 80 "$ARTIFACT_DIR/go-server.log" >&2
    exit 1
  fi
  sleep 1
done
if [[ "$ready" != true ]]; then
  printf 'Go service did not become ready within 180 seconds\n' >&2
  exit 1
fi

NETBOX_ORACLE_URL="http://127.0.0.1:${NETBOX_ORACLE_HTTP_PORT}" \
NETBOX_GO_URL="http://127.0.0.1:${NETBOX_GO_HTTP_PORT}" \
NETBOX_ORACLE_TOKEN="0123456789abcdef0123456789abcdef01234567" \
NETBOX_GO_USERNAME="$GO_USERNAME" \
NETBOX_GO_PASSWORD="$GO_PASSWORD" \
NETBOX_COMPAT_ARTIFACT_DIR="$ARTIFACT_DIR" \
  node "$ROOT/tests/compatibility/run.mjs"

printf 'checking committed profile state and append-only change logs\n'
docker compose -f "$COMPOSE_FILE" exec -T go-postgres \
  psql -X -v ON_ERROR_STOP=1 -U netbox_compat -d netbox_compat -Atc "
    SELECT json_build_object(
      'profile_rows',
        (SELECT count(*) FROM go_dcim_sites) +
        (SELECT count(*) FROM go_dcim_manufacturers) +
        (SELECT count(*) FROM go_dcim_rack_roles) +
        (SELECT count(*) FROM go_dcim_rack_types) +
        (SELECT count(*) FROM go_dcim_racks) +
        (SELECT count(*) FROM go_dcim_device_roles) +
        (SELECT count(*) FROM go_dcim_device_types) +
        (SELECT count(*) FROM go_dcim_interface_templates) +
        (SELECT count(*) FROM go_dcim_devices) +
        (SELECT count(*) FROM go_dcim_interfaces) +
        (SELECT count(*) FROM go_ipam_vrfs) +
        (SELECT count(*) FROM go_ipam_prefixes) +
        (SELECT count(*) FROM go_ipam_ip_addresses),
      'changes', (SELECT count(*) FROM go_object_changes),
      'invalid_action_changes', (
        SELECT count(*) FROM go_object_changes WHERE action NOT IN ('create', 'update', 'delete')
      ),
      'malformed_changes', (
        SELECT count(*) FROM go_object_changes
        WHERE (action = 'create' AND (before_data IS NOT NULL OR after_data IS NULL))
           OR (action = 'update' AND (before_data IS NULL OR after_data IS NULL))
           OR (action = 'delete' AND (before_data IS NULL OR after_data IS NOT NULL))
      ),
      'out_of_order_changes', (
        SELECT count(*) FROM (
          SELECT occurred_at, lag(occurred_at) OVER (ORDER BY id) AS previous_at
          FROM go_object_changes
        ) ordered_changes
        WHERE previous_at > occurred_at
      ),
      'failed_mutation_changes', (
        SELECT count(*) FROM go_object_changes
        WHERE coalesce(after_data->>'name', before_data->>'name', '') IN (
          'compatibility-device-occupied', 'compatibility-device-bounds'
        )
        OR coalesce(after_data->>'address', before_data->>'address', '') IN (
          '0.0.0.1/0', '198.51.100.0/24', '198.51.100.255/24', '2001:db8:100::/64',
          '198.51.100.30/32'
        )
      )
    );
  " >"$ARTIFACT_DIR/go-durable-state.json"

docker compose -f "$COMPOSE_FILE" exec -T oracle-postgres \
  psql -X -v ON_ERROR_STOP=1 -U netbox_oracle -d netbox_oracle -Atc "
    SELECT json_build_object(
      'profile_rows',
        (SELECT count(*) FROM dcim_site) +
        (SELECT count(*) FROM dcim_manufacturer) +
        (SELECT count(*) FROM dcim_rackrole) +
        (SELECT count(*) FROM dcim_racktype) +
        (SELECT count(*) FROM dcim_rack) +
        (SELECT count(*) FROM dcim_devicerole) +
        (SELECT count(*) FROM dcim_devicetype) +
        (SELECT count(*) FROM dcim_interfacetemplate) +
        (SELECT count(*) FROM dcim_device) +
        (SELECT count(*) FROM dcim_interface) +
        (SELECT count(*) FROM ipam_vrf) +
        (SELECT count(*) FROM ipam_prefix) +
        (SELECT count(*) FROM ipam_ipaddress),
      'changes', (SELECT count(*) FROM core_objectchange),
      'invalid_action_changes', (
        SELECT count(*) FROM core_objectchange WHERE action NOT IN ('create', 'update', 'delete')
      ),
      'malformed_changes', (
        SELECT count(*) FROM core_objectchange
        WHERE (action = 'create' AND (prechange_data IS NOT NULL OR postchange_data IS NULL))
           OR (action = 'update' AND (prechange_data IS NULL OR postchange_data IS NULL))
           OR (action = 'delete' AND (prechange_data IS NULL OR postchange_data IS NOT NULL))
      )
    );
  " >"$ARTIFACT_DIR/oracle-durable-state.json"

node -e '
  const fs = require("node:fs");
  for (const name of ["go", "oracle"]) {
    const state = JSON.parse(fs.readFileSync(`${process.argv[1]}/${name}-durable-state.json`, "utf8"));
    if (state.profile_rows !== 0) throw new Error(`${name}: cleanup left ${state.profile_rows} profile rows`);
    if (state.changes < 1) throw new Error(`${name}: mutations wrote no change records`);
    for (const field of ["invalid_action_changes", "malformed_changes"]) {
      if (state[field] !== 0) throw new Error(`${name}: ${field}=${state[field]}`);
    }
    if (name === "go" && (state.out_of_order_changes !== 0 || state.failed_mutation_changes !== 0)) {
      throw new Error(`Go durable-state invariant failed: ${JSON.stringify(state)}`);
    }
  }
' "$ARTIFACT_DIR"
