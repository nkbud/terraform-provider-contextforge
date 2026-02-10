#!/bin/bash

# Script to push a release tag to GitHub
# This will trigger the automated release workflow
#
# Usage: ./push-release-tag.sh [VERSION]
# Example: ./push-release-tag.sh v0.2.0
#
# If no version is provided, defaults to v0.1.0

set -e

VERSION=${1:-v0.1.0}

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ Error: Invalid version format. Expected format: vX.Y.Z (e.g., v0.1.0)"
    echo "Usage: $0 [VERSION]"
    exit 1
fi

echo "Pushing release tag $VERSION to GitHub..."
git push origin "$VERSION"

echo "✓ Tag $VERSION pushed successfully!"
echo ""
echo "The GitHub Actions release workflow should now be running."
echo "Check the status at: https://github.com/nkbud/terraform-provider-contextforge/actions"
echo ""
echo "Once complete, you can verify the release at:"
echo "https://github.com/nkbud/terraform-provider-contextforge/releases"
