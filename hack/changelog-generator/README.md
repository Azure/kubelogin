# Changelog Generator

This tool automatically generates CHANGELOG.md entries for kubelogin releases by fetching merged pull requests from GitHub and categorizing them appropriately.

## Features

- Fetches merged PRs since the last release tag
- Automatically categorizes PRs into:
  - What's Changed (general changes)
  - Enhancements (new features)
  - Bug Fixes (bug fixes)
  - Maintenance (dependency updates, CVE fixes, chores)
  - Doc Update (documentation changes)
- Identifies and lists new contributors
- Generates a Full Changelog comparison link
- Follows the existing CHANGELOG.md format

## Quick Start

### Via GitHub Actions (Recommended)

1. Go to the [Actions tab](https://github.com/Azure/kubelogin/actions/workflows/update-changelog.yml)
2. Click "Run workflow"
3. Fill in the required inputs:
   - **Version number**: e.g., `0.2.15` (without the 'v' prefix)
   - **Previous version tag**: e.g., `v0.2.14` (with the 'v' prefix)
4. Click "Run workflow"
5. Review and merge the generated PR
6. Trigger the [Release workflow](https://github.com/Azure/kubelogin/actions/workflows/release.yml)

### Via Make Target

```bash
export GITHUB_TOKEN="your_github_token"

# SINCE_TAG is optional; omit it to auto-detect the latest tag
VERSION=0.2.15 make changelog

# Or specify the previous tag explicitly
VERSION=0.2.15 SINCE_TAG=v0.2.14 make changelog
```

This generates a `changelog-entry.md` file that you can manually insert into CHANGELOG.md.

### Running Directly

```bash
export GITHUB_TOKEN="your_github_token"

# SINCE_TAG is optional; omit it to auto-detect the latest tag
go run hack/changelog-generator/main.go \
  --version="0.2.15" \
  --repo="Azure/kubelogin" \
  --output="changelog-entry.md"

# Or specify the previous tag explicitly
go run hack/changelog-generator/main.go \
  --version="0.2.15" \
  --since-tag="v0.2.14" \
  --repo="Azure/kubelogin" \
  --output="changelog-entry.md"
```

## PR Categorization

PRs are categorized first by **GitHub labels**, then by **title patterns**:

| Category | Labels | Title prefixes / patterns |
|---|---|---|
| Bug Fixes | `bug`, `fix` | `fix:`, `bugfix:`, `bug fix:`, `hotfix:` |
| Enhancements | `enhancement`, `feature` | `feat:`, `feature:`, `add support`, `new feature` |
| Maintenance | `maintenance`, `dependencies`, `chore` | `bump `, `update `, `CVE-`, `fix cve`, `chore` |
| Doc Update | `documentation`, `docs` | `docs:`, `doc:`, `documentation`, `install doc` |
| What's Changed | *(default)* | *(everything else)* |

### New Contributor Detection

A contributor is marked as "new" if they have a merged PR in the current release but **no** merged PRs before the previous release tag.

## Example Output

```markdown
## [0.2.15]

### What's Changed

* Add new authentication method by @username in https://github.com/Azure/kubelogin/pull/123

### Enhancements

* Add Y support by @username in https://github.com/Azure/kubelogin/pull/124

### Bug Fixes

* Fix nil pointer in cache.Replace by @username in https://github.com/Azure/kubelogin/pull/127

### Maintenance

* Bump Go to 1.24.12 by @dependabot in https://github.com/Azure/kubelogin/pull/125

### Doc Update

* Update installation guide by @username in https://github.com/Azure/kubelogin/pull/126

### New Contributors

* @newuser made their first contribution in https://github.com/Azure/kubelogin/pull/123

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.2.14...v0.2.15
```

## Integration with Release Workflow

1. **Generate** changelog entry (this tool) → creates a PR
2. **Merge** the changelog PR
3. **Trigger** the [Release workflow](https://github.com/Azure/kubelogin/actions/workflows/release.yml)
   - Reads version from CHANGELOG.md
   - Creates a draft release
   - Builds binaries for all platforms
   - Publishes artifacts

## Troubleshooting

**"No PRs found"**
- Verify the tag exists: `git tag -l | grep <tag>`
- Check that PRs were merged after the `since_tag` date

**"API rate limit exceeded"**
- Ensure `GITHUB_TOKEN` is set with `repo` scope
- Wait for rate limit reset (typically 1 hour)

**Wrong categorization**
- Add appropriate labels to PRs before running the tool
- Or manually edit the generated changelog before merging

## Requirements

- `GITHUB_TOKEN` environment variable with `repo` scope
- Go 1.24 or later

