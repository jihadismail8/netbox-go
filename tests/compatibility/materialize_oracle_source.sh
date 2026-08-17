#!/usr/bin/env bash
set -Eeuo pipefail

usage() {
  printf 'usage: %s SOURCE_REPOSITORY TRUSTED_COMMIT TREE_PATH OUTPUT_DIRECTORY\n' \
    "${0##*/}" >&2
}

fail() {
  printf 'oracle source materialization failed: %s\n' "$1" >&2
  exit 1
}

if (( $# != 4 )); then
  usage
  exit 2
fi

SOURCE_REPOSITORY=$1
TRUSTED_COMMIT=$2
TREE_PATH=$3
OUTPUT_DIRECTORY=$4

if [[ ! "$TRUSTED_COMMIT" =~ ^[0-9a-f]{40}$ ]]; then
  fail "trusted commit must be a full lowercase SHA-1 object ID"
fi
if [[ -z "$TREE_PATH" ||
  "$TREE_PATH" == -* ||
  "$TREE_PATH" == /* ||
  "$TREE_PATH" == "." ||
  "$TREE_PATH" == ./* ||
  "$TREE_PATH" == */ ||
  "$TREE_PATH" == *//* ||
  "/$TREE_PATH/" == *"/../"* ||
  "/$TREE_PATH/" == *"/./"* ]]; then
  fail "tree path must be a canonical non-option relative repository path"
fi
if [[ -e "$OUTPUT_DIRECTORY" || -L "$OUTPUT_DIRECTORY" ]]; then
  fail "output directory already exists: $OUTPUT_DIRECTORY"
fi

# Every Git result used as source authority must ignore replacement refs. A
# local refs/replace entry must never be able to substitute another commit or
# tree for the reviewed Compatibility Baseline.
authority_git() {
  GIT_GRAFT_FILE=/dev/null GIT_NO_REPLACE_OBJECTS=1 \
    command git --no-replace-objects -c advice.graftFileDeprecated=false \
    -C "$SOURCE_REPOSITORY" "$@"
}

if [[ "$(authority_git rev-parse --is-inside-work-tree 2>/dev/null)" != "true" ]]; then
  fail "source repository is not a Git worktree: $SOURCE_REPOSITORY"
fi

ACTUAL_HEAD="$(authority_git rev-parse --verify 'HEAD^{commit}')"
if [[ "$ACTUAL_HEAD" != "$TRUSTED_COMMIT" ]]; then
  fail "source pin mismatch: expected $TRUSTED_COMMIT, found $ACTUAL_HEAD"
fi

if [[ -n "$(authority_git status --porcelain=v1 --untracked-files=no)" ]]; then
  fail "source repository has tracked modifications"
fi

if [[ "$(authority_git cat-file -t "$TRUSTED_COMMIT:$TREE_PATH" 2>/dev/null)" != "tree" ]]; then
  fail "tree path is not a tree at the trusted commit: $TREE_PATH"
fi

mkdir -p -- "$OUTPUT_DIRECTORY"
TREE_LIST="$(mktemp /tmp/netbox-oracle-tree.XXXXXX)"
cleanup_partial() {
  rm -f -- "$TREE_LIST"
  rm -rf -- "$OUTPUT_DIRECTORY"
}
trap cleanup_partial ERR INT TERM
authority_git ls-tree -r -z --full-tree "$TRUSTED_COMMIT" -- "$TREE_PATH" \
  >"$TREE_LIST"

ENTRY_COUNT=0
while IFS= read -r -d '' RECORD; do
  if [[ "$RECORD" != *$'\t'* ]]; then
    fail "trusted tree contains a malformed entry"
  fi
  HEADER=${RECORD%%$'\t'*}
  RELATIVE_PATH=${RECORD#*$'\t'}
  read -r MODE OBJECT_TYPE OBJECT_ID EXTRA <<<"$HEADER"
  if [[ -n "${EXTRA:-}" ||
    ! "$MODE" =~ ^[0-7]{6}$ ||
    "$OBJECT_TYPE" != "blob" ||
    ! "$OBJECT_ID" =~ ^[0-9a-f]{40,64}$ ]]; then
    fail "trusted tree contains an unsupported entry header"
  fi
  if [[ "$RELATIVE_PATH" != "$TREE_PATH"/* ||
    "$RELATIVE_PATH" == *$'\n'* ||
    "$RELATIVE_PATH" == *$'\r'* ||
    "$RELATIVE_PATH" == *//* ||
    "/$RELATIVE_PATH/" == *"/../"* ||
    "/$RELATIVE_PATH/" == *"/./"* ]]; then
    fail "trusted tree contains a non-canonical path"
  fi
  if [[ "$MODE" != "100644" && "$MODE" != "100755" ]]; then
    fail "trusted tree contains an unsupported mode at $RELATIVE_PATH: $MODE"
  fi
  DESTINATION="$OUTPUT_DIRECTORY/$RELATIVE_PATH"
  mkdir -p -- "${DESTINATION%/*}"
  authority_git cat-file blob "$OBJECT_ID" >"$DESTINATION"
  if [[ "$MODE" == "100755" ]]; then
    chmod 0755 -- "$DESTINATION"
  else
    chmod 0644 -- "$DESTINATION"
  fi
  ((ENTRY_COUNT += 1))
done <"$TREE_LIST"
rm -f -- "$TREE_LIST"

if (( ENTRY_COUNT == 0 )) || [[ ! -d "$OUTPUT_DIRECTORY/$TREE_PATH" ]]; then
  fail "trusted commit did not produce the requested tree: $TREE_PATH"
fi
trap - ERR INT TERM
