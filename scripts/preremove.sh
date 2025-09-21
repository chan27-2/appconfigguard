#!/bin/bash

# Pre-removal script for AppConfigGuard
# This script runs before the package is removed

echo "Removing AppConfigGuard..."

# Note: This script runs before removal, so the binary might still be available
# but we don't want to execute it during uninstallation
