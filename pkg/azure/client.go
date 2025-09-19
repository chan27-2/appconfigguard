package azure

import (
	"context"
	"errors"
	"fmt"
	"os"

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
				Value: *setting.Value,
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
			Value: *setting.Value,
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
			err := c.setSetting(ctx, change.Key, change.Value, change.Label, change.Tags)
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
func (c *Client) setSetting(ctx context.Context, key, value, label string, tags map[string]string) error {
	options := &azappconfig.SetSettingOptions{}

	if label != "" {
		options.Label = &label
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
