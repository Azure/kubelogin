# Changelog Generator

This tool automatically generates CHANGELOG.md entries for kubelogin releases by fetching merged pull requests from GitHub and categorizing them appropriately.

## Features

- Fetches merged PRs since the last release tag
- Automatically categorizes PRs into:
  - What's Changed (general changes)
  - Enhancements (new features)
  - Maintenance (dependency updates, CVE fixes, chores)
  - Doc Update (documentation changes)
- Identifies and lists new contributors
- Generates a Full Changelog comparison link
- Follows the existing CHANGELOG.md format

## Usage

### Via GitHub Actions (Recommended)

The easiest way to generate a changelog is through the GitHub Actions workflow:

1. Go to the [Actions tab](https://github.com/Azure/kubelogin/actions/workflows/update-changelog.yml)
2. Click "Run workflow"
3. Fill in the required inputs:
   - **Version number**: e.g., `0.2.15` (without the 'v' prefix)
   - **Previous version tag**: e.g., `v0.2.14` (with the 'v' prefix)
4. Click "Run workflow"

The workflow will:
- Generate the changelog entry
- Update CHANGELOG.md
- Create a pull request with the changes

### Manual Usage

You can also run the tool locally:

```bash
# Set your GitHub token
export GITHUB_TOKEN="your_github_token"

# Run the generator
go run hack/changelog-generator/main.go \
  --version="0.2.15" \
  --since-tag="v0.2.14" \
  --repo="Azure/kubelogin" \
  --output="changelog-entry.md"
```

This will create a `changelog-entry.md` file with the generated entry that you can manually insert into CHANGELOG.md.

## How It Works

### PR Categorization

The tool categorizes PRs based on:

1. **GitHub Labels**: Checks for labels like `maintenance`, `enhancement`, `documentation`, etc.
2. **Title Patterns**: Analyzes PR titles for common prefixes:
   - `bump `, `update `, `CVE-` → Maintenance
   - `docs:`, `doc:` → Documentation
   - `feat:`, `feature:` → Enhancement

### New Contributor Detection

The tool identifies new contributors by:
1. Fetching all historical PRs before the previous release tag
2. Building a list of existing contributors
3. Comparing current PR authors against this list
4. Marking contributors who don't appear in the historical list as "new"

## Customization

You can customize the categorization logic by editing the `categorizeByLabelsAndTitle` function in `main.go`.

## Integration with Release Workflow

After the changelog PR is merged:
1. The updated CHANGELOG.md is now in the main branch
2. Trigger the [Release workflow](https://github.com/Azure/kubelogin/actions/workflows/release.yml)
3. The release workflow will:
   - Read the version from CHANGELOG.md
   - Create a draft release
   - Build binaries
   - Publish artifacts

## Requirements

- GitHub Personal Access Token with `repo` scope
- Go 1.24 or later

## Example Output

```markdown
## [0.2.15]

### What's Changed

* Add new authentication method by @username in https://github.com/Azure/kubelogin/pull/123

### Maintenance

* Bump Go to 1.24.12 by @dependabot in https://github.com/Azure/kubelogin/pull/124

### New Contributors

* @newuser made their first contribution in https://github.com/Azure/kubelogin/pull/123

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.2.14...v0.2.15
```
