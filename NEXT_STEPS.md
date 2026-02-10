# Fix for Provider Publishing Error - Next Steps

## Summary of Changes

This PR resolves the provider publishing error by preparing the repository for its first release. The error occurred because the Terraform Registry requires at least one published release with a tag in the format `vVERSION` (e.g., `v0.1.0`).

## What Has Been Done

1. ✅ **Added RELEASE_INSTRUCTIONS.md** - Comprehensive documentation on creating and managing releases
2. ✅ **Added push-release-tag.sh** - Helper script to push tags and trigger the release workflow
3. ✅ **Updated README.md** - Added note about release requirements with link to instructions

## What Needs to Be Done Next

After merging this PR, you need to create and push a release tag to trigger the automated release process.

### Step 1: Create the Tag

On the main branch (after merging this PR), create the v0.1.0 tag:

```bash
git checkout main
git pull origin main
git tag -a v0.1.0 -m "Initial release v0.1.0"
```

### Step 2: Push the Tag

#### Option A: Use the Helper Script (Recommended)

```bash
# For v0.1.0 (default if no version specified)
./push-release-tag.sh

# Or explicitly specify the version
./push-release-tag.sh v0.1.0

# For future releases, pass the version as an argument
./push-release-tag.sh v0.2.0
```

This script will push the tag to GitHub and provide links to monitor the release process.

#### Option B: Manual Tag Push

```bash
git push origin v0.1.0
```

### After Pushing the Tag

Once the tag is pushed:

1. The GitHub Actions workflow will automatically trigger (`.github/workflows/release.yml`)
2. The workflow will:
   - Build binaries for multiple platforms (Linux, macOS, Windows, FreeBSD)
   - Generate SHA256 checksums
   - Sign the release with GPG
   - Create a GitHub Release with all artifacts
3. The Terraform Registry will detect the new release and make the provider available

### Verification Steps

After pushing the tag, verify the release:

1. **Check Workflow Status**: https://github.com/nkbud/terraform-provider-contextforge/actions
2. **View Release**: https://github.com/nkbud/terraform-provider-contextforge/releases
3. **Confirm Tags**: https://github.com/nkbud/terraform-provider-contextforge/tags

## Prerequisites

Before pushing the tag, ensure these secrets are configured in the repository:

- `GPG_PRIVATE_KEY` - For signing releases
- `PASSPHRASE` - For the GPG key

If these secrets are not configured, the release workflow will fail. See `.github/workflows/release.yml` for details.

## Troubleshooting

If the release workflow fails, check:

1. GPG signing keys are properly configured in repository secrets
2. All tests pass: `make testacc`
3. Dependencies are up to date: `go mod tidy`
4. Review workflow logs for specific errors

## Future Releases

For subsequent releases, follow the same process with incremented version numbers:

```bash
# Create the tag
git tag -a v0.2.0 -m "Release v0.2.0"

# Push the tag
git push origin v0.2.0
```

See [RELEASE_INSTRUCTIONS.md](RELEASE_INSTRUCTIONS.md) for comprehensive release management documentation.
