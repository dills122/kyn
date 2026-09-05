# Recipe: Keep Components, Stories, and Design Metadata in Sync

## The review problem

A component changes, but its Storybook story is left behind. Reviewers repeat
the same request, and visual regressions arrive later. Some components also have
design metadata that should trigger a downstream review or publish step.

Kyn can block the missing story update and emit a separate design-review flag.

## Example layout

```text
src/ui/button/
├── button.component.ts
├── button.component.html
├── button.stories.ts
└── figma.button.json
```

## Policy

```yaml
version: 2

families:
  - id: web-component
    groups:
      source:
        include:
          - "src/**/*.component.ts"
          - "src/**/*.component.html"
      stories:
        include:
          - "src/**/*.stories.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
      figma: "{dir}/figma.{base}.json"

rules:
  - id: story-sync
    family: web-component
    severity: error
    if:
      changedAny: [source]
      kinExists: [story]
    assert:
      kinChanged: [story]
    message: "Component changed but its existing story did not."

  - id: design-review-signal
    family: web-component
    severity: info
    if:
      changedAny: [source]
      kinExists: [figma]
    actions:
      emit: [designReviewRequired]
    message: "A component with a Figma sidecar changed."
```

`kinExists` makes the story rule apply only to components that already have a
story. Remove that condition and use `assert.kinExists` if your policy is that
every component must create one.

## See the failure locally

```bash
kyn check -c kyn.config.yaml \
  --files src/ui/button/button.component.ts,src/ui/button/button.component.html
```

The important part of the report is:

```text
[ERROR] story-sync
Status: fail
Expected files:
  - src/ui/button/button.stories.ts

Flags:
  - designReviewRequired
```

Kyn exits 1 because the failed rule has `error` severity. The emitted flag is
still present for machine consumers.

## Make it pass

```bash
kyn check -c kyn.config.yaml \
  --files src/ui/button/button.component.ts,src/ui/button/button.component.html,src/ui/button/button.stories.ts
```

The story assertion now passes and the command exits 0. The design-review flag
remains because the component has a Figma sidecar.

## Put the flag to work

JSON output contains a sorted `flags` array:

```bash
kyn check -c kyn.config.yaml --format json
```

A later CI step can inspect `designReviewRequired` and select a design-review
or publishing job. Kyn only reports the signal; it does not call external APIs.

## Adapt the recipe

- Add a `spec` kin and warning rule for component tests.
- Exclude generated component directories under `groups.source.exclude`.
- Use separate families when React, Vue, or Angular areas follow different
  naming conventions.

The repository's
[executable frontend-recipe fixture](https://github.com/dills122/kyn/tree/main/e2e/projects/frontend-recipe)
runs this v2 policy against both the failing and passing command examples above.
