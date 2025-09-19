package json

import (
	"reflect"
	"testing"
)

func TestFlattener_Flatten(t *testing.T) {
	flattener := NewFlattener()

	tests := []struct {
		name     string
		input    interface{}
		expected map[string]string
	}{
		{
			name: "simple object",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nested object",
			input: map[string]interface{}{
				"app": map[string]interface{}{
					"name":    "test",
					"version": "1.0.0",
				},
				"database": map[string]interface{}{
					"host": "localhost",
					"port": 5432,
				},
			},
			expected: map[string]string{
				"app.name":    "test",
				"app.version": "1.0.0",
				"database.host": "localhost",
				"database.port": "5432",
			},
		},
		{
			name: "array",
			input: map[string]interface{}{
				"features": []interface{}{"auth", "logging", "cache"},
			},
			expected: map[string]string{
				"features.0": "auth",
				"features.1": "logging",
				"features.2": "cache",
			},
		},
		{
			name: "mixed types",
			input: map[string]interface{}{
				"enabled": true,
				"count":   42,
				"rate":    3.14,
			},
			expected: map[string]string{
				"enabled": "true",
				"count":   "42",
				"rate":    "3.14",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := flattener.Flatten(tt.input)
			if err != nil {
				t.Errorf("Flatten() error = %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Flatten() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFlattener_Unflatten(t *testing.T) {
	flattener := NewFlattener()

	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]interface{}
	}{
		{
			name: "simple flat",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nested structure",
			input: map[string]string{
				"app.name":    "test",
				"app.version": "1.0.0",
				"database.host": "localhost",
				"database.port": "5432",
			},
			expected: map[string]interface{}{
				"app": map[string]interface{}{
					"name":    "test",
					"version": "1.0.0",
				},
				"database": map[string]interface{}{
					"host": "localhost",
					"port": "5432",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := flattener.Unflatten(tt.input)
			if err != nil {
				t.Errorf("Unflatten() error = %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Unflatten() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFlattener_FlattenAndValidate(t *testing.T) {
	flattener := NewFlattener()

	tests := []struct {
		name           string
		input          interface{}
		expectedErrors int
	}{
		{
			name: "valid configuration with feature flags and Key Vault",
			input: map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
					"port": 5432,
				},
				"feature": map[string]interface{}{
					"new_ui": "true",
				},
				"enable": map[string]interface{}{
					"beta_features": "false",
				},
				"secrets": map[string]interface{}{
					"api_key": "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/api-key)",
					"db_password": "https://myvault.vault.azure.net/secrets/db-password",
				},
			},
			expectedErrors: 0,
		},
		{
			name: "configuration with invalid Key Vault reference",
			input: map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
				},
				"secrets": map[string]interface{}{
					"invalid_key": "@Microsoft.KeyVault(SecretUri=https://example.com/secrets/mysecret)",
				},
			},
			expectedErrors: 1,
		},
		{
			name: "valid configuration with regular values",
			input: map[string]interface{}{
				"app": map[string]interface{}{
					"name":    "test",
					"version": "1.0.0",
				},
				"database": map[string]interface{}{
					"host":     "localhost",
					"port":     5432,
					"ssl_mode": "require",
				},
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errors, err := flattener.FlattenAndValidate(tt.input)
			if err != nil {
				t.Errorf("FlattenAndValidate() error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("FlattenAndValidate() returned nil result")
				return
			}

			if len(errors) != tt.expectedErrors {
				t.Errorf("FlattenAndValidate() expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}

			// Verify that result is properly flattened
			if len(result) == 0 {
				t.Errorf("FlattenAndValidate() returned empty result")
			}
		})
	}
}

func TestFlattener_ValidateConfiguration(t *testing.T) {
	flattener := NewFlattener()

	config := map[string]string{
		"database.host":                   "localhost",
		"feature.new_ui":                  "true",
		"enable.beta_features":            "false",
		"secrets.api_key":                 "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/api-key)",
		"secrets.invalid_kv":              "@Microsoft.KeyVault(SecretUri=https://example.com/secrets/mysecret)",
		"regular.setting":                 "some_value",
	}

	errors, err := flattener.ValidateConfiguration(config)

	if err != nil {
		t.Errorf("ValidateConfiguration() error = %v", err)
		return
	}

	// Should have one error for the invalid Key Vault reference
	if len(errors) != 1 {
		t.Errorf("ValidateConfiguration() expected 1 error, got %d: %v", len(errors), errors)
	}

	if len(errors) > 0 {
		if errors[0].Type != "parsing_error" {
			t.Errorf("ValidateConfiguration() expected parsing_error type, got %s", errors[0].Type)
		}
	}
}
