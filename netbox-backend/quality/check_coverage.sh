#!/usr/bin/env bash

set -euo pipefail

if [[ "$#" -ne 5 ]]; then
	echo "usage: check_coverage.sh BASELINE PROFILE MEASURED_PACKAGES ALL_PACKAGES EXCLUSION_MANIFEST" >&2
	exit 2
fi

baseline="$1"
profile="$2"
measured_packages="$3"
all_packages="$4"
exclusion_manifest="$5"

for required in "$baseline" "$profile" "$measured_packages" "$all_packages" "$exclusion_manifest"; do
	if [[ ! -f "$required" ]]; then
		echo "coverage policy input does not exist: $required" >&2
		exit 2
	fi
done

policy_tmp="$(mktemp -d /tmp/netbox-go-coverage-policy.XXXXXX)"
cleanup() {
	rm -f "$policy_tmp/baseline-packages" \
		"$policy_tmp/baseline-exclusions" \
		"$policy_tmp/actual-measured" \
		"$policy_tmp/actual-all" \
		"$policy_tmp/expected-all"
	rmdir "$policy_tmp"
}
trap cleanup EXIT

if ! awk -F '\t' '
	BEGIN { valid = 1 }
	NF != 2 {
		printf "coverage baseline line %d must contain exactly two tab-separated fields\n", NR > "/dev/stderr"
		valid = 0
		next
	}
	$1 != "schema_version" &&
	$1 != "owner" &&
	$1 != "unit_scope" &&
	$1 != "coverage_mode" &&
	$1 != "baseline_covered_statements" &&
	$1 != "baseline_total_statements" &&
	$1 != "package" &&
	$1 != "excluded_package" {
		printf "coverage baseline line %d has unknown key %s\n", NR, $1 > "/dev/stderr"
		valid = 0
	}
	$2 == "" {
		printf "coverage baseline line %d has an empty value\n", NR > "/dev/stderr"
		valid = 0
	}
	END { if (!valid) exit 1 }
' "$baseline"; then
	exit 2
fi

metadata_value() {
	local key="$1"
	local values
	values="$(awk -F '\t' -v key="$key" '$1 == key { print $2 }' "$baseline")"
	if [[ -z "$values" || "$values" == *$'\n'* ]]; then
		echo "coverage baseline must contain exactly one $key row" >&2
		exit 2
	fi
	printf '%s' "$values"
}

schema_version="$(metadata_value schema_version)"
owner="$(metadata_value owner)"
unit_scope="$(metadata_value unit_scope)"
coverage_mode="$(metadata_value coverage_mode)"
baseline_covered="$(metadata_value baseline_covered_statements)"
baseline_total="$(metadata_value baseline_total_statements)"

if [[ "$schema_version" != "1" ]]; then
	echo "unsupported coverage baseline schema_version: $schema_version" >&2
	exit 2
fi
if [[ -z "$owner" || "$unit_scope" != "module-packages-minus-reviewed-exclusions-v1" ]]; then
	echo "coverage baseline has an invalid owner or unit_scope" >&2
	exit 2
fi
if [[ "$coverage_mode" != "atomic" ]]; then
	echo "coverage baseline mode must be atomic; found $coverage_mode" >&2
	exit 2
fi
if [[ ! "$baseline_covered" =~ ^[0-9]+$ || ! "$baseline_total" =~ ^[1-9][0-9]*$ ]]; then
	echo "coverage baseline statement counts must be non-negative integers with a positive total" >&2
	exit 2
fi
if ((baseline_covered > baseline_total)); then
	echo "coverage baseline covered statements exceed total statements" >&2
	exit 2
fi

awk -F '\t' '$1 == "package" { print $2 }' "$baseline" >"$policy_tmp/baseline-packages"
awk -F '\t' '$1 == "excluded_package" { print $2 }' "$baseline" >"$policy_tmp/baseline-exclusions"

for package_file in "$policy_tmp/baseline-packages" "$policy_tmp/baseline-exclusions"; do
	if [[ ! -s "$package_file" ]]; then
		echo "coverage baseline must list measured packages and reviewed exclusions" >&2
		exit 2
	fi
	if ! LC_ALL=C sort -c -u "$package_file" 2>/dev/null; then
		echo "coverage baseline package rows must be sorted and unique: $package_file" >&2
		exit 2
	fi
	if grep -Evq '^[A-Za-z0-9._/-]+$' "$package_file"; then
		echo "coverage baseline contains an invalid package path: $package_file" >&2
		exit 2
	fi
done

normalize_package_file() {
	local source="$1"
	local destination="$2"
	if [[ ! -s "$source" ]]; then
		echo "coverage package inventory is empty: $source" >&2
		exit 2
	fi
	if grep -Eq '^[[:space:]]*$' "$source"; then
		echo "coverage package inventory contains an empty row: $source" >&2
		exit 2
	fi
	if grep -Evq '^[A-Za-z0-9._/-]+$' "$source"; then
		echo "coverage package inventory contains an invalid package path: $source" >&2
		exit 2
	fi
	LC_ALL=C sort -u "$source" >"$destination"
}

normalize_package_file "$measured_packages" "$policy_tmp/actual-measured"
normalize_package_file "$all_packages" "$policy_tmp/actual-all"

if ! diff -u "$policy_tmp/baseline-packages" "$policy_tmp/actual-measured"; then
	echo "coverage measured package set differs from the reviewed baseline; classify the package change and record a fresh measured baseline" >&2
	exit 1
fi

LC_ALL=C sort -u "$policy_tmp/baseline-packages" "$policy_tmp/baseline-exclusions" >"$policy_tmp/expected-all"
if ! diff -u "$policy_tmp/expected-all" "$policy_tmp/actual-all"; then
	echo "module packages are missing from the coverage baseline or were excluded without review" >&2
	echo "Every package must be measured or declared in $exclusion_manifest with an owner and removal milestone." >&2
	exit 1
fi

for ownership_field in owner removal_milestone; do
	if ! awk -v field="$ownership_field" '
		$1 == field ":" && NF >= 2 { found = 1 }
		END { exit !found }
	' "$exclusion_manifest"; then
		echo "coverage exclusion manifest has no $ownership_field: $exclusion_manifest" >&2
		exit 1
	fi
done

while IFS= read -r excluded_package; do
	if ! awk -v expected="$excluded_package" '
		$1 == "package:" {
			value = $0
			sub(/^[[:space:]]*package:[[:space:]]*/, "", value)
			if (value == expected) found = 1
		}
		END { exit !found }
	' "$exclusion_manifest"; then
		echo "coverage exclusion is not declared in $exclusion_manifest: $excluded_package" >&2
		exit 1
	fi
done <"$policy_tmp/baseline-exclusions"

profile_mode="$(sed -n '1{s/^mode: //;p;q;}' "$profile")"
if [[ "$profile_mode" != "$coverage_mode" ]]; then
	echo "coverage profile mode is ${profile_mode:-missing}; expected $coverage_mode" >&2
	exit 2
fi

if ! coverage_counts="$(awk '
	NR == 1 { next }
	NF != 3 || $2 !~ /^[0-9]+$/ || $3 !~ /^[0-9]+$/ {
		printf "invalid coverage profile line %d\n", NR > "/dev/stderr"
		invalid = 1
		next
	}
	{
		blocks++
		total += $2
		if ($3 > 0) covered += $2
	}
	END {
		if (invalid || blocks == 0 || total == 0) exit 1
		printf "%.0f\t%.0f", covered, total
	}
' "$profile")"; then
	echo "coverage profile has no valid statement data: $profile" >&2
	exit 2
fi

IFS=$'\t' read -r current_covered current_total <<<"$coverage_counts"
baseline_percent="$(awk -v covered="$baseline_covered" -v total="$baseline_total" 'BEGIN { printf "%.4f", 100 * covered / total }')"
current_percent="$(awk -v covered="$current_covered" -v total="$current_total" 'BEGIN { printf "%.4f", 100 * covered / total }')"

if ((current_covered * baseline_total < baseline_covered * current_total)); then
	echo "coverage regression: current ${current_percent}% (${current_covered}/${current_total}) is below baseline ${baseline_percent}% (${baseline_covered}/${baseline_total})" >&2
	echo "Add or restore tests; do not lower the recorded baseline to make the gate pass." >&2
	exit 1
fi

echo "coverage ${current_percent}% (${current_covered}/${current_total}) meets baseline ${baseline_percent}% (${baseline_covered}/${baseline_total}); packages: $(wc -l <"$policy_tmp/actual-measured" | tr -d ' ') measured, $(wc -l <"$policy_tmp/baseline-exclusions" | tr -d ' ') reviewed exclusion(s)"
