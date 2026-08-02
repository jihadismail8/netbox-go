#!/usr/bin/env bash

set -Eeuo pipefail

readonly REPOSITORY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
readonly PROJECT_NAME="netbox-go-smoke-$(date +%s)-$$"
readonly HEALTH_TIMEOUT_SECONDS="${NETBOX_SMOKE_TIMEOUT_SECONDS:-300}"
readonly RUN_BROWSER_E2E="${NETBOX_SMOKE_RUN_BROWSER_E2E:-0}"
readonly BUILD_MODE="${NETBOX_SMOKE_BUILD_MODE:-offline}"
readonly ARTIFACT_RELATIVE="tests/deployment/.artifacts/${PROJECT_NAME}"
readonly ARTIFACT_DIR="${REPOSITORY_ROOT}/${ARTIFACT_RELATIVE}"

case "${BUILD_MODE}" in
  offline)
    readonly COMPOSE_FILE="${REPOSITORY_ROOT}/tests/deployment/compose.smoke.yaml"
    ;;
  production)
    readonly COMPOSE_FILE="${REPOSITORY_ROOT}/docker-compose.yml"
    ;;
  *)
    echo "NETBOX_SMOKE_BUILD_MODE must be offline or production" >&2
    exit 2
    ;;
esac

if ! [[ "${HEALTH_TIMEOUT_SECONDS}" =~ ^[1-9][0-9]*$ ]]; then
  echo "NETBOX_SMOKE_TIMEOUT_SECONDS must be a positive integer" >&2
  exit 2
fi
if [[ "${RUN_BROWSER_E2E}" != "0" && "${RUN_BROWSER_E2E}" != "1" ]]; then
  echo "NETBOX_SMOKE_RUN_BROWSER_E2E must be 0 or 1" >&2
  exit 2
fi

# The smoke test performs all assertions from inside the Compose network. Use a
# random high host-port block solely to avoid colliding with a developer stack.
readonly PORT_BASE=$((20000 + (RANDOM % 30000)))
export NETBOX_POSTGRES_BIND=127.0.0.1
export NETBOX_POSTGRES_PORT="${PORT_BASE}"
export NETBOX_HTTP_BIND=127.0.0.1
export NETBOX_HTTP_PORT="$((PORT_BASE + 1))"
export NETBOX_GRPC_BIND=127.0.0.1
export NETBOX_GRPC_PORT="$((PORT_BASE + 2))"
export NETBOX_SMOKE_REPOSITORY_ROOT="${REPOSITORY_ROOT}"
export NETBOX_SMOKE_ARTIFACT_DIR="${ARTIFACT_RELATIVE}"
export NETBOX_SMOKE_IMAGE_TAG="${PROJECT_NAME}"

readonly -a COMPOSE=(
  docker compose
  --project-name "${PROJECT_NAME}"
  --file "${COMPOSE_FILE}"
)

cleanup() {
  local status=$?
  trap - EXIT INT TERM

  if ((status != 0)); then
    echo "Compose smoke test failed; service state and logs follow." >&2
    "${COMPOSE[@]}" ps >&2 || true
    "${COMPOSE[@]}" logs --no-color >&2 || true
  fi

  "${COMPOSE[@]}" down --volumes --remove-orphans >/dev/null 2>&1 || true
  if [[ "${BUILD_MODE}" == "offline" ]]; then
    docker image rm --force "netbox-go-smoke:${NETBOX_SMOKE_IMAGE_TAG}" >/dev/null 2>&1 || true
    rm -rf "${ARTIFACT_DIR}"
  fi
  exit "${status}"
}
trap cleanup EXIT INT TERM

fail() {
  echo "compose smoke: $*" >&2
  return 1
}

wait_for_healthy_service() {
  local service=$1
  local deadline=$((SECONDS + HEALTH_TIMEOUT_SECONDS))
  local container_id=""
  local state=""

  while ((SECONDS < deadline)); do
    container_id="$("${COMPOSE[@]}" ps --quiet "${service}")"
    if [[ -n "${container_id}" ]]; then
      state="$(
        docker inspect \
          --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' \
          "${container_id}" 2>/dev/null || true
      )"

      case "${state}" in
        healthy)
          return 0
          ;;
        exited | dead)
          fail "${service} stopped before becoming healthy"
          return 1
          ;;
      esac
    fi

    sleep 2
  done

  fail "${service} did not become healthy within ${HEALTH_TIMEOUT_SECONDS} seconds (last state: ${state:-missing})"
}

command -v docker >/dev/null 2>&1 || fail "docker is required"
docker compose version >/dev/null 2>&1 || fail "the Docker Compose v2 plugin is required"

if [[ "${BUILD_MODE}" == "offline" ]]; then
  command -v go >/dev/null 2>&1 || fail "Go is required for the offline smoke build"
  command -v node >/dev/null 2>&1 || fail "Node.js is required for the offline smoke build"
  command -v npm >/dev/null 2>&1 || fail "npm is required for the offline smoke build"
  docker image inspect postgres:16-alpine >/dev/null 2>&1 ||
    fail "cached postgres:16-alpine image is required for the offline smoke build"

  mkdir -p "${ARTIFACT_DIR}/web"
  (
    cd "${REPOSITORY_ROOT}/netbox-backend"
    env CGO_ENABLED=0 GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
      go build -trimpath -o "${ARTIFACT_DIR}/netbox_go" ./cmd/netbox_go
    env CGO_ENABLED=0 GOCACHE=/tmp/netbox-go-cache GOFLAGS=-buildvcs=false \
      go build -trimpath -o "${ARTIFACT_DIR}/netbox_go_admin" ./cmd/netbox_go_admin
  )
  (
    cd "${REPOSITORY_ROOT}/netbox-frontend"
    npm run toolchain:check
    npm run build:only
    cp -R dist "${ARTIFACT_DIR}/web/dist"
  )
fi

rendered_config="$("${COMPOSE[@]}" config)"
if [[ "${rendered_config}" == *"/docker-entrypoint-initdb.d"* ]]; then
  fail "PostgreSQL must not mount an initialization path; schema bootstrap belongs to the Go application"
fi

"${COMPOSE[@]}" up --detach postgres
wait_for_healthy_service postgres

postgres_container_id="$("${COMPOSE[@]}" ps --quiet postgres)"
postgres_data_mount_type="$(
  docker inspect \
    --format '{{range .Mounts}}{{if eq .Destination "/var/lib/postgresql/data"}}{{.Type}}{{end}}{{end}}' \
    "${postgres_container_id}"
)"
[[ "${postgres_data_mount_type}" == "volume" ]] || fail "PostgreSQL data must use a named Docker volume"

"${COMPOSE[@]}" up --detach --build netbox-go
wait_for_healthy_service netbox-go

"${COMPOSE[@]}" exec --no-TTY netbox-go \
  wget -qO- http://127.0.0.1:8080/health >/dev/null

public_table_count="$(
  "${COMPOSE[@]}" exec --no-TTY postgres \
    psql --username netbox --dbname netbox --tuples-only --no-align \
    --command "SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname = 'public';" |
    tr -d '[:space:]'
)"
[[ "${public_table_count}" =~ ^[1-9][0-9]*$ ]] || fail "application startup did not bootstrap any public PostgreSQL tables"

content_type_count="$(
  "${COMPOSE[@]}" exec --no-TTY postgres \
    psql --username netbox --dbname netbox --tuples-only --no-align \
    --command "SELECT count(*) FROM django_content_type;" |
    tr -d '[:space:]'
)"
[[ "${content_type_count}" =~ ^[1-9][0-9]*$ ]] || fail "application startup did not seed django_content_type"

# A second application start exercises the idempotent bootstrap path against
# the same non-empty database. Neither the schema nor seed cardinality may grow.
"${COMPOSE[@]}" restart netbox-go >/dev/null
wait_for_healthy_service netbox-go

public_table_count_after_restart="$(
  "${COMPOSE[@]}" exec --no-TTY postgres \
    psql --username netbox --dbname netbox --tuples-only --no-align \
    --command "SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname = 'public';" |
    tr -d '[:space:]'
)"
content_type_count_after_restart="$(
  "${COMPOSE[@]}" exec --no-TTY postgres \
    psql --username netbox --dbname netbox --tuples-only --no-align \
    --command "SELECT count(*) FROM django_content_type;" |
    tr -d '[:space:]'
)"

[[ "${public_table_count_after_restart}" == "${public_table_count}" ]] ||
  fail "the second startup changed the number of public tables"
[[ "${content_type_count_after_restart}" == "${content_type_count}" ]] ||
  fail "the second startup changed the number of content-type rows"

browser_evidence=""
if [[ "${RUN_BROWSER_E2E}" == "1" ]]; then
  browser_admin_username="browser-e2e-admin"
  browser_admin_password="Browser-E2E-only!2026-${RANDOM}${RANDOM}"
  browser_limited_username="browser-e2e-viewer"
  browser_limited_password="Browser-E2E-view-only!2026-${RANDOM}${RANDOM}"
  {
    printf '%s\n' "${browser_admin_password}"
    printf '%s\n' "${browser_admin_password}"
  } | "${COMPOSE[@]}" exec --no-TTY netbox-go \
    /app/netbox_go_admin bootstrap \
    --config /app/configs/netbox_go.docker.yml \
    --username "${browser_admin_username}" \
    --email browser-e2e@invalid.example >/dev/null

  {
    printf '%s\n' "${browser_admin_password}"
    printf '%s\n' "${browser_limited_password}"
    printf '%s\n' "${browser_limited_password}"
  } | "${COMPOSE[@]}" exec --no-TTY netbox-go \
    /app/netbox_go_admin create-user \
    --config /app/configs/netbox_go.docker.yml \
    --actor-username "${browser_admin_username}" \
    --username "${browser_limited_username}" \
    --email browser-e2e-viewer@invalid.example >/dev/null

  printf '%s\n' "${browser_admin_password}" | "${COMPOSE[@]}" exec --no-TTY netbox-go \
    /app/netbox_go_admin grant-permission \
    --config /app/configs/netbox_go.docker.yml \
    --actor-username "${browser_admin_username}" \
    --username "${browser_limited_username}" \
    --permission dcim.view_site >/dev/null

  NETBOX_E2E_BASE_URL="http://127.0.0.1:${NETBOX_HTTP_PORT}" \
    NETBOX_E2E_USERNAME="${browser_admin_username}" \
    NETBOX_E2E_PASSWORD="${browser_admin_password}" \
    NETBOX_E2E_LIMITED_USERNAME="${browser_limited_username}" \
    NETBOX_E2E_LIMITED_PASSWORD="${browser_limited_password}" \
    "${REPOSITORY_ROOT}/tests/browser/run.sh"
  unset browser_admin_password browser_limited_password
  browser_evidence=", and real-browser core workflows"
fi

echo "Compose smoke passed: PostgreSQL volume, fresh bootstrap, application health, restart idempotence${browser_evidence} verified."
