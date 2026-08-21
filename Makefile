.PHONY: build test e2e vet fmt lint hooks ci

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X kyn/internal/cli.Version=$(VERSION) -X kyn/internal/cli.Commit=$(COMMIT) -X kyn/internal/cli.Date=$(DATE)

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

hooks:
	./scripts/setup-git-hooks.sh

ci: lint test build e2e
