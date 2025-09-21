# AppConfigGuard

A Go-based CLI tool for safely and transparently managing Azure App Configuration.

AppConfigGuard helps developers and DevOps teams preview, verify, and synchronize cloud configuration with local JSON files—reducing risk and making config updates predictable and automation-friendly.

## 🚀 Quick Start

Get started in 3 simple steps:

```bash
# 1. Build the tool
make build

# 2. Setup authentication (choose one method)
make setup-auth-cli    # For development (Azure CLI)
# OR
make setup-auth-keys   # For production (Access Keys)

# 3. Run a demo
make demo
```

That's it! The tool will automatically detect your authentication method and show you how to use it.

## 🎯 Project Goals

- Provide a safe, transparent workflow for updating Azure App Configuration
- Support diff-first workflows so users always see exactly what will change before committing
- Enable strict synchronization so remote config matches source-controlled JSON exactly
- Be simple to use interactively, but also automation-friendly in CI/CD pipelines
- Rely on standard Azure credentials (no custom auth schemes)
- Implement atomic updates to avoid partial/inconsistent states

## 🔑 Key Features

### 1. Diff-first Workflow

Fetch current state from Azure App Configuration and compare with the provided local JSON file. Display a colorized, well-formatted diff showing:

- Added keys
- Updated values
- Removed keys

### 2. Safe by Default

Runs in dry-run mode by default (no remote writes). User must explicitly confirm (`--apply`) before any changes are made.

### 3. Flexible Modes

- `--apply`: Apply the changes after preview
- `--strict`: Remove any keys in Azure App Config that are not in the local file
- `--ci`: Non-interactive mode for pipelines, with machine-readable output + exit codes

### 4. JSON Mapping Support

Flatten nested JSON into App Config keys using conventional naming (dot-notation). Support arrays and deeply nested objects. Ensure reversibility when exporting from App Config back into JSON.

### 5. CI/CD Ready

Designed to integrate into GitHub Actions, Azure DevOps, and other CI/CD platforms.

Exit codes reflect outcome:

- `0` → No changes
- `1` → Changes detected but not applied
- `2` → Errors occurred

Optional `--output json` for structured diffs.

### 6. Atomic Updates

Apply all changes in a single logical transaction (commit all-or-nothing). Retry with exponential backoff on transient failures.

## 💻 Example Usage

### Preview changes without applying:

```bash
appconfigguard --file=myconfig.json --endpoint=https://example.azconfig.io
```

### Preview, then apply changes:

```bash
appconfigguard --file=myconfig.json --endpoint=https://example.azconfig.io --apply
```

### Strict sync (removes extra keys from Azure):

```bash
appconfigguard --file=myconfig.json --endpoint=https://example.azconfig.io --strict --apply
```

### Pipeline-friendly JSON diff output:

```bash
appconfigguard --file=myconfig.json --endpoint=https://example.azconfig.io --ci --output=json
```

## 📐 Technical Design

### Language & Runtime

Implemented in Go for portability and performance. Distributed as a single static binary (cross-platform: Linux, Windows, macOS).

### Core Modules

- **Azure API Client**: Wraps Azure App Configuration REST/SDK. Handles authentication via Azure Identity (managed identity, CLI login, env vars).
- **JSON Parser & Flattener**: Converts nested JSON into flat key/value pairs for Azure. Reversible flatten/unflatten functions.
- **Diff Engine**: Compares local vs remote state. Outputs structured diff (add/update/remove). Provides pretty colorized console output + optional JSON.
- **Sync Engine**: Applies diffs atomically. Supports strict mode (removal of extra keys). Implements retry & rollback on failure.
- **CLI Layer**: Built with Cobra. Supports flags, subcommands, and rich help text.

## 🚀 Installation

### Option 1: Download Pre-built Binaries (Recommended)

Download the latest release for your platform from the [GitHub Releases page](https://github.com/chan27-2/appconfigguard/releases).

Supported platforms:

- **Linux**: `appconfigguard_Linux_x86_64.tar.gz`
- **macOS**: `appconfigguard_Darwin_x86_64.tar.gz` or `appconfigguard_Darwin_arm64.tar.gz`
- **Windows**: `appconfigguard_Windows_x86_64.zip`

### Option 2: Install via Go (for Go developers)

If you have Go 1.21+ installed:

```bash
go install github.com/chan27-2/appconfigguard@latest
```

This will install the latest version to your `$GOPATH/bin`.

### Option 3: Build from Source

```bash
git clone https://github.com/chan27-2/appconfigguard.git
cd appconfigguard
make build
# Binary will be available at ./bin/appconfigguard
```

### Option 4: Homebrew (macOS)

```bash
brew tap chan27-2/appconfigguard
brew install appconfigguard
```

_Note: Homebrew formulae are automatically maintained via GoReleaser._

### Option 5: Linux Package Managers

Install via native package managers:

```bash
# Ubuntu/Debian
sudo apt install ./appconfigguard_*.deb

# Fedora/CentOS/RHEL
sudo dnf install ./appconfigguard_*.rpm

# Alpine Linux
sudo apk add --allow-untrusted appconfigguard_*.apk
```

### Option 6: Docker

Run via container:

```bash
# Pull and run
docker run --rm chan27-2/appconfigguard --help

# With Azure CLI auth
docker run --rm -v ~/.azure:/root/.azure chan27-2/appconfigguard --file config.json --endpoint https://example.azconfig.io
```

### Option 7: Windows Package Managers

```bash
# Winget
winget install chan27-2.appconfigguard

# Scoop
scoop bucket add chan27-2 https://github.com/chan27-2/appconfigguard
scoop install appconfigguard
```

_Note: Winget and Scoop manifests are automatically maintained via GoReleaser._

## 🔐 Authentication

AppConfigGuard uses Azure Identity for authentication. The following methods are supported in order of precedence:

1. **Environment variables**: `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
2. **Managed Identity**: When running on Azure resources with managed identity enabled
3. **Azure CLI**: `az login` credentials
4. **Interactive**: Browser-based authentication (fallback)

## 📋 Configuration File Format

AppConfigGuard expects a JSON file that will be flattened using dot-notation. Example:

```json
{
  "app": {
    "name": "MyApp",
    "version": "1.0.0"
  },
  "database": {
    "host": "localhost",
    "port": 5432
  }
}
```

This becomes:

- `app.name` = "MyApp"
- `app.version` = "1.0.0"
- `database.host` = "localhost"
- `database.port` = "5432"

## 🎯 Use Cases

- **Developers** want to preview config changes before updating Azure
- **Teams** want to enforce config-as-code by syncing JSON files in Git to App Config
- **CI/CD pipelines** need safe, automated config deployment with visible change tracking
- **Large enterprises** want consistency and transparency in how configs are managed

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

```bash
git clone https://github.com/chan27-2/appconfigguard.git
cd appconfigguard
go mod download
go build -o bin/appconfigguard ./cmd
```

### Running Tests

```bash
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📂 Repository

- **GitHub**: https://github.com/chan27-2/appconfigguard
- **Issues**: https://github.com/chan27-2/appconfigguard/issues
- **Discussions**: https://github.com/chan27-2/appconfigguard/discussions

## 🙏 Acknowledgments

- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Azure App Configuration](https://docs.microsoft.com/en-us/azure/azure-app-configuration/)
