#!/usr/bin/env bash

set -euo pipefail

quality_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
checker="$quality_dir/check_coverage.sh"
test_tmp="$(mktemp -d /tmp/netbox-go-coverage-policy-test.XXXXXX)"
cleanup() {
	rm -f "$test_tmp"/*
	rmdir "$test_tmp"
}
trap cleanup EXIT

baseline="$test_tmp/baseline.tsv"
profile="$test_tmp/coverage.out"
measured="$test_tmp/measured-packages"
all="$test_tmp/all-packages"
exclusions="$test_tmp/legacy-exclusions.yml"

write_baseline() {
	local extra_exclusion="${1:-}"
	{
		printf 'schema_version\t1\n'
		printf 'owner\tbackend-rewrite\n'
		printf 'unit_scope\tmodule-packages-minus-reviewed-exclusions-v1\n'
		printf 'coverage_mode\tatomic\n'
		printf 'baseline_covered_statements\t4\n'
		printf 'baseline_total_statements\t5\n'
		printf 'package\tnetbox-go/pkg/measured\n'
		printf 'excluded_package\tnetbox-go/internal/service\n'
		if [[ -n "$extra_exclusion" ]]; then
			printf 'excluded_package\t%s\n' "$extra_exclusion"
		fi
	} >"$baseline"
}

write_profile() {
	local covered_count="$1"
	{
		printf 'mode: atomic\n'
		printf 'netbox-go/pkg/measured/a.go:1.1,1.2 4 %s\n' "$covered_count"
		printf 'netbox-go/pkg/measured/a.go:2.1,2.2 1 0\n'
	} >"$profile"
}

expect_failure() {
	local expected="$1"
	shift
	local output
	if output="$(bash "$checker" "$@" 2>&1)"; then
		echo "coverage checker unexpectedly accepted: $expected" >&2
		exit 1
	fi
	if ! grep -Fq "$expected" <<<"$output"; then
		echo "coverage checker failed without the expected diagnostic: $expected" >&2
		echo "$output" >&2
		exit 1
	fi
}

write_baseline
write_profile 1
printf 'netbox-go/pkg/measured\n' >"$measured"
printf 'netbox-go/internal/service\nnetbox-go/pkg/measured\n' >"$all"
printf 'owner: backend-rewrite\nremoval_milestone: legacy-retirement\nexclusions:\n  - scope: unit-runtime\n    package: netbox-go/internal/service\n' >"$exclusions"
bash "$checker" "$baseline" "$profile" "$measured" "$all" "$exclusions" >/dev/null

write_profile 0
expect_failure "coverage regression" "$baseline" "$profile" "$measured" "$all" "$exclusions"

write_profile 1
printf 'netbox-go/pkg/measured\nnetbox-go/pkg/new\n' >"$measured"
printf 'netbox-go/internal/service\nnetbox-go/pkg/measured\nnetbox-go/pkg/new\n' >"$all"
expect_failure "coverage measured package set differs" "$baseline" "$profile" "$measured" "$all" "$exclusions"

printf 'netbox-go/pkg/measured\n' >"$measured"
write_baseline netbox-go/pkg/hidden
printf 'netbox-go/internal/service\nnetbox-go/pkg/hidden\nnetbox-go/pkg/measured\n' >"$all"
expect_failure "coverage exclusion is not declared" "$baseline" "$profile" "$measured" "$all" "$exclusions"

echo "coverage policy self-test passed"
