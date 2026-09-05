# Kyn

Kyn turns the file relationships your team already checks in review into a fast,
deterministic local and CI policy.

## Highlights

- Keep Storybook stories aligned with component changes.
- Require tests when handlers or services change.
- Warn when Terraform modules change without their documentation.
- Emit machine-readable results for CI without a service, daemon, or plugin.

## Start in five minutes

```bash
brew tap dills122/tap
brew install --cask dills122/tap/kyn

cd your-repository
kyn init --preset web-ui
kyn check -c kyn.config.yaml
```

Not using Homebrew? [Choose another install method](install.md).

## What Kyn evaluates

Kyn collects changed files, groups them into logical **families**, resolves
related files called **kin**, and evaluates your **rules**. It reports the same
ordered result locally and in CI.

For example, when `button.component.ts` changes, a rule can require
`button.stories.ts` in the same change set. Kyn tells the developer exactly
which file is expected and exits with a stable code CI can act on.

## Explore

- [Getting Started](getting-started.md)
- [How Kyn works](concepts.md)
- [Real-world recipes](recipes/frontend.md)
- [CI adoption](ci.md)
- [What's new in v0.1.3](releases/v0.1.3.md)
