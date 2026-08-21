.PHONY: build test e2e vet fmt lint hooks ci

build:
	go build ./cmd/kyn

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
