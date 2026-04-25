.PHONY: build test vet fmt lint hooks ci

build:
	go build ./cmd/kyn

test:
	go test ./...

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

ci: lint test build
