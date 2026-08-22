# Presets

Presets create a valid version 2 config that is close to a common repository
shape. They are starting points, not frameworks or hidden runtime behavior.

| Preset | Source relationship | Generated policy |
| --- | --- | --- |
| `web-ui` | Components, stories, tests | Story error and test warning |
| `api` | Go handlers/services and tests | Missing test change is an error |
| `proto` | Protobuf contracts and generated Go | Missing generated change is an error |
| `iac` | Terraform and module README | Missing docs change is a warning |

## Generate a config

```bash
kyn init --preset api
```

The default preset is `web-ui`. Write a different path with:

```bash
kyn init --preset iac --config .kyn/kyn.config.yaml
```

Kyn refuses to replace an existing file unless `--force` is present.

## Customize safely

1. Change `groups.source.include` to match real source paths.
2. Adjust suffix stripping and kin templates for your naming conventions.
3. Preview one representative path:

   ```bash
   kyn check -c kyn.config.yaml \
     --files path/to/representative-source \
     --dry-run-resolve
   ```

4. Run both a failing and passing change set.
5. Start noisy policies at `warn` before making them strict.

See [Configure families and rules](config.md) for every field and
[real-world recipes](recipes/frontend.md) for complete adaptations.
