package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
)

// ConfigItem represents a single configuration item
type ConfigItem struct {
	Key   string
	Value string
	Label string
	Tags  map[string]string
}

// Client wraps the Azure App Configuration client
type Client struct {
	client *azappconfig.Client
}

// NewClient creates a new Azure App Configuration client
// It first tries access key authentication via APP_CONFIG_CONNECTION_STRING environment variable,
// then falls back to Azure Identity (managed identity, CLI login, etc.)
func NewClient(endpoint string) (*Client, error) {
	var client *azappconfig.Client
	var err error

	// Try connection string authentication first (for access keys)
	if connStr := os.Getenv("APP_CONFIG_CONNECTION_STRING"); connStr != "" {
		client, err = azappconfig.NewClientFromConnectionString(connStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create App Config client from connection string: %w", err)
		}
	} else {
		// Fall back to Azure Identity
		cred, credErr := azidentity.NewDefaultAzureCredential(nil)
		if credErr != nil {
			return nil, fmt.Errorf("failed to create Azure credential: %w", credErr)
		}

		client, err = azappconfig.NewClient(endpoint, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create App Config client: %w", err)
		}
	}

	return &Client{
		client: client,
	}, nil
}

// FetchAll retrieves all configuration items from Azure App Config
func (c *Client) FetchAll(ctx context.Context, labelFilter string) ([]ConfigItem, error) {
	var items []ConfigItem

	selector := azappconfig.SettingSelector{}
	if labelFilter != "" {
		selector.LabelFilter = &labelFilter
	}

	pager := c.client.NewListSettingsPager(selector, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list settings: %w", err)
		}

		for _, setting := range page.Settings {
			if setting.Key == nil || setting.Value == nil {
				continue
			}

			item := ConfigItem{
				Key:   *setting.Key,
				Value: c.normalizeRetrievedValue(*setting.Value),
			}

			if setting.Label != nil {
				item.Label = *setting.Label
			}

			if setting.Tags != nil {
				item.Tags = setting.Tags
			}

			items = append(items, item)
		}
	}

	return items, nil
}

// FetchByKeys retrieves specific configuration items by keys
func (c *Client) FetchByKeys(ctx context.Context, keys []string, labelFilter string) ([]ConfigItem, error) {
	var items []ConfigItem

	for _, key := range keys {
		setting, err := c.client.GetSetting(ctx, key, &azappconfig.GetSettingOptions{
			Label: &labelFilter,
		})
		if err != nil {
			// If key doesn't exist, continue (it might be a new key)
			var respErr *azcore.ResponseError
			if !errors.As(err, &respErr) || respErr.StatusCode != 404 {
				return nil, fmt.Errorf("failed to get setting %s: %w", key, err)
			}
			continue
		}

		if setting.Key == nil || setting.Value == nil {
			continue
		}

		item := ConfigItem{
			Key:   *setting.Key,
			Value: c.normalizeRetrievedValue(*setting.Value),
		}

		if setting.Label != nil {
			item.Label = *setting.Label
		}

		if setting.Tags != nil {
			item.Tags = setting.Tags
		}

		items = append(items, item)
	}

	return items, nil
}

// ApplyChanges applies a batch of changes atomically
func (c *Client) ApplyChanges(ctx context.Context, changes []ChangeOperation) error {
	// TODO: Implement batch operations with atomicity
	// For now, apply changes one by one
	for _, change := range changes {
		switch change.Operation {
		case "add", "update":
			contentType := c.detectContentType(change.Value)
			actualValue := c.formatValueForStorage(change.Value, contentType)
			err := c.setSetting(ctx, change.Key, actualValue, change.Label, change.Tags, contentType)
			if err != nil {
				return fmt.Errorf("failed to set setting %s: %w", change.Key, err)
			}
		case "delete":
			err := c.deleteSetting(ctx, change.Key, change.Label)
			if err != nil {
				return fmt.Errorf("failed to delete setting %s: %w", change.Key, err)
			}
		}
	}

	return nil
}

// ChangeOperation represents a single change to apply
type ChangeOperation struct {
	Operation string            // "add", "update", "delete"
	Key       string
	Value     string
	Label     string
	Tags      map[string]string
}

// setSetting creates or updates a setting
func (c *Client) setSetting(ctx context.Context, key, value, label string, tags map[string]string, contentType *string) error {
	options := &azappconfig.SetSettingOptions{}

	if label != "" {
		options.Label = &label
	}

	if contentType != nil {
		options.ContentType = contentType
	}

	// Note: Tags are not supported in SetSettingOptions
	// Tags would need to be set via a separate API call if required

	_, err := c.client.SetSetting(ctx, key, &value, options)
	return err
}

// deleteSetting removes a setting
func (c *Client) deleteSetting(ctx context.Context, key, label string) error {
	_, err := c.client.DeleteSetting(ctx, key, &azappconfig.DeleteSettingOptions{
		Label: &label,
	})

	return err
}

// DetectContentType determines the content type based on the value format (public method)
func (c *Client) DetectContentType(value string) *string {
	return c.detectContentType(value)
}

// detectContentType determines the content type based on the value format
func (c *Client) detectContentType(value string) *string {
	// Check for Key Vault references
	if c.isKeyVaultReference(value) {
		contentType := "application/vnd.microsoft.appconfig.keyvaultref+json;charset=utf-8"
		return &contentType
	}

	// Default to text/plain for regular values
	contentType := "text/plain"
	return &contentType
}

// isKeyVaultReference checks if a value is a Key Vault reference
func (c *Client) isKeyVaultReference(value string) bool {
	// Check for Microsoft.KeyVault format
	if strings.HasPrefix(value, "@Microsoft.KeyVault(") && strings.HasSuffix(value, ")") {
		return true
	}

	// Check for direct Key Vault URI format
	if strings.HasPrefix(value, "https://") && strings.Contains(value, "vault.azure.net") {
		return true
	}

	return false
}

// FormatValueForStorage formats the value for storage based on content type (public method)
func (c *Client) FormatValueForStorage(value string, contentType *string) string {
	return c.formatValueForStorage(value, contentType)
}

// formatValueForStorage formats the value for storage based on content type
func (c *Client) formatValueForStorage(value string, contentType *string) string {
	if contentType != nil && *contentType == "application/vnd.microsoft.appconfig.keyvaultref+json;charset=utf-8" {
		return c.formatKeyVaultReference(value)
	}

	return value
}

// formatKeyVaultReference converts various Key Vault reference formats to the standard JSON format
func (c *Client) formatKeyVaultReference(value string) string {
	var uri string

	// Handle Microsoft.KeyVault format
	if strings.HasPrefix(value, "@Microsoft.KeyVault(") && strings.HasSuffix(value, ")") {
		content := strings.TrimPrefix(strings.TrimSuffix(value, ")"), "@Microsoft.KeyVault(")
		params := make(map[string]string)

		// Parse parameters
		for _, param := range strings.Split(content, ";") {
			if parts := strings.SplitN(param, "=", 2); len(parts) == 2 {
				params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		if secretURI, ok := params["SecretUri"]; ok {
			uri = secretURI
		}
	} else {
		// Direct URI format
		uri = value
	}

	// Format as JSON
	kvRef := map[string]string{
		"uri": uri,
	}

	jsonBytes, err := json.Marshal(kvRef)
	if err != nil {
		// Fallback to original value if JSON marshaling fails
		return value
	}

	return string(jsonBytes)
}

// NormalizeRetrievedValue normalizes values retrieved from Azure App Configuration (public method)
// This handles cases where Key Vault references might be stored in different formats
func (c *Client) NormalizeRetrievedValue(value string) string {
	return c.normalizeRetrievedValue(value)
}

// normalizeRetrievedValue normalizes values retrieved from Azure App Configuration
// This handles cases where Key Vault references might be stored in different formats
func (c *Client) normalizeRetrievedValue(value string) string {
	// Check if it's a JSON Key Vault reference (from proper storage with content-type)
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		var kvRef map[string]interface{}
		if err := json.Unmarshal([]byte(value), &kvRef); err == nil {
			if uri, ok := kvRef["uri"].(string); ok && strings.Contains(uri, "vault.azure.net") {
				// Convert back to Microsoft.KeyVault format for consistent comparison
				return fmt.Sprintf("@Microsoft.KeyVault(SecretUri=%s)", uri)
			}
		}
	}

	// Check if it's already a Microsoft.KeyVault format (could be from old storage)
	if strings.HasPrefix(value, "@Microsoft.KeyVault(") {
		return value
	}

	// Check if it's a direct Key Vault URI (could be from old storage)
	if strings.HasPrefix(value, "https://") && strings.Contains(value, "vault.azure.net") {
		return fmt.Sprintf("@Microsoft.KeyVault(SecretUri=%s)", value)
	}

	// Return as-is for regular values
	return value
}
