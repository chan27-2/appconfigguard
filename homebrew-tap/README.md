# Homebrew Tap for AppConfigGuard

This is the Homebrew tap for [AppConfigGuard](https://github.com/chan27-2/appconfigguard), a CLI tool for safely managing Azure App Configuration.

## Setup Instructions

1. Create a new GitHub repository for the tap: `homebrew-appconfigguard`
2. Copy the `appconfigguard.rb` formula to the repository
3. Update the version and SHA256 hash when creating releases

## Usage

Once the tap is published, users can install AppConfigGuard with:

```bash
brew tap chan27-2/appconfigguard
brew install appconfigguard
```

## Automation

The formula can be automatically updated using GoReleaser by uncommenting the `brews` section in `.goreleaser.yml` in the main repository.

To enable automated publishing to Homebrew:

1. Update the `.goreleaser.yml` file in the main repo
2. Create the Homebrew tap repository
3. Set up the necessary permissions for GoReleaser to push to the tap repo
