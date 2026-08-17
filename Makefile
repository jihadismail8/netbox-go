SHELL := /bin/bash

.PHONY: check backend-check frontend-check generated-check contracts-check docs-check frontend-install source-digest source-manifest deployment-smoke deployment-image-smoke browser-e2e compatibility-test compatibility-comparator-test compatibility-oracle-source-test

# Run the complete deterministic repository gate used by CI.
check: backend-check frontend-check generated-check contracts-check docs-check

backend-check:
	$(MAKE) -C netbox-backend check

frontend-install:
	cd netbox-frontend && npm ci

frontend-check: frontend-install
	cd netbox-frontend && npm run toolchain:check
	cd netbox-frontend && npm run format:check
	cd netbox-frontend && npm run lint
	cd netbox-frontend && npm run typecheck
	cd netbox-frontend && npm run typecheck:test
	cd netbox-frontend && npm run test:coverage
	cd netbox-frontend && npm run build

generated-check:
	node scripts/generate_contract_inventory.mjs --check

contracts-check: compatibility-oracle-source-test
	@upstream_status="$$(GIT_GRAFT_FILE=/dev/null GIT_NO_REPLACE_OBJECTS=1 git --no-replace-objects -c advice.graftFileDeprecated=false -C netbox status --porcelain=v1 --untracked-files=no)"; \
	if [[ -n "$$upstream_status" ]]; then \
		echo "contracts-check: pinned NetBox checkout has tracked modifications" >&2; \
		exit 1; \
	fi
	node scripts/validate_contract_inventory.mjs contracts/netbox/v4.4.6-post7/inventory
	node scripts/validate_traceability.test.mjs
	node scripts/validate_capability_profile.mjs contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml
	@descriptor="$$(mktemp /tmp/netbox-go-contract.XXXXXX.pb)"; \
	trap 'rm -f "$$descriptor"' EXIT; \
	cd netbox-backend && BUF_CACHE_DIR=/tmp/netbox-go-buf-cache ../.tools/bin/buf build -o "$$descriptor"; \
	cd .. && node scripts/generate_contract_docs.mjs --check \
		contracts/netbox/v4.4.6-post7/profiles/core-workflow-v1.yaml "$$descriptor"
	node scripts/validate_openapi.mjs netbox-backend/api/openapi/netbox-go-v1.yaml

docs-check:
	node scripts/check_markdown_links.mjs

# Produce the versioned source-v2 identifier for owned source used by retained
# evidence. The pinned upstream oracle is recorded separately by Git SHA.
source-digest:
	node scripts/source_digest.mjs

# Emit the complete mode-aware canonical manifest committed by a two-digest
# claim. Store it under docs/evidence; that directory is digest-excluded.
source-manifest:
	node scripts/source_digest.mjs --manifest

# Destructive only to its uniquely named Compose project; intentionally kept
# outside the fast repository gate because it builds images and owns Docker.
deployment-smoke:
	NETBOX_SMOKE_BUILD_MODE=offline ./tests/deployment/compose_smoke.sh

# Connected-CI check for the production multi-stage Dockerfile and its pinned
# base images. The default smoke packages host-built artifacts without pulls.
deployment-image-smoke:
	NETBOX_SMOKE_BUILD_MODE=production ./tests/deployment/compose_smoke.sh

# Real Chrome over the disposable root Compose stack. The driver has no npm
# dependency and retains credential-free diagnostics when an assertion fails.
browser-e2e:
	NETBOX_SMOKE_BUILD_MODE=offline NETBOX_SMOKE_RUN_BROWSER_E2E=1 ./tests/deployment/compose_smoke.sh

# Differential profile gate against the exact checked-in NetBox oracle. The
# harness refuses source/config drift and uses only already-cached images.
compatibility-test:
	./tests/compatibility/run.sh

# Fast proof that the semantic comparator rejects an intentional divergence.
compatibility-comparator-test:
	node ./tests/compatibility/comparator_self_test.mjs

# Hermetic proof that the production oracle materializer excludes worktree-only
# content and ignores local Git replacement refs.
compatibility-oracle-source-test:
	./tests/compatibility/oracle_source_self_test.sh
