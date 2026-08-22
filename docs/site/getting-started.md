# Getting Started

This walkthrough takes a repository from no policy to a useful local check.

## 1. Install

Use the [platform recommendation](install.md), then confirm the binary:

```bash
kyn --version
```

## 2. Generate a starter policy

From the repository root:

```bash
kyn init --preset web-ui
```

This writes `kyn.config.yaml`. Choose `api`, `proto`, or `iac` instead
when one of those layouts is closer to your project.

## 3. Adapt the paths

Open `kyn.config.yaml` and make the family globs match your repository. A
family identifies source files and resolves their related files—called kin.
Rules then say what must happen when one of those files changes.

Start with one relationship your team already enforces in review, such as:

- component source → Storybook story
- API handler → test
- Terraform module → module README

See [Configure families and rules](config.md) for each field.

## 4. Preview resolution

Use one real source path from your repository:

```bash
kyn check -c kyn.config.yaml \
  --files path/to/example.component.ts \
  --dry-run-resolve
```

This shows the family instance and kin paths without evaluating policy. Fix the
globs and templates until those paths are right.

## 5. Run the policy

Pass the source path and its related file:

```bash
kyn check -c kyn.config.yaml \
  --files path/to/example.component.ts,path/to/example.stories.ts \
  --show-passes
```

Inside a Git repository, the shortest normal command is:

```bash
kyn check -c kyn.config.yaml
```

With no explicit input mode, Kyn compares `origin/main...HEAD`. Configure a
different default with `KYN_BASE_REF` and `KYN_HEAD_REF`, or pass
`--base` and `--head` explicitly.

## 6. Understand a failure

```bash
kyn explain -c kyn.config.yaml --files path/to/example.component.ts
```

`explain` reports why each applicable rule passed or failed. Once the policy
feels right locally, follow the [CI adoption recipe](ci.md).
