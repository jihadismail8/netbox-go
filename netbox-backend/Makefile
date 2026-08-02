SHELL := /bin/bash

PROJECT_NAME := netbox-go
GO ?= go
GOFMT ?= gofmt
LOCAL_GOLANGCI_LINT := ../.tools/bin/golangci-lint
GOLANGCI_LINT ?= $(if $(wildcard $(LOCAL_GOLANGCI_LINT)),$(LOCAL_GOLANGCI_LINT),golangci-lint)
GOLANGCI_LINT_VERSION := 2.12.2
PROTOC ?= protoc
PROTOC_VERSION := 3.21.12
LOCAL_BUF := ../.tools/bin/buf
BUF ?= $(if $(wildcard $(LOCAL_BUF)),$(LOCAL_BUF),buf)
BUF_VERSION := 1.71.0
LOCAL_PROTOC_GEN_GO := ../.tools/bin/protoc-gen-go
PROTOC_GEN_GO ?= $(if $(wildcard $(LOCAL_PROTOC_GEN_GO)),$(LOCAL_PROTOC_GEN_GO),$(HOME)/go/bin/protoc-gen-go)
PROTOC_GEN_GO_VERSION := 1.36.10
LOCAL_PROTOC_GEN_GO_GRPC := ../.tools/bin/protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC ?= $(if $(wildcard $(LOCAL_PROTOC_GEN_GO_GRPC)),$(LOCAL_PROTOC_GEN_GO_GRPC),$(HOME)/go/bin/protoc-gen-go-grpc)
PROTOC_GEN_GO_GRPC_VERSION := 1.5.1
BUF_CACHE_DIR ?= /tmp/netbox-go-buf-cache
GOCACHE ?= /tmp/netbox-go-cache
GOLANGCI_LINT_CACHE ?= /tmp/netbox-go-golangci-cache
GOFLAGS ?= -buildvcs=false

# Sponge generated these tests as clients for an already-running gRPC server.
# They are compile-checked below, but they are not hermetic unit tests. This is
# the only package omitted from the default package-wide unit and vet commands.
# Ownership and the removal milestone are recorded in quality/legacy-exclusions.yml.
LEGACY_LIVE_CLIENT_PACKAGE := $(PROJECT_NAME)/internal/service
HERMETIC_SERVICE_TEST := ^TestRegisterAllService$$
LEGACY_PROTO_ROOTS := api/netbox_go/v1 api/types
CANONICAL_PROTO_ROOT := api/proto

.PHONY: fmt
# Format all Go source files. This is the only mutating quality target.
fmt:
	@$(GOFMT) -s -w .

.PHONY: fmt-check
# Verify Go formatting without changing the worktree.
fmt-check:
	@unformatted="$$($(GOFMT) -l .)"; \
	if [[ -n "$$unformatted" ]]; then \
		echo "The following Go files are not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

.PHONY: vet
# Run go vet over every package except the classified legacy live-client package.
vet: legacy-exclusions-check
	@packages="$$(GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) list ./... | grep -v '^$(LEGACY_LIVE_CLIENT_PACKAGE)$$')"; \
	test -n "$$packages"; \
	GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) vet $$packages

.PHONY: lint
# Run the pinned golangci-lint policy without applying fixes.
lint: golangci-version-check legacy-exclusions-check
	@packages="$$(GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) list -f '{{.Dir}}' ./... | \
		sed "s|^$$(pwd)|.|" | \
		grep -Ev '^\./api/(netbox_go/v1|types)$$|^\./internal/(cache|dao|ecode|handler|model|routers|service)($$|/)')"; \
	test -n "$$packages"; \
	GOCACHE="$(GOCACHE)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" GOFLAGS="$(GOFLAGS)" \
		$(GOLANGCI_LINT) run --config .golangci.yml $$packages

.PHONY: golangci-version-check
# Refuse an unpinned golangci-lint binary so local and CI results agree.
golangci-version-check:
	@command -v $(GOLANGCI_LINT) >/dev/null || { \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) is required"; \
		exit 1; \
	}
	@actual="$$($(GOLANGCI_LINT) version --short 2>/dev/null | awk '{ print $$1 }')"; \
	if [[ "$$actual" != "$(GOLANGCI_LINT_VERSION)" ]]; then \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) is required; found $${actual:-unknown}"; \
		exit 1; \
	fi

.PHONY: unit
# Run every hermetic unit test with the race detector and no cached results.
unit: legacy-exclusions-check
	@packages="$$(GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) list ./... | grep -v '^$(LEGACY_LIVE_CLIENT_PACKAGE)$$')"; \
	test -n "$$packages"; \
	GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) test -race -short -count=1 $$packages
	@GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) test -race -short -count=1 \
		-run '$(HERMETIC_SERVICE_TEST)' ./internal/service

.PHONY: test
# Backward-compatible alias for the hermetic unit gate.
test: unit

.PHONY: legacy-client-compile
# Compile, but never execute, generator-owned tests that require a live server.
legacy-client-compile: legacy-exclusions-check
	@GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) test -run '^$$' ./internal/service

.PHONY: proto-source-classification-check
# Keep frozen Sponge protobufs separate from the canonical Buf module. A new
# source root must be classified explicitly instead of being silently absorbed
# into the legacy compile-only exclusion.
proto-source-classification-check:
	@legacy_files="$$(find $(LEGACY_PROTO_ROOTS) -type f -name '*.proto' -print | LC_ALL=C sort)"; \
	canonical_files="$$(find $(CANONICAL_PROTO_ROOT) -type f -name '*.proto' -print | LC_ALL=C sort)"; \
	unclassified="$$(find api -type f -name '*.proto' \
		! -path 'api/netbox_go/v1/*' \
		! -path 'api/types/*' \
		! -path 'api/proto/*' -print | LC_ALL=C sort)"; \
	if [[ -z "$$legacy_files" ]]; then \
		echo "no frozen legacy protobuf sources found under $(LEGACY_PROTO_ROOTS)"; \
		exit 1; \
	fi; \
	if [[ -z "$$canonical_files" ]]; then \
		echo "no canonical protobuf sources found under $(CANONICAL_PROTO_ROOT)"; \
		exit 1; \
	fi; \
	if [[ -n "$$unclassified" ]]; then \
		echo "unclassified protobuf sources found outside the frozen and canonical roots:"; \
		echo "$$unclassified"; \
		exit 1; \
	fi

.PHONY: legacy-proto-compile
# Compile frozen protobuf sources to a temporary descriptor without generation.
legacy-proto-compile: protoc-version-check proto-source-classification-check
	@descriptor="$$(mktemp /tmp/netbox-go-legacy-proto.XXXXXX.pb)"; \
	trap 'rm -f "$$descriptor"' EXIT; \
	files="$$(find $(LEGACY_PROTO_ROOTS) -type f -name '*.proto' -print | LC_ALL=C sort)"; \
	$(PROTOC) --proto_path=. --proto_path=third_party --include_imports \
		--descriptor_set_out="$$descriptor" $$files; \
	test -s "$$descriptor"

.PHONY: api-proto-check
# Validate and reproducibly compile the canonical handwritten protobuf module.
api-proto-check: protoc-version-check protobuf-tool-version-check
	@BUF_CACHE_DIR="$(BUF_CACHE_DIR)" $(BUF) format --diff --exit-code
	@BUF_CACHE_DIR="$(BUF_CACHE_DIR)" $(BUF) lint
	@descriptor="$$(mktemp /tmp/netbox-go-core-proto.XXXXXX.pb)"; \
	tmp="$$(mktemp -d /tmp/netbox-go-core-generated.XXXXXX)"; \
	trap 'rm -f "$$descriptor"; rm -rf "$$tmp"' EXIT; \
	BUF_CACHE_DIR="$(BUF_CACHE_DIR)" $(BUF) build -o "$$descriptor"; \
	PATH="$(dir $(PROTOC_GEN_GO)):$(dir $(PROTOC_GEN_GO_GRPC)):$$PATH" \
		BUF_CACHE_DIR="$(BUF_CACHE_DIR)" $(BUF) generate --template buf.gen.yaml --output "$$tmp"; \
	diff -ru gen/go "$$tmp/gen/go"

.PHONY: protobuf-tool-version-check
protobuf-tool-version-check:
	@actual="$$($(BUF) --version 2>/dev/null)"; \
	if [[ "$$actual" != "$(BUF_VERSION)" ]]; then \
		echo "buf $(BUF_VERSION) is required; found $${actual:-unknown}"; \
		exit 1; \
	fi
	@actual="$$($(PROTOC_GEN_GO) --version 2>/dev/null | awk '{ print $$2 }' | sed 's/^v//')"; \
	if [[ "$$actual" != "$(PROTOC_GEN_GO_VERSION)" ]]; then \
		echo "protoc-gen-go $(PROTOC_GEN_GO_VERSION) is required; found $${actual:-unknown}"; \
		exit 1; \
	fi
	@actual="$$($(PROTOC_GEN_GO_GRPC) --version 2>/dev/null | awk '{ print $$2 }')"; \
	if [[ "$$actual" != "$(PROTOC_GEN_GO_GRPC_VERSION)" ]]; then \
		echo "protoc-gen-go-grpc $(PROTOC_GEN_GO_GRPC_VERSION) is required; found $${actual:-unknown}"; \
		exit 1; \
	fi

.PHONY: protoc-version-check
# Refuse an unpinned protoc binary so descriptor checks are reproducible.
protoc-version-check:
	@command -v $(PROTOC) >/dev/null || { \
		echo "protoc $(PROTOC_VERSION) is required"; \
		exit 1; \
	}
	@actual="$$($(PROTOC) --version 2>/dev/null | awk '{ print $$2 }')"; \
	if [[ "$$actual" != "$(PROTOC_VERSION)" ]]; then \
		echo "protoc $(PROTOC_VERSION) is required; found $${actual:-unknown}"; \
		exit 1; \
	fi

.PHONY: legacy-exclusions-check
# Ensure the temporary exclusion cannot silently absorb handwritten tests.
legacy-exclusions-check:
	@test -f quality/legacy-exclusions.yml
	@files="$$(find internal/service -maxdepth 1 -type f -name '*_client_test.go' -print)"; \
	if [[ -z "$$files" ]]; then \
		echo "legacy live-client classification matched no files"; \
		exit 1; \
	fi; \
	unowned="$$(while IFS= read -r file; do \
		head -n 1 "$$file" | grep -Fqx '// Code generated by https://github.com/go-dev-frame/sponge' || echo "$$file"; \
	done <<< "$$files")"; \
	if [[ -n "$$unowned" ]]; then \
		echo "handwritten or unowned tests matched the legacy live-client exclusion:"; \
		echo "$$unowned"; \
		exit 1; \
	fi; \
	unclassified="$$(find internal/service -maxdepth 1 -type f -name '*_test.go' \
		! -name '*_client_test.go' ! -name 'service_test.go' -print)"; \
	if [[ -n "$$unclassified" ]]; then \
		echo "unclassified tests would be skipped by the service-package exclusion:"; \
		echo "$$unclassified"; \
		exit 1; \
	fi; \
	declared="$$(grep -E '^func (Test|Benchmark|Example)' internal/service/service_test.go || true)"; \
	if [[ "$$declared" != 'func TestRegisterAllService(t *testing.T) {' ]]; then \
		echo "service_test.go contains an unclassified test, benchmark, or example:"; \
		echo "$$declared"; \
		exit 1; \
	fi

.PHONY: coverage-policy-test
# Prove the coverage checker rejects regressions, package drift, and unowned exclusions.
coverage-policy-test:
	@bash quality/check_coverage_test.sh

.PHONY: coverage-check
# Measure the hermetic unit package set without writing an artifact into the worktree.
coverage-check: legacy-exclusions-check coverage-policy-test
	@tmp="$$(mktemp -d /tmp/netbox-go-coverage.XXXXXX)"; \
	profile="$$tmp/coverage.out"; \
	measured="$$tmp/measured-packages"; \
	all="$$tmp/all-packages"; \
	cleanup() { \
		rm -f "$$profile" "$$measured" "$$all"; \
		rmdir "$$tmp"; \
	}; \
	trap cleanup EXIT; \
	GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) list ./... | LC_ALL=C sort >"$$all"; \
	grep -Fqx '$(LEGACY_LIVE_CLIENT_PACKAGE)' "$$all"; \
	grep -Fvx '$(LEGACY_LIVE_CLIENT_PACKAGE)' "$$all" >"$$measured"; \
	test -s "$$measured"; \
	packages="$$(tr '\n' ' ' <"$$measured")"; \
	GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) test -short -count=1 \
		-covermode=atomic -coverprofile="$$profile" $$packages; \
	bash quality/check_coverage.sh quality/coverage-baseline.tsv "$$profile" \
		"$$measured" "$$all" quality/legacy-exclusions.yml

.PHONY: cover
# Backward-compatible alias for the non-mutating coverage gate.
cover: coverage-check

.PHONY: check
# Run the complete, non-mutating, self-contained backend V0 quality gate.
check: fmt-check vet lint unit coverage-check build legacy-client-compile legacy-proto-compile api-proto-check

.PHONY: ci-lint
# Backward-compatible non-mutating alias for formatting and lint checks.
ci-lint: fmt-check lint


.PHONY: graph
# Generate interactive visual function dependency graphs
graph:
	@echo "generating graph ......"
	@cp -f cmd/netbox_go/main.go .
	go-callvis -skipbrowser -format=svg -nostd -file=netbox_go netbox-go
	@rm -f main.go netbox_go.gv



.PHONY: proto
# Generate *.go and template code by proto files, the default is all the proto files in the api directory. you can specify the proto file, multiple files are separated by commas, e.g. make proto FILES=api/user/v1/user.proto
proto:
	@bash scripts/protoc.sh $(FILES)
	go mod tidy
	@gofmt -s -w .


.PHONY: proto-doc
# Generate doc from *.proto files
proto-doc:
	@bash scripts/proto-doc.sh


.PHONY: build
# Compile every backend package without writing an artifact into the worktree.
build:
	@GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" $(GO) build ./...

.PHONY: build-linux
# Build the deployable Linux binary into cmd/netbox_go.
build-linux:
	@echo "building 'netbox_go', linux binary file will output to 'cmd/netbox_go'"
	@cd cmd/netbox_go && GOCACHE="$(GOCACHE)" GOFLAGS="$(GOFLAGS)" \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build



.PHONY: run
# Build and run service, you can specify the configuration file, e.g. make run Config=configs/dev.yml
run:
	@bash scripts/run.sh $(Config)


.PHONY: run-nohup
# Run service with nohup in local, e.g. make run-nohup CMD=start, you can specify the configuration file, e.g. make run-nohup CMD=start Config=configs/dev.yml, if you want to stop the service, e.g. make run-nohup CMD=stop
CMD ?= start
run-nohup:
	@bash scripts/run-nohup.sh $(CMD) $(Config)


.PHONY: run-docker
# Run service in local docker, if you want to update the service, run the make run-docker command again
run-docker: image-build-local
	@bash scripts/deploy-docker.sh


.PHONY: binary-package
# Packaged binary files
binary-package: build-linux
	@bash scripts/binary-package.sh


.PHONY: deploy-binary
# Deploy with an SSH agent/key, e.g. make deploy-binary REMOTE_USER=root REMOTE_HOST=server.example
deploy-binary: binary-package
	@test -n "$(REMOTE_USER)" || (echo "REMOTE_USER is required" >&2; exit 2)
	@test -n "$(REMOTE_HOST)" || (echo "REMOTE_HOST is required" >&2; exit 2)
	@bash scripts/deploy-binary.sh "$(REMOTE_USER)" "$(REMOTE_HOST)"


.PHONY: image-build-local
# Build image for local docker, tag=latest, use binary files to build
image-build-local: build-linux
	@bash scripts/image-build-local.sh


.PHONY: image-build
# Build image for remote repositories, use binary files to build, e.g. make image-build REPO_HOST=addr TAG=latest
image-build:
	@if [ -z "$(REPO_HOST)" ]; then \
		echo "ERROR: REPO_HOST is required. Usage: make image-build REPO_HOST=xxx"; \
		exit 1; \
	fi
	@bash scripts/image-build.sh $(REPO_HOST) $(TAG)


.PHONY: image-build2
# Build image for remote repositories, phase II build, e.g. make image-build2 REPO_HOST=addr TAG=latest
image-build2:
	@if [ -z "$(REPO_HOST)" ]; then \
		echo "ERROR: REPO_HOST is required. Usage: make image-build2 REPO_HOST=xxx"; \
		exit 1; \
	fi
	@bash scripts/image-build2.sh $(REPO_HOST) $(TAG)


.PHONY: image-push
# Push docker image to remote repositories, e.g. make image-push REPO_HOST=addr TAG=latest
image-push:
	@if [ -z "$(REPO_HOST)" ]; then \
		echo "ERROR: REPO_HOST is required. Usage: make image-push REPO_HOST=xxx TAG=latest"; \
		exit 1; \
	fi
	@bash scripts/image-push.sh $(REPO_HOST) $(TAG)


.PHONY: deploy-k8s
# Deploy service to k8s
deploy-k8s:
	@bash scripts/deploy-k8s.sh


.PHONY: image-build-rpc-test
# Build grpc test image for remote repositories, e.g. make image-build-rpc-test REPO_HOST=addr TAG=latest
image-build-rpc-test:
	@bash scripts/image-rpc-test.sh $(REPO_HOST) $(TAG)


.PHONY: patch
# Patch some dependent code. 1. Add database initialization code, e.g. make patch TYPE=mysql, also supports mongodb, postgresql, sqlite. 2. Add dependent types.proto file, e.g. make patch TYPE=types-pb
patch:
	@bash scripts/patch.sh $(TYPE)


.PHONY: copy-proto
# Copy proto file from the grpc server directory, multiple directories or proto files separated by commas. default is to copy all proto files, e.g. make copy-proto SERVER=yourServerDir, copy specified proto files, e.g. make copy-proto SERVER=yourServerDir PROTO_FILE=yourProtoFile1,yourProtoFile2
copy-proto:
	@sponge patch copy-proto --server-dir=$(SERVER) --proto-file=$(PROTO_FILE)


.PHONY: modify-proto-pkg-name
# Modify the 'package' and 'go_package' names of all proto files in the 'api' directory
modify-proto-pkg-name:
	@sponge patch modify-proto-package --dir=api --server-dir=.


.PHONY: update-config
# Update internal/config code base on yaml file
update-config:
	@sponge config --server-dir=.


.PHONY: clean
# Clean binary file, cover.out, template file
clean:
	@rm -vrf cmd/netbox_go/netbox_go*
	@rm -vrf cover.out
	@rm -vrf main.go netbox_go.gv
	@rm -vrf internal/ecode/*.go.gen*
	@rm -vrf internal/routers/*.go.gen*
	@rm -vrf internal/handler/*.go.gen*
	@rm -vrf internal/service/*.go.gen*
	@rm -rf netbox_go-binary.tar.gz
	@echo "clean finished"


# Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[1;36m  %-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
