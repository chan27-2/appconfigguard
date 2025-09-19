package validator

import (
	"testing"
)

func TestValidateAndParseValue(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		key      string
		value    string
		expected ValueType
		hasError bool
	}{
		{
			name:     "regular value",
			key:      "database.host",
			value:    "localhost",
			expected: ValueTypeRegular,
			hasError: false,
		},
		{
			name:     "feature flag - enabled",
			key:      "feature.new_ui",
			value:    "true",
			expected: ValueTypeFeatureFlag,
			hasError: false,
		},
		{
			name:     "feature flag - disabled",
			key:      "enable.beta_features",
			value:    "false",
			expected: ValueTypeFeatureFlag,
			hasError: false,
		},
		{
			name:     "Key Vault reference - Microsoft format",
			key:      "secrets.api_key",
			value:    "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/mysecret/version)",
			expected: ValueTypeKeyVault,
			hasError: false,
		},
		{
			name:     "Key Vault reference - direct URI",
			key:      "secrets.database_password",
			value:    "https://myvault.vault.azure.net/secrets/db-password",
			expected: ValueTypeKeyVault,
			hasError: false,
		},
		{
			name:     "regular https URL (not Key Vault)",
			key:      "api.endpoint",
			value:    "https://api.example.com/v1",
			expected: ValueTypeRegular,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.ValidateAndParseValue(tt.key, tt.value)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError && result.Type != tt.expected {
				t.Errorf("expected type %s, got %s", tt.expected, result.Type)
			}
		})
	}
}

func TestParseKeyVaultReference(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name           string
		value          string
		expectedVault  string
		expectedSecret string
		expectedVersion string
		hasError       bool
	}{
		{
			name:            "Microsoft format with version",
			value:           "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/mysecret/v1)",
			expectedVault:   "https://myvault.vault.azure.net",
			expectedSecret:  "mysecret",
			expectedVersion: "v1",
			hasError:        false,
		},
		{
			name:            "Microsoft format without version",
			value:           "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/mysecret)",
			expectedVault:   "https://myvault.vault.azure.net",
			expectedSecret:  "mysecret",
			expectedVersion: "",
			hasError:        false,
		},
		{
			name:            "Direct URI format",
			value:           "https://myvault.vault.azure.net/secrets/db-password/v2",
			expectedVault:   "https://myvault.vault.azure.net",
			expectedSecret:  "db-password",
			expectedVersion: "v2",
			hasError:        false,
		},
		{
			name:     "Invalid format",
			value:    "invalid-format",
			hasError: true,
		},
		{
			name:     "Non-Key Vault URL",
			value:    "https://example.com/secrets/mysecret",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.parseKeyVaultReference(tt.value)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.VaultURL != tt.expectedVault {
				t.Errorf("expected vault URL %s, got %s", tt.expectedVault, result.VaultURL)
			}

			if result.SecretName != tt.expectedSecret {
				t.Errorf("expected secret name %s, got %s", tt.expectedSecret, result.SecretName)
			}

			if result.SecretVersion != tt.expectedVersion {
				t.Errorf("expected secret version %s, got %s", tt.expectedVersion, result.SecretVersion)
			}
		})
	}
}

func TestIsFeatureFlagKey(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		key      string
		expected bool
	}{
		{"feature.new_ui", true},
		{"flag.beta_feature", true},
		{"enable.dark_mode", true},
		{"disable.analytics", true},
		{"database.enabled", true},
		{"api.feature", true},
		{"database.host", false},
		{"api.endpoint", false},
		{"regular_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := v.isFeatureFlagKey(tt.key)
			if result != tt.expected {
				t.Errorf("expected %v for key %s, got %v", tt.expected, tt.key, result)
			}
		})
	}
}

func TestValidateConfiguration(t *testing.T) {
	v := NewValidator()

	config := map[string]string{
		"database.host":                   "localhost",
		"feature.new_ui":                  "true",
		"enable.beta_features":            "false",
		"secrets.api_key":                 "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/api-key)",
		"api.endpoint":                    "https://api.example.com/v1", // This is a regular URL, not KV
		"regular.setting":                 "some_value",
	}

	errors, err := v.ValidateConfiguration(config)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have no errors since all configurations are valid
	if len(errors) != 0 {
		t.Errorf("expected 0 validation errors, got %d: %v", len(errors), errors)
	}
}
