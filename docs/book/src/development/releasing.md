# Releasing

Use this process to create and publish a kubelogin release. You need repository write access and permission to run GitHub Actions.

## Release sequence

1. Generate the changelog pull request.
2. Review and merge the changelog pull request.
3. Run the Release workflow from `main`.
4. Review the draft release and its assets.
5. Publish the draft release.
6. Verify the release and downstream workflows.

Do not run the Release workflow until the changelog pull request is merged.

## 1. Generate the changelog pull request

Make sure all intended changes are merged before you generate the changelog.

1. Open the [Update Changelog workflow](https://github.com/Azure/kubelogin/actions/workflows/update-changelog.yml).
2. Select **Run workflow**.
3. Select the `main` branch.
4. Enter the new version without the `v` prefix, such as `0.2.20`.
5. Enter the previous tag with the `v` prefix, such as `v0.2.19`.
6. Run the workflow.

The previous tag is optional. The workflow uses the latest release when this field is empty.

The workflow creates a pull request with the new `CHANGELOG.md` entry. For local generation, see the [changelog generator guide](../../../../hack/changelog-generator/README.md).

## 2. Review and merge the changelog

Review these items before you merge the pull request:

- The first entry uses the exact heading `## [x.y.z]`.
- The entry includes all intended pull requests since the previous tag.
- The entry does not include changes from an earlier release.
- The categories and new-contributor entries are correct.
- The full changelog link compares the previous tag with the new tag.

Edit generated text when a pull request title does not describe the user impact. Merge the pull request into `main` after approval.

## 3. Run the Release workflow

1. Confirm that the changelog pull request is merged.
2. Open the [Release workflow](https://github.com/Azure/kubelogin/actions/workflows/release.yml).
3. Select **Run workflow**.
4. Select the `main` branch.
5. Run the workflow.

The workflow reads the first version from `CHANGELOG.md`. It then does these tasks:

- Creates a draft release named `vX.Y.Z release`.
- Builds Linux, Windows, and macOS binaries.
- Creates platform ZIP archives and SHA-256 files.
- Creates `kubelogin-version.txt`.
- Uploads all release assets to the draft release.

The workflow creates a draft release. It does not publish the release.

## 4. Review the draft release

Open the [kubelogin releases page](https://github.com/Azure/kubelogin/releases) and select the draft release.

Confirm these items:

- The tag is `vX.Y.Z`.
- The release notes match the new changelog entry.
- Every job in the Release workflow succeeded.
- The draft contains eight ZIP archives and eight matching SHA-256 files.
- The draft contains `kubelogin-version.txt`.

You can verify a checksum with the GitHub CLI:

```bash
gh release download vX.Y.Z \
  --repo Azure/kubelogin \
  --pattern 'kubelogin-linux-amd64.zip*'

sha256sum --check kubelogin-linux-amd64.zip.sha256
```

## 5. Publish and verify the release

Publish the draft only after all checks pass.

1. Select **Edit** on the draft release.
2. Select **Publish release**.
3. Confirm that the release appears on the [releases page](https://github.com/Azure/kubelogin/releases).
4. Confirm that the new tag exists.
5. Monitor the [Docker Build and Publish workflow](https://github.com/Azure/kubelogin/actions/workflows/docker-publish.yml).

Publishing the GitHub release starts the Docker workflow. The workflow publishes versioned and `latest` images to GitHub Container Registry.

Run the [WinGet publishing workflow](https://github.com/Azure/kubelogin/actions/workflows/publish-winget.yaml) separately when its installer package is available.

## Failed workflow runs

Use **Re-run failed jobs** when the Release workflow fails. Existing assets are replaced during a retry.

Do not delete a published release or tag to retry a failed downstream workflow. Re-run the failed downstream workflow instead.
