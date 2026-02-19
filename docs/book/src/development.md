# Development

## Prerequisites

### System Dependencies

kubelogin uses secure token storage that requires platform-specific libraries:

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install libsecret-1-0 libsecret-1-dev
```

#### Linux (CentOS/RHEL/Fedora)
```bash
# CentOS/RHEL
sudo yum install libsecret-devel

# Fedora
sudo dnf install libsecret-devel
```

#### macOS
No additional dependencies required (uses Keychain)

#### Windows
No additional dependencies required (uses Windows Credential Manager)

### Go Dependencies
- Go 1.23 or later
- Make

## Building

```bash
make build
```

## Testing

```bash
make test
```

**Note**: Tests require the system dependencies listed above. If you encounter errors related to `libsecret-1.so` or "encrypted storage isn't possible", ensure the libsecret library is installed.

## Releases

### Automated Changelog Generation

The project includes an automated changelog generation tool that creates properly formatted CHANGELOG.md entries from merged pull requests.

#### Using the GitHub Actions Workflow

1. Navigate to the [Update Changelog workflow](https://github.com/Azure/kubelogin/actions/workflows/update-changelog.yml)
2. Click "Run workflow"
3. Provide the required inputs:
   - **Version number**: The new version (e.g., `0.2.15`) without the 'v' prefix
   - **Previous version tag**: The tag to compare from (e.g., `v0.2.14`) with the 'v' prefix
4. Click "Run workflow"

The workflow will:
- Fetch all merged PRs since the previous version
- Categorize them (What's Changed, Enhancements, Bug Fixes, Maintenance, Doc Update)
- Identify new contributors
- Generate a formatted changelog entry
- Create a pull request with the updated CHANGELOG.md

#### Running Locally

You can generate a changelog entry locally using the `gh` CLI and `make`:

```bash
# Authenticate with the gh CLI (one-time setup)
gh auth login

VERSION=0.2.15 make changelog
```

The tool uses `gh api` for all GitHub API calls, so only `gh auth login`
is required. In CI the `GH_TOKEN` environment variable is used instead.

See [hack/changelog-generator/README.md](../../hack/changelog-generator/README.md) for more details.

### Release Process

After the changelog is updated:

1. Review and merge the changelog PR
2. Trigger the [Release workflow](https://github.com/Azure/kubelogin/actions/workflows/release.yml)
3. The workflow will:
   - Read the version from CHANGELOG.md
   - Create a draft GitHub release
   - Build binaries for all platforms
   - Upload artifacts to the release
