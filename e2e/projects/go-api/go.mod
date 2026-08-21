// This is a standalone module so the fixture's Go source files are real,
// compilable Go (kyn's glob patterns match "*.go" files) without being
// picked up by the parent kyn module's own `go build`/`go vet`/`go test
// ./...` — it is example content for the e2e CLI harness, not part of
// kyn itself.
module kyn-e2e-fixture-go-api

go 1.22
