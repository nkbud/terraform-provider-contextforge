# Release Instructions

This document provides instructions for creating and publishing releases for this Terraform provider.

## Publishing Your First Release

The Terraform Registry requires at least one published release with a tag name in the format `vVERSION` where VERSION is a semantic version (e.g., `v0.1.0`, `v1.0.0`).

### Steps to Create and Publish a Release

1. **Create a Git Tag**
   
   Create an annotated tag with a semantic version:
   ```bash
   git tag -a v0.1.0 -m "Initial release v0.1.0"
   ```

2. **Push the Tag to GitHub**
   
   Push the tag to trigger the automated release workflow:
   ```bash
   git push origin v0.1.0
   ```

3. **Verify the Release**
   
   The GitHub Actions workflow (`.github/workflows/release.yml`) will automatically:
   - Build binaries for multiple platforms
   - Generate checksums
   - Sign the release artifacts
   - Create a GitHub release
   
   You can monitor the release process at:
   https://github.com/nkbud/terraform-provider-contextforge/actions

4. **Publish to Terraform Registry**
   
   Once the GitHub release is created, the Terraform Registry will automatically detect it and make it available for users.

## Creating Subsequent Releases

For future releases, follow the same process with an incremented version number:

```bash
# For patch releases (bug fixes)
git tag -a v0.1.1 -m "Release v0.1.1"
git push origin v0.1.1

# For minor releases (new features, backwards compatible)
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# For major releases (breaking changes)
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Prerequisites for Releases

Before creating a release, ensure:

1. **GPG Key Configured**: The repository secrets must have `GPG_PRIVATE_KEY` and `PASSPHRASE` configured for signing releases
2. **All Tests Pass**: Run `make testacc` or your test suite
3. **Documentation Updated**: Ensure docs are current with `make generate`
4. **CHANGELOG Updated**: Document changes in your changelog (if applicable)

## Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.0.0): Incompatible API changes
- **MINOR** version (0.X.0): New functionality, backwards compatible
- **PATCH** version (0.0.X): Backwards compatible bug fixes

## Troubleshooting

### "No releases found that are in the valid format" Error

This error occurs when the Terraform Registry cannot find any releases with tags in the format `vVERSION`. Ensure:

1. At least one tag exists with the format `v*` (e.g., `v0.1.0`)
2. The tag has been pushed to GitHub (check: https://github.com/nkbud/terraform-provider-contextforge/tags)
3. The release workflow has completed successfully
4. A GitHub release exists (check: https://github.com/nkbud/terraform-provider-contextforge/releases)

### Release Workflow Fails

If the GitHub Actions workflow fails:

1. Check that GPG signing keys are properly configured in repository secrets
2. Verify that go.mod and dependencies are up to date
3. Ensure the `.goreleaser.yml` configuration is valid
4. Check the workflow logs for specific error messages
