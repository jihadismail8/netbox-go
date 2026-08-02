#!/usr/bin/env bash

set -euo pipefail

readonly service_name="netbox_go"
readonly archive="./${service_name}-binary.tar.gz"

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 username host" >&2
  echo "Authentication must use an SSH agent or explicitly configured SSH key." >&2
  exit 2
fi

readonly username="$1"
readonly host="$2"
readonly remote="${username}@${host}"

if [[ ! -f "${archive}" ]]; then
  echo "Binary archive not found: ${archive}" >&2
  exit 1
fi

scp -- "${archive}" "${remote}:/tmp/"
ssh -- "${remote}" \
  "cd /tmp && tar -xzf '${service_name}-binary.tar.gz' && bash '/tmp/${service_name}-binary/deploy.sh'"
