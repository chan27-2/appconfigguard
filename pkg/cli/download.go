package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chan27-2/appconfigguard/pkg/azure"
	jsonpkg "github.com/chan27-2/appconfigguard/pkg/json"
	"github.com/spf13/cobra"
)

var (
	downloadOutputFile string
	downloadLabel      string
	downloadTags       string
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download configuration from Azure App Configuration and save as JSON",
	Long: `Download configuration from Azure App Configuration and convert it to a local JSON file
in the same format that the tool expects for uploads.

This command fetches all configuration settings from the specified Azure App Configuration
store and converts them back into structured JSON format. The resulting file can then be
used with the main sync command.

EXAMPLES:
  # Download all configuration to config.json
  appconfigguard download --endpoint=https://mystorage.azconfig.io --output=config.json

  # Download configuration with specific label
  appconfigguard download --endpoint=https://mystorage.azconfig.io --output=config.json --label=production

  # Download configuration with specific tags
  appconfigguard download --endpoint=https://mystorage.azconfig.io --output=config.json --tags="env=prod,team=backend"`,
	RunE: runDownload,
}

func init() {
	downloadCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Azure App Configuration endpoint URL (required)")
	downloadCmd.Flags().StringVarP(&downloadOutputFile, "output", "o", "", "Output file path for the downloaded configuration (required)")
	downloadCmd.Flags().StringVarP(&downloadLabel, "label", "l", "", "App Configuration label filter (optional)")
	downloadCmd.Flags().StringVar(&downloadTags, "tags", "", "App Configuration tags filter as key=value pairs (optional)")

	downloadCmd.MarkFlagRequired("endpoint")
	downloadCmd.MarkFlagRequired("output")
}

func runDownload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("📥 Downloading configuration from Azure App Configuration...")

	// Create Azure client
	azureClient, err := azure.NewClient(endpoint)
	if err != nil {
		return fmt.Errorf("failed to create Azure client: %w", err)
	}

	// Fetch configuration from Azure
	fmt.Printf("Fetching configuration from: %s\n", endpoint)
	if downloadLabel != "" {
		fmt.Printf("Using label filter: %s\n", downloadLabel)
	}

	configItems, err := azureClient.FetchAll(ctx, downloadLabel)
	if err != nil {
		return fmt.Errorf("failed to fetch configuration: %w", err)
	}

	if len(configItems) == 0 {
		fmt.Println("⚠️  No configuration items found in Azure App Configuration")
		return fmt.Errorf("no configuration items found")
	}

	fmt.Printf("✅ Found %d configuration items\n", len(configItems))

	// Convert ConfigItems to flat map
	flatConfig := make(map[string]string)
	for _, item := range configItems {
		flatConfig[item.Key] = item.Value
	}

	// Initialize JSON flattener
	jsonFlattener := jsonpkg.NewFlattener()

	// Validate the configuration
	validationErrors, err := jsonFlattener.ValidateConfiguration(flatConfig)
	if err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}

	if len(validationErrors) > 0 {
		fmt.Println("⚠️  Configuration validation warnings:")
		for _, validationErr := range validationErrors {
			fmt.Printf("   %s: %s\n", colorize(validationErr.Key, colorYellow), validationErr.Message)
		}
		fmt.Println()
	}

	// Unflatten the configuration back to structured JSON
	fmt.Println("🔄 Converting to structured JSON...")
	structuredConfig, err := jsonFlattener.Unflatten(flatConfig)
	if err != nil {
		return fmt.Errorf("failed to unflatten configuration: %w", err)
	}

	// Write to file
	fmt.Printf("💾 Saving to file: %s\n", downloadOutputFile)
	file, err := os.Create(downloadOutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(structuredConfig); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	fmt.Printf("✅ Successfully downloaded and saved configuration to %s\n", downloadOutputFile)
	return nil
}
