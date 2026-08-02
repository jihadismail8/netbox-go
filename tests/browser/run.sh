#!/usr/bin/env bash

set -Eeuo pipefail

readonly REPOSITORY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
readonly CHROME_BIN="${NETBOX_E2E_CHROME_BIN:-$(command -v google-chrome || command -v google-chrome-stable || command -v chromium || true)}"
readonly PROFILE_DIR="$(mktemp -d /tmp/netbox-go-browser-profile.XXXXXX)"
readonly ARTIFACT_DIR="${NETBOX_E2E_ARTIFACT_DIR:-$(mktemp -d /tmp/netbox-go-browser-artifacts.XXXXXX)}"
readonly CHROME_LOG="${ARTIFACT_DIR}/chrome.log"
chrome_pid=""

cleanup() {
  local status=$?
  trap - EXIT INT TERM
  if [[ -n "${chrome_pid}" ]]; then
    kill "${chrome_pid}" >/dev/null 2>&1 || true
    wait "${chrome_pid}" >/dev/null 2>&1 || true
  fi
  rm -rf "${PROFILE_DIR}"
  if ((status == 0)); then
    rm -f "${CHROME_LOG}"
  else
    echo "browser e2e: credential-free diagnostics retained at ${ARTIFACT_DIR}" >&2
  fi
  exit "${status}"
}
trap cleanup EXIT INT TERM

[[ -n "${CHROME_BIN}" && -x "${CHROME_BIN}" ]] || {
  echo "browser e2e: Google Chrome or Chromium is required" >&2
  exit 2
}
command -v node >/dev/null 2>&1 || {
  echo "browser e2e: Node.js is required" >&2
  exit 2
}
: "${NETBOX_E2E_BASE_URL:?browser e2e: NETBOX_E2E_BASE_URL is required}"
: "${NETBOX_E2E_USERNAME:?browser e2e: NETBOX_E2E_USERNAME is required}"
: "${NETBOX_E2E_PASSWORD:?browser e2e: NETBOX_E2E_PASSWORD is required}"
: "${NETBOX_E2E_LIMITED_USERNAME:?browser e2e: NETBOX_E2E_LIMITED_USERNAME is required}"
: "${NETBOX_E2E_LIMITED_PASSWORD:?browser e2e: NETBOX_E2E_LIMITED_PASSWORD is required}"

mkdir -p "${ARTIFACT_DIR}"
chmod 700 "${ARTIFACT_DIR}"

"${CHROME_BIN}" \
  --headless=new \
  --no-sandbox \
  --disable-gpu \
  --disable-background-networking \
  --disable-component-update \
  --disable-default-apps \
  --disable-extensions \
  --disable-sync \
  --metrics-recording-only \
  --no-first-run \
  --remote-debugging-address=127.0.0.1 \
  --remote-debugging-port=0 \
  --user-data-dir="${PROFILE_DIR}" \
  about:blank >"${CHROME_LOG}" 2>&1 &
chrome_pid=$!

for _ in {1..100}; do
  [[ -s "${PROFILE_DIR}/DevToolsActivePort" ]] && break
  kill -0 "${chrome_pid}" >/dev/null 2>&1 || {
    echo "browser e2e: Chrome exited before exposing DevTools" >&2
    exit 1
  }
  sleep 0.1
done

[[ -s "${PROFILE_DIR}/DevToolsActivePort" ]] || {
  echo "browser e2e: Chrome did not expose DevTools within 10 seconds" >&2
  exit 1
}

export NETBOX_E2E_CDP_PORT
export NETBOX_E2E_ARTIFACT_DIR="${ARTIFACT_DIR}"
NETBOX_E2E_CDP_PORT="$(sed -n '1p' "${PROFILE_DIR}/DevToolsActivePort")"

node "${REPOSITORY_ROOT}/tests/browser/browser_e2e.mjs"
