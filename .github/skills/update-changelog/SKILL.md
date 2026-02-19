---
name: update-changelog
description: >
  Generate and update CHANGELOG.md for a new kubelogin release. Use this
  when asked to prepare release notes or update the changelog for a new
  version. Fetches merged pull requests since the previous release,
  categorizes them, identifies new contributors, and inserts a formatted
  entry at the top of CHANGELOG.md.
---

## Overview

The `make changelog` target runs `hack/changelog-generator/main.go`, which:

1. Calls `gh api repos/Azure/kubelogin/releases/latest` to determine the
   previous release tag (when `SINCE_TAG` is not supplied).
2. Calls `gh api repos/Azure/kubelogin/commits/<tag>` to get the tag date.
3. Fetches all merged pull requests since that date via
   `gh api --paginate repos/Azure/kubelogin/pulls?state=closed&...`.
4. Categorizes each PR by GitHub label, then by title prefix:
   - **Bug Fixes** — label `bug`/`fix`; prefix `fix:`, `bugfix:`, `hotfix:`
   - **Enhancements** — label `enhancement`/`feature`; prefix `feat:`
   - **Maintenance** — label `dependencies`/`chore`; prefix `bump `, `update `
   - **Doc Update** — label `documentation`/`docs`; prefix `docs:`
   - **What's Changed** — everything else
5. Identifies first-time contributors by comparing PR authors against all
   prior merged PR authors.
6. Writes a formatted entry to `changelog-entry.md`.
7. Inserts that entry immediately after the `# Change Log` header in
   `CHANGELOG.md`.

## Steps to follow

1. Determine the new version number (e.g. `0.2.15`) and, optionally, the
   previous tag to compare from (e.g. `v0.2.14`).
   - If the previous tag is not provided, the tool will auto-detect the
     latest stable release.

2. Generate the changelog entry and update `CHANGELOG.md`:

   ```bash
   # SINCE_TAG is optional – omit to auto-detect the latest release tag
   VERSION=0.2.15 make changelog
   # or explicitly:
   VERSION=0.2.15 SINCE_TAG=v0.2.14 make changelog
   ```

   Authentication is handled automatically by the `gh` CLI. Ensure you
   are authenticated (`gh auth login`) or that `GH_TOKEN`/`GITHUB_TOKEN`
   is set in the environment.

3. Review the updated `CHANGELOG.md` and edit entries as needed for
   clarity before committing.

4. After the changelog is merged to the default branch, trigger the
   [Release workflow](../../.github/workflows/release.yml) to create the
   GitHub release and build binaries.
