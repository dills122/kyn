# Recipe: Keep Terraform Modules and Their READMEs Aligned

## The review problem

A module's inputs, resources, or outputs change while its usage documentation
stays stale. The right rollout policy may be advisory at first rather than an
immediate hard gate.

Kyn can report the missing README update as a warning, then let each pipeline
choose whether warnings block.

## Example layout

```text
terraform/network/
├── main.tf
└── README.md
```

## Policy

```yaml
version: 2

families:
  - id: terraform-module
    groups:
      source:
        include:
          - "terraform/**/*.tf"
      docs:
        include:
          - "terraform/**/README.md"
    kin:
      readme: "{dir}/README.md"

rules:
  - id: terraform-docs-sync
    family: terraform-module
    severity: warn
    if:
      changedAny: [source]
      kinExists: [readme]
    assert:
      kinChanged: [readme]
    message: "Terraform module changed but module README did not."
```

## Advisory rollout

```bash
kyn check -c kyn.config.yaml \
  --files terraform/network/main.tf
```

The rule result fails and names `terraform/network/README.md` as expected, but
the command exits 0 because the default threshold blocks only errors. This is a
useful first rollout: teams see the policy without breaking every pull request.

## Strict rollout

After the signal proves useful:

```bash
kyn check -c kyn.config.yaml \
  --files terraform/network/main.tf \
  --fail-on warn
```

The same result now exits 1. You can make the policy intrinsically strict by
changing its severity to `error`, or keep it at `warn` and let different
pipelines choose their threshold.

## Pass with documentation

```bash
kyn check -c kyn.config.yaml \
  --files terraform/network/main.tf,terraform/network/README.md \
  --fail-on warn
```

The README is part of the change set, so the rule passes and Kyn exits 0.

## Adapt the recipe

- Add kin for generated examples, schemas, or module changelogs.
- Use `groups.source.exclude` for generated `.tf` files.
- Split modules into families when repository areas have different documentation
  requirements.

The repository's
[executable Terraform fixture](https://github.com/dills122/kyn/tree/main/e2e/projects/terraform-iac)
tests advisory, strict, and passing outcomes.
