# AppConfigGuard Distribution Guide

This document outlines the distribution methods available for AppConfigGuard and how to use them.

## 🚀 Distribution Methods

### 1. GitHub Releases (Automated) ✅

**Status**: Fully configured and ready to use.

**What it provides**:

- Cross-platform binaries (Linux, macOS, Windows)
- Both x86_64 and ARM64 architectures
- Automatic changelog generation
- Checksums for verification

**How it works**:

- Push a git tag (e.g., `v1.0.0`) to trigger automated release
- GitHub Actions runs tests and builds binaries
- GoReleaser creates the GitHub release with all assets

**Usage for users**:

```bash
# Download from: https://github.com/chan27-2/appconfigguard/releases
# Extract and run the binary for your platform
```

### 2. Go Install (For Go Developers) ✅

**Status**: Ready to use (works with any public GitHub repo).

**Usage for users**:

```bash
go install github.com/chan27-2/appconfigguard@latest
# Or for a specific version:
go install github.com/chan27-2/appconfigguard@v1.0.0
```

### 3. Homebrew (macOS) 🔄

**Status**: Formula prepared, tap repository needs to be created.

**Setup required**:

1. Create a new GitHub repository: `chan27-2/homebrew-appconfigguard`
2. Copy `homebrew-tap/appconfigguard.rb` to the repository
3. Uncomment the `brews` section in `.goreleaser.yml`
4. Push a release to enable automated Homebrew publishing

**Usage for users** (once tap is set up):

```bash
brew tap chan27-2/appconfigguard
brew install appconfigguard
```

### 4. Linux Package Managers ✅

**Status**: Fully configured and ready.

**Supported formats**:

- **DEB**: Ubuntu, Debian, and derivatives
- **RPM**: Fedora, CentOS, RHEL, SUSE, and derivatives
- **APK**: Alpine Linux

**Features**:

- Proper package metadata and dependencies
- Includes documentation and example config
- Post-install script with setup instructions
- Automatic package building via GoReleaser

**Usage for users**:

```bash
# Ubuntu/Debian
sudo dpkg -i appconfigguard_*.deb
# or
sudo apt install ./appconfigguard_*.deb

# Fedora/CentOS/RHEL
sudo rpm -i appconfigguard_*.rpm
# or
sudo dnf install ./appconfigguard_*.rpm

# Alpine
sudo apk add --allow-untrusted appconfigguard_*.apk
```

### 5. Docker Containers ✅

**Status**: Fully configured and ready.

**Features**:

- Multi-architecture support (AMD64 and ARM64)
- Minimal Alpine Linux base image
- OCI-compliant labels
- Includes example config file
- Automatic manifest creation for multi-arch images

**Usage for users**:

```bash
# Pull the latest version
docker pull chan27-2/appconfigguard:latest

# Run with help
docker run --rm chan27-2/appconfigguard --help

# Run with Azure CLI authentication (mount azcli config)
docker run --rm -v ~/.azure:/root/.azure chan27-2/appconfigguard --file config.json --endpoint https://example.azconfig.io

# Run with access key authentication
docker run --rm -e APP_CONFIG_CONNECTION_STRING="your-connection-string" chan27-2/appconfigguard --file config.json --endpoint https://example.azconfig.io
```

### 6. Windows Package Managers ✅

**Status**: Fully configured and ready.

**Winget**: Automated manifest publishing to Microsoft Winget Community Repository.

**Scoop**: Automated manifest publishing to Scoop bucket (requires bucket repository setup).

**Setup required**:

1. **Winget**: Manifests are automatically submitted via GoReleaser
2. **Scoop**: Create repository `chan27-2/appconfigguard` for the bucket

## 📋 Release Process

### Creating a Release

1. **Prepare your code**: Make sure all changes are committed and tests pass
2. **Update version**: The version comes from the git tag
3. **Create release**: Use the provided script or manually create a tag

#### Option A: Using the release script (Recommended)

```bash
make release-tag VERSION=v1.0.0
```

#### Option B: Manual process

```bash
# Run tests
make test

# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

#### Option C: Remove a release tag

If you need to remove a tag (e.g., if there was an error or you want to recreate it):

```bash
# Remove a tag locally and remotely
make release-remove-tag VERSION=v1.0.0
# Or using the script directly
./scripts/release.sh --remove v1.0.0
```

### Automated Release Flow

1. Git tag is pushed → GitHub Actions triggers
2. Tests run on multiple platforms (Linux, macOS, Windows)
3. GoReleaser builds binaries for all platforms
4. GitHub release is created with all assets
5. (Future) Homebrew tap is automatically updated

## 🛠️ Configuration Files

- **`.goreleaser.yml`**: GoReleaser configuration for building and releasing (includes all distribution methods)
- **`.github/workflows/release.yml`**: GitHub Actions workflow for automated releases
- **`Makefile`**: Contains release targets (`release-check`, `release-snapshot`, `release-tag`)
- **`scripts/release.sh`**: Automated release script with validation
- **`scripts/postinstall.sh`**: Post-installation script for Linux packages
- **`scripts/preremove.sh`**: Pre-removal script for Linux packages
- **`Dockerfile`**: Multi-stage Docker build configuration
- **`homebrew-tap/`**: Homebrew formula and setup instructions

## 🔧 Development Commands

```bash
# Test release configuration
make release-check

# Create a local snapshot release (for testing)
make release-snapshot

# Create a new release tag
make release-tag VERSION=v1.0.0

# Remove a release tag
make release-remove-tag VERSION=v1.0.0

# Manual release (requires GITHUB_TOKEN)
export GITHUB_TOKEN=your_token_here
make release
```

## 📊 Release Assets

Each release automatically includes:

- **Binaries**: `appconfigguard_<OS>_<ARCH>.tar.gz` or `.zip`
- **Linux Packages**: `.deb`, `.rpm`, and `.apk` packages
- **Docker Images**: Multi-arch images pushed to Docker Hub
- **Package Manager Manifests**: Winget and Scoop manifests
- **Checksums**: `checksums.txt` with SHA256 hashes
- **Changelog**: Automatically generated from git commits

## 🎯 Next Steps

1. **Create your first release**: Try `make release-tag VERSION=v0.1.0`
2. **Set up Homebrew tap**: Create the repository and enable automated publishing
3. **Set up Scoop bucket**: Create repository `chan27-2/appconfigguard` for Scoop manifests
4. **Test Docker builds**: Try `make release-snapshot` to test Docker image creation
5. **Test Linux packages**: Verify package installation works on target distributions

## 🔍 Troubleshooting

- **Release fails**: Check GitHub Actions logs
- **GoReleaser issues**: Run `make release-check` to validate config
- **Missing assets**: Ensure all builds complete successfully in CI

## 📝 Notes

- All distribution methods are configured for zero-dependency static binaries
- Cross-compilation is handled automatically by GoReleaser
- The project follows semantic versioning (vMAJOR.MINOR.PATCH)
- CI/CD ensures releases are only created from passing builds
