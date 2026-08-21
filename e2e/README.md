# End-to-End Tests

Black-box tests that build the real `kyn` binary and run it as a
subprocess against realistic example projects, the way a CI pipeline
would. These complement (not replace) the unit tests under `internal/*`,
which exercise packages in-process with synthetic inputs.

## Layout

```
e2e/
  e2e_test.go        # builds the binary once, runs project scenarios
  workflows_test.go  # multi-step docs workflows: preset init, config migrate
  projects/
    <project-name>/
      kyn.config.yaml
      scenarios.json   # CLI invocations + expected exit code / output
      testdata/*.golden
      ...real fixture source files...
```

Each `projects/<name>` directory is a small, realistic repo: a v1 or v2
Kyn config plus source files that a `kyn check` run would actually see.
`scenarios.json` lists the CLI invocations to run against it:

```json
{
  "name": "story-not-updated-fails",
  "args": ["check", "-c", "kyn.config.yaml", "--files", "...", "--format", "text"],
  "wantExitCode": 1,
  "golden": "story-not-updated-fails.golden"
}
```

`golden` compares full stdout against a checked-in fixture. Use
`contains` / `notContains` instead when exact output isn't worth
pinning.

## Running

```bash
make e2e
# or
go test ./e2e/... -v
```

## Updating goldens

After an intentional output change:

```bash
go test ./e2e/... -update
```

Review the resulting diff before committing.

## Adding a project

1. Create `e2e/projects/<name>/` with a `kyn.config.yaml` and enough real
   source files to exercise it (kin resolution checks the filesystem, so
   fixture files must actually exist).
2. Add `scenarios.json` describing the CLI runs to assert.
3. Run `go test ./e2e/... -update` once to generate goldens, then review
   them.
