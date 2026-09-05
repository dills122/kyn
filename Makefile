.PHONY: build test e2e vet fmt lint coverage module-path hooks ci

MODULE_PATH := github.com/dills122/kyn
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X $(MODULE_PATH)/internal/cli.Version=$(VERSION) -X $(MODULE_PATH)/internal/cli.Commit=$(COMMIT) -X $(MODULE_PATH)/internal/cli.Date=$(DATE)

# Keep this in sync with the "Enforce minimum internal coverage" step in
# .github/workflows/ci.yml.
COVERAGE_THRESHOLD := 80.0

build:
	go build -ldflags "$(LDFLAGS)" -o ./bin/kyn ./cmd/kyn

test:
	go test ./cmd/... ./internal/...

e2e:
	go test ./e2e/... -v

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

lint:
	FILES="$$(gofmt -l cmd internal)"; \
	if [ -n "$$FILES" ]; then \
		echo "Files need gofmt:"; \
		echo "$$FILES"; \
		exit 1; \
	fi
	go vet ./...
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found on PATH; install it to match the CI lint gate: https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	fi
	golangci-lint run

# Mirrors the "coverage" job in .github/workflows/ci.yml: same test scope,
# same threshold, so a local pass here means that job will pass too.
coverage:
	go test ./internal/... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tee coverage.txt
	TOTAL="$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}')"; \
	echo "Total coverage: $${TOTAL}%"; \
	awk "BEGIN {exit !($${TOTAL} >= $(COVERAGE_THRESHOLD))}"

module-path:
	test "$$(go list -m)" = "$(MODULE_PATH)"

hooks:
	./scripts/setup-git-hooks.sh

# Mirrors the checks GitHub Actions enforces in .github/workflows/ci.yml
# (lint job, test matrix, coverage job, e2e job): a passing `make ci` should
# mean CI passes too. It does not reproduce the release-config, docs,
# cross-platform build-matrix, or container smoke-test jobs, which need
# tooling (goreleaser, mkdocs, Docker/QEMU) this target does not assume.
ci: module-path lint test coverage build e2e
