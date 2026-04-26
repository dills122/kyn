# Presets

Kyn v2 ships starter presets through `kyn init --preset`.

## Available Presets

### `web-ui`

Best for component-driven frontend repos.

Tracks:

- component source files
- Storybook stories
- component tests

Command:

```bash
kyn init --preset web-ui
```

### `api`

Best for backend/API services with handler or service layer tests.

Tracks:

- Go handler/service source files
- Go test files

Command:

```bash
kyn init --preset api
```

### `proto`

Best for protobuf-first repos that need generated artifact sync.

Tracks:

- `.proto` contract files
- generated Go protobuf output

Command:

```bash
kyn init --preset proto
```

### `iac`

Best for Terraform-style infrastructure modules.

Tracks:

- Terraform source files
- module `README.md`

Command:

```bash
kyn init --preset iac
```

## Recommended Workflow

1. Generate the nearest preset:
   `kyn init --preset <preset>`
2. Adjust glob patterns and kin templates for your repo.
3. Run:
   `kyn check --dry-run-resolve`
4. Add CI with one of the examples from [ci.md](ci.md).
