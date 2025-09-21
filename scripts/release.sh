#!/bin/bash

# AppConfigGuard Release Script
# This script helps create releases for the AppConfigGuard project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Please provide a version number (e.g., ./release.sh v1.0.0)"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Version must be in format v1.2.3 (got: $VERSION)"
    exit 1
fi

print_status "Creating release $VERSION"

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
    print_error "Working directory is not clean. Please commit or stash changes first."
    exit 1
fi

# Run tests
print_status "Running tests..."
make test

# Check GoReleaser config
print_status "Checking GoReleaser configuration..."
make release-check

# Create and push git tag
print_status "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

print_status "Release $VERSION created successfully!"
print_status "GitHub Actions will automatically build and publish the release."
print_status "Check the Actions tab in GitHub to monitor the release process."

# Instructions for next steps
echo ""
print_status "Next steps:"
echo "1. Monitor the GitHub Actions workflow"
echo "2. Once complete, the release will be available at:"
echo "   https://github.com/chan27-2/appconfigguard/releases/tag/$VERSION"
echo "3. Users can now install via:"
echo "   - Download binaries from GitHub releases"
echo "   - go install github.com/chan27-2/appconfigguard@$VERSION"
echo "   - Homebrew (once tap is set up)"
