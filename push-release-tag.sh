#!/bin/bash

# Script to push the v0.1.0 release tag to GitHub
# This will trigger the automated release workflow

set -e

echo "Pushing release tag v0.1.0 to GitHub..."
git push origin v0.1.0

echo "✓ Tag pushed successfully!"
echo ""
echo "The GitHub Actions release workflow should now be running."
echo "Check the status at: https://github.com/nkbud/terraform-provider-contextforge/actions"
echo ""
echo "Once complete, you can verify the release at:"
echo "https://github.com/nkbud/terraform-provider-contextforge/releases"
