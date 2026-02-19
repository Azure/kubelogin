# Changelog Automation Quick Reference

This document provides a quick reference for using the automated changelog generation system.

## Quick Start

### Generate Changelog via GitHub Actions

1. Go to: https://github.com/Azure/kubelogin/actions/workflows/update-changelog.yml
2. Click "Run workflow"
3. Fill in:
   - **version**: `0.2.15` (without 'v')
   - **since_tag**: `v0.2.14` (with 'v')
4. Click "Run workflow"
5. Wait for PR to be created
6. Review and merge the PR

### Generate Changelog Locally

```bash
export GITHUB_TOKEN="your_github_token"

go run hack/changelog-generator/main.go \
  --version="0.2.15" \
  --since-tag="v0.2.14" \
  --output="changelog-entry.md"
```

## PR Categorization Rules

### Maintenance
- Titles starting with: `bump `, `update `, `chore`
- Titles containing: `CVE-`, `dependencies`
- Labels: `maintenance`, `dependencies`, `chore`

### Documentation
- Titles starting with: `docs:`, `doc:`
- Titles containing: `documentation`, `install doc`
- Labels: `documentation`, `docs`

### Enhancements
- Titles starting with: `feat:`, `feature:`
- Titles containing: `add support`, `new feature`
- Labels: `enhancement`, `feature`

### What's Changed
- Everything else not matching above categories

## New Contributor Detection

A contributor is marked as "new" if they:
- Have a merged PR in the current release
- Have NO merged PRs before the previous release tag

## Output Format

```markdown
## [0.2.15]

### What's Changed

* Feature X by @username in https://github.com/Azure/kubelogin/pull/123

### Enhancements

* Add Y support by @username in https://github.com/Azure/kubelogin/pull/124

### Maintenance

* Bump dependency Z by @dependabot in https://github.com/Azure/kubelogin/pull/125

### Doc Update

* Update installation guide by @username in https://github.com/Azure/kubelogin/pull/126

### New Contributors

* @newuser made their first contribution in https://github.com/Azure/kubelogin/pull/123

**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.2.14...v0.2.15
```

## Workflow Integration

1. **Changelog Generation** (this tool)
   - Generates CHANGELOG.md entry
   - Creates PR for review

2. **Merge Changelog PR**
   - Review and merge the changelog PR

3. **Release Workflow** (existing)
   - Reads version from CHANGELOG.md
   - Creates draft release
   - Builds binaries
   - Publishes artifacts

## Troubleshooting

### "No PRs found"
- Check that PRs were merged after the `since_tag` date
- Verify the tag exists: `git tag -l | grep <tag>`

### "API rate limit exceeded"
- Ensure `GITHUB_TOKEN` is set
- Token needs `repo` scope
- Wait for rate limit reset (typically 1 hour)

### Wrong categorization
- Add appropriate labels to PRs before running the tool
- Or manually edit the generated changelog after review

## Tips for Better Changelogs

1. **Use clear PR titles** - They appear directly in the changelog
2. **Add labels to PRs** - Helps with automatic categorization
3. **Follow conventional commits** - Use prefixes like `feat:`, `docs:`, `chore:`
4. **Review before merging** - Always review the generated changelog PR

## Links

- [Full Documentation](../docs/book/src/development.md#releases)
- [Tool Source Code](main.go)
- [GitHub Actions Workflow](../../.github/workflows/update-changelog.yml)
