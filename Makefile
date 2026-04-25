.PHONY: build test vet fmt

build:
	go build ./cmd/kyn

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

