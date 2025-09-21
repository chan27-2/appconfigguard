#!/bin/bash

# Post-installation script for AppConfigGuard
# This script runs after the package is installed

echo "AppConfigGuard has been installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Run 'appconfigguard --help' to see available commands"
echo "  2. Create a config file from the example:"
echo "     cp /usr/share/doc/appconfigguard/config.example.json ~/myconfig.json"
echo "  3. Set up Azure authentication (choose one method):"
echo "     - Azure CLI: az login"
echo "     - Access Key: Set APP_CONFIG_CONNECTION_STRING environment variable"
echo ""
echo "Documentation: https://github.com/chan27-2/appconfigguard"
echo "Example config: /usr/share/doc/appconfigguard/config.example.json"
