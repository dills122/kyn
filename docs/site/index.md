# Kyn

Kyn catches a common review omission: a source file changed, but its related
test, story, generated file, or documentation did not join the same change set.
Define the relationship once, then enforce it locally and in CI.

Kyn checks which paths changed and whether related files exist. It does not edit
files, inspect their contents, or run tests. Your test suite and reviewers still
decide whether the accompanying change is correct.

## Start with a rule you can prove

```bash
brew tap dills122/tap
brew install --cask dills122/tap/kyn
kyn --version
```

Then follow [Your first check](getting-started.md). The guided exercise uses two
disposable files to produce one known match, one intentional failure, an
explanation, and one pass before Git history or repository conventions enter
the picture.

Not using Homebrew? [Choose another install method](install.md).

## Bring the check into your repository

After the guided exercise, generate the starter closest to your layout:

```bash
cd your-repository
kyn init --preset web-ui
```

Adapt one real path, preview its resolved kin, and deliberately test both
failure and success before adopting the policy in CI. Presets are starting
points with concrete layout assumptions, not framework detection.

## Where Kyn helps

- Keep Storybook stories aligned with component changes.
- Require tests when handlers or services change.
- Warn when Terraform modules change without their documentation.
- Emit machine-readable results for CI without a service, daemon, or plugin.

## What Kyn evaluates

Kyn collects changed files, groups them into logical **families**, resolves
related files called **kin**, and evaluates your **rules**. It reports the same
ordered result locally and in CI.

For example, when `button.component.ts` changes, a rule can require
`button.stories.ts` in the same change set. Kyn tells the developer exactly
which file is expected and exits with a stable code CI can act on.

## Explore

- [Your first check](getting-started.md)
- [How Kyn works](concepts.md)
- [Choose and adapt a preset](presets.md)
- [Real-world recipes](recipes/frontend.md)
- [CI adoption](ci.md)
- [What's new in v0.1.3](releases/v0.1.3.md)
