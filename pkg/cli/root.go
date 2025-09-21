package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chan27-2/appconfigguard/pkg/azure"
	"github.com/chan27-2/appconfigguard/pkg/diff"
	jsonpkg "github.com/chan27-2/appconfigguard/pkg/json"
	"github.com/chan27-2/appconfigguard/pkg/sync"
	"github.com/spf13/cobra"
)

// ANSI color codes for terminal output
const (
	colorReset = "\033[0m"
	colorYellow = "\033[33m"
)

// colorize adds ANSI color to text
func colorize(text, color string) string {
	return color + text + colorReset
}

var (
	// Global flags
	filePath    string
	endpoint    string
	apply       bool
	strict      bool
	ci          bool
	output      string
	label       string
	tags        string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "appconfigguard",
	Short: "Safely manage Azure App Configuration with local JSON files",
	Long: `AppConfigGuard is a CLI tool that helps developers and DevOps teams safely preview,
verify, and synchronize cloud configuration with local JSON files—reducing risk and making
config updates predictable and automation-friendly.

AUTHENTICATION:
  AppConfigGuard supports multiple authentication methods (tried in order):

  1. Access Key (Recommended for production):
     Set APP_CONFIG_CONNECTION_STRING environment variable:
     export APP_CONFIG_CONNECTION_STRING="Endpoint=https://your-store.azconfig.io;Id=your-id;Secret=your-secret"

  2. Azure CLI (for development):
     Run 'az login' to authenticate with your Azure account

  3. Managed Identity (when running on Azure resources)

  4. Environment Variables (client ID/secret/tenant)

GETTING ACCESS KEYS:
  For production use, create access keys in the Azure portal or via Azure CLI:
  az appconfig credential list --name <store-name> --resource-group <rg> --query "[].{name:name, value:value}"

EXAMPLES:
  # Quick start - preview changes
  appconfigguard --file=config.json --endpoint=https://mystorage.azconfig.io

  # Apply changes with confirmation
  appconfigguard --file=config.json --endpoint=https://mystorage.azconfig.io --apply

  # Strict sync (removes keys not in local file)
  appconfigguard --file=config.json --endpoint=https://mystorage.azconfig.io --strict --apply

  # CI/CD mode with JSON output
  appconfigguard --file=config.json --endpoint=https://mystorage.azconfig.io --ci --output=json

  # Use specific label
  appconfigguard --file=config.json --endpoint=https://mystorage.azconfig.io --label=production

  # Download configuration from Azure
  appconfigguard download --endpoint=https://mystorage.azconfig.io --output=config.json`,
	RunE: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to local JSON configuration file (required)")
	rootCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Azure App Configuration endpoint URL (required)")
	rootCmd.Flags().BoolVar(&apply, "apply", false, "Apply the changes after preview (default: dry-run only)")
	rootCmd.Flags().BoolVar(&strict, "strict", false, "Remove keys from Azure App Config that are not in the local file")
	rootCmd.Flags().BoolVar(&ci, "ci", false, "Non-interactive CI/CD mode with machine-readable output")
	rootCmd.Flags().StringVarP(&output, "output", "o", "console", "Output format: console, json")
	rootCmd.Flags().StringVarP(&label, "label", "l", "", "App Configuration label filter (optional)")
	rootCmd.Flags().StringVar(&tags, "tags", "", "App Configuration tags filter as key=value pairs (optional)")

	rootCmd.MarkFlagRequired("file")
	rootCmd.MarkFlagRequired("endpoint")

	// Add subcommands
	rootCmd.AddCommand(downloadCmd)
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", filePath)
	}

	// Initialize components
	jsonFlattener := jsonpkg.NewFlattener()
	diffEngine := diff.NewEngine()

	// Parse local JSON file
	localConfig, err := parseLocalConfig(filePath, jsonFlattener)
	if err != nil {
		return fmt.Errorf("failed to parse local config: %w", err)
	}

	// Validate configuration
	validationErrors, err := jsonFlattener.ValidateConfiguration(localConfig)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	// Display validation errors if any
	if len(validationErrors) > 0 {
		fmt.Println("⚠️  Configuration validation warnings:")
		for _, validationErr := range validationErrors {
			fmt.Printf("   %s: %s\n", colorize(validationErr.Key, colorYellow), validationErr.Message)
		}
		fmt.Println()
	}

	// Create Azure client and fetch remote config
	azureClient, err := azure.NewClient(endpoint)
	if err != nil {
		return fmt.Errorf("failed to create Azure client: %w", err)
	}

	remoteConfig, err := azureClient.FetchAll(ctx, label)
	if err != nil {
		return fmt.Errorf("failed to fetch remote config: %w", err)
	}

	// Convert remote config to map for comparison
	remoteMap := make(map[string]string)
	for _, item := range remoteConfig {
		remoteMap[item.Key] = item.Value
	}

	// Generate diff
	changes, err := diffEngine.Compare(localConfig, remoteConfig, strict)
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	// Handle output based on mode
	if output == "json" {
		return outputJSON(changes, diffEngine)
	}

	// Display changes
	fmt.Println(diffEngine.FormatConsole(changes))

	// Exit codes for CI mode
	if ci {
		if diffEngine.HasChanges(changes) {
			os.Exit(1) // Changes detected
		}
		os.Exit(0) // No changes
	}

	// Apply changes if requested
	if apply {
		if !diffEngine.HasChanges(changes) {
			fmt.Println("No changes to apply.")
			return nil
		}

		// Confirm with user (unless in CI mode)
		if !ci {
			fmt.Print("\nDo you want to apply these changes? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		syncEngine := sync.NewEngine(azureClient)

		// Validate changes
		if err := syncEngine.ValidateChanges(changes); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Apply changes
		fmt.Println("Applying changes...")
		if err := syncEngine.ApplyChanges(ctx, changes, strict); err != nil {
			return fmt.Errorf("failed to apply changes: %w", err)
		}

		fmt.Println("Changes applied successfully!")
	} else {
		if diffEngine.HasChanges(changes) {
			fmt.Println("\nUse --apply to apply these changes.")
		}
	}

	return nil
}

// parseLocalConfig reads and flattens the local JSON configuration file
func parseLocalConfig(filePath string, flattener *jsonpkg.Flattener) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Flatten the JSON structure
	return flattener.Flatten(data)
}

// outputJSON outputs changes in JSON format
func outputJSON(changes []diff.Change, diffEngine *diff.Engine) error {
	jsonData, err := diffEngine.FormatJSON(changes)
	if err != nil {
		return fmt.Errorf("failed to format JSON output: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
