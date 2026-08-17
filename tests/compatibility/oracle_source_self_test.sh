#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MATERIALIZER="$SCRIPT_DIR/materialize_oracle_source.sh"
TEST_ROOT="$(mktemp -d /tmp/netbox-oracle-source-self-test.XXXXXX)"
REPOSITORY="$TEST_ROOT/oracle-repository"
OUTPUT_DIRECTORY="$TEST_ROOT/materialized"

cleanup() {
  rm -rf -- "$TEST_ROOT"
}
trap cleanup EXIT

fail() {
  printf 'oracle source self-test failed: %s\n' "$1" >&2
  exit 1
}

git init --quiet "$REPOSITORY"
git -C "$REPOSITORY" config user.name 'Oracle Source Self-Test'
git -C "$REPOSITORY" config user.email 'oracle-source-self-test@example.invalid'
mkdir -p "$REPOSITORY/netbox"
printf 'version: "4.4.6"\n' >"$REPOSITORY/netbox/release.yaml"
printf 'TRUSTED_ORACLE = True\n' >"$REPOSITORY/netbox/authority.py"
git -C "$REPOSITORY" add netbox
GIT_AUTHOR_DATE='2000-01-01T00:00:00Z' \
GIT_COMMITTER_DATE='2000-01-01T00:00:00Z' \
  git -C "$REPOSITORY" commit --quiet --message 'trusted oracle source'
TRUSTED_COMMIT="$(git -C "$REPOSITORY" rev-parse HEAD)"

# These worktree-only files must never enter the materialized oracle. The .py
# file is untracked; the bytecode is ignored through the repository-local
# exclude file so the trusted commit itself remains unchanged.
printf 'UNTRACKED_OVERRIDE = True\n' >"$REPOSITORY/netbox/untracked_override.py"
mkdir -p "$REPOSITORY/netbox/__pycache__"
printf '*.pyc\n' >>"$REPOSITORY/.git/info/exclude"
printf 'ignored bytecode payload\n' \
  >"$REPOSITORY/netbox/__pycache__/authority.cpython-312.pyc"
if [[ "$(git -C "$REPOSITORY" status --porcelain=v1 --untracked-files=all -- netbox/untracked_override.py)" != '?? netbox/untracked_override.py' ]]; then
  fail "Python fixture is not untracked"
fi
if ! git -C "$REPOSITORY" check-ignore --quiet \
  netbox/__pycache__/authority.cpython-312.pyc; then
  fail "bytecode fixture is not ignored"
fi

REPLACEMENT_RELEASE_BLOB="$(
  printf 'version: "99.99.99-replacement"\n' |
    git -C "$REPOSITORY" hash-object -w --stdin
)"
REPLACEMENT_AUTHORITY_BLOB="$(
  printf 'TRUSTED_ORACLE = False\n' |
    git -C "$REPOSITORY" hash-object -w --stdin
)"
REPLACEMENT_ONLY_BLOB="$(
  printf 'REPLACEMENT_REF_WAS_USED = True\n' |
    git -C "$REPOSITORY" hash-object -w --stdin
)"
REPLACEMENT_NETBOX_TREE="$(
  printf '100644 blob %s\tauthority.py\n100644 blob %s\trelease.yaml\n100644 blob %s\treplacement_only.py\n' \
    "$REPLACEMENT_AUTHORITY_BLOB" \
    "$REPLACEMENT_RELEASE_BLOB" \
    "$REPLACEMENT_ONLY_BLOB" |
    git -C "$REPOSITORY" mktree
)"
REPLACEMENT_ROOT_TREE="$(
  printf '040000 tree %s\tnetbox\n' "$REPLACEMENT_NETBOX_TREE" |
    git -C "$REPOSITORY" mktree
)"
REPLACEMENT_COMMIT="$(
  printf 'malicious replacement commit\n' |
    GIT_AUTHOR_DATE='2000-01-02T00:00:00Z' \
      GIT_COMMITTER_DATE='2000-01-02T00:00:00Z' \
      git -C "$REPOSITORY" commit-tree "$REPLACEMENT_ROOT_TREE"
)"
git -C "$REPOSITORY" replace "$TRUSTED_COMMIT" "$REPLACEMENT_COMMIT"

# Repository-local attributes are untracked Git metadata and can make ordinary
# `git archive` omit or substitute committed content. Direct tree/blob
# materialization must ignore them just as it ignores replacement refs.
printf 'netbox/authority.py export-ignore\nnetbox/release.yaml export-subst\n' \
  >"$REPOSITORY/.git/info/attributes"
if [[ "$(git -C "$REPOSITORY" check-attr export-ignore -- netbox/authority.py)" != \
  'netbox/authority.py: export-ignore: set' ]]; then
  fail "fixture repository-local export attribute is not active"
fi
printf '%s %s\n' "$TRUSTED_COMMIT" "$REPLACEMENT_COMMIT" \
  >"$REPOSITORY/.git/info/grafts"
git -C "$REPOSITORY" config advice.graftFileDeprecated false
if ! git --no-replace-objects \
  -C "$REPOSITORY" merge-base --is-ancestor \
  "$REPLACEMENT_COMMIT" "$TRUSTED_COMMIT"; then
  fail "fixture repository-local graft is not active"
fi

ACTIVE_RELEASE="$(git -C "$REPOSITORY" show "$TRUSTED_COMMIT:netbox/release.yaml")"
if [[ "$ACTIVE_RELEASE" != 'version: "99.99.99-replacement"' ]]; then
  fail "fixture replacement ref is not active"
fi

if "$MATERIALIZER" \
  "$REPOSITORY" \
  "$REPLACEMENT_COMMIT" \
  netbox \
  "$TEST_ROOT/wrong-head-output" \
  >"$TEST_ROOT/wrong-head.stdout" 2>"$TEST_ROOT/wrong-head.stderr"; then
  fail "materializer accepted a trusted commit other than HEAD"
fi

printf 'DIRTY_TRACKED_OVERRIDE = True\n' >>"$REPOSITORY/netbox/authority.py"
if "$MATERIALIZER" \
  "$REPOSITORY" \
  "$TRUSTED_COMMIT" \
  netbox \
  "$TEST_ROOT/dirty-output" \
  >"$TEST_ROOT/dirty.stdout" 2>"$TEST_ROOT/dirty.stderr"; then
  fail "materializer accepted tracked worktree modifications"
fi
printf 'TRUSTED_ORACLE = True\n' >"$REPOSITORY/netbox/authority.py"

ln -s authority.py "$REPOSITORY/netbox/unsupported-link"
git -C "$REPOSITORY" add netbox/unsupported-link
GIT_AUTHOR_DATE='2000-01-03T00:00:00Z' \
GIT_COMMITTER_DATE='2000-01-03T00:00:00Z' \
  git -C "$REPOSITORY" commit --quiet --message 'unsupported oracle link'
UNSUPPORTED_COMMIT="$(
  git --no-replace-objects -C "$REPOSITORY" rev-parse HEAD
)"
if "$MATERIALIZER" \
  "$REPOSITORY" \
  "$UNSUPPORTED_COMMIT" \
  netbox \
  "$TEST_ROOT/unsupported-output" \
  >"$TEST_ROOT/unsupported.stdout" 2>"$TEST_ROOT/unsupported.stderr"; then
  fail "materializer accepted an unsupported Git entry type"
fi
git --no-replace-objects -C "$REPOSITORY" reset --quiet --hard "$TRUSTED_COMMIT"

"$MATERIALIZER" \
  "$REPOSITORY" \
  "$TRUSTED_COMMIT" \
  netbox \
  "$OUTPUT_DIRECTORY"

if ! grep -qx 'version: "4.4.6"' "$OUTPUT_DIRECTORY/netbox/release.yaml"; then
  fail "materialized release did not come from the trusted commit"
fi
if ! grep -qx 'TRUSTED_ORACLE = True' "$OUTPUT_DIRECTORY/netbox/authority.py"; then
  fail "materialized source content did not come from the trusted commit"
fi
if [[ -e "$OUTPUT_DIRECTORY/netbox/untracked_override.py" ]]; then
  fail "untracked Python content entered the materialized oracle"
fi
if [[ -e "$OUTPUT_DIRECTORY/netbox/__pycache__/authority.cpython-312.pyc" ]]; then
  fail "ignored bytecode entered the materialized oracle"
fi
if [[ -e "$OUTPUT_DIRECTORY/netbox/replacement_only.py" ]]; then
  fail "replacement-ref content entered the materialized oracle"
fi

cleanup
trap - EXIT
if [[ -e "$TEST_ROOT" ]]; then
  fail "temporary test directory was not removed"
fi

printf 'oracle source materializer self-test passed\n'
