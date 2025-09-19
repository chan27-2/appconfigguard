package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValueType represents the type of a configuration value
type ValueType string

const (
	ValueTypeRegular    ValueType = "regular"
	ValueTypeFeatureFlag ValueType = "feature_flag"
	ValueTypeKeyVault    ValueType = "keyvault"
)

// SpecialValue represents a parsed special configuration value
type SpecialValue struct {
	Type        ValueType
	Original    string
	ParsedValue interface{}
}

// FeatureFlag represents a feature flag configuration
type FeatureFlag struct {
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
}

// KeyVaultReference represents a Key Vault secret reference
type KeyVaultReference struct {
	VaultURL    string
	SecretName  string
	SecretVersion string
}

// Validator handles validation of configuration values
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateAndParseValue analyzes a value and determines its type with validation
func (v *Validator) ValidateAndParseValue(key, value string) (*SpecialValue, error) {
	// Check for Microsoft.KeyVault format first
	if strings.HasPrefix(value, "@Microsoft.KeyVault(") {
		kvRef, err := v.parseKeyVaultReference(value)
		if err == nil {
			// Additional validation for Key Vault reference
			if validationErr := v.validateKeyVaultReference(kvRef); validationErr != nil {
				return nil, fmt.Errorf("Key Vault validation error: %w", validationErr)
			}
			return &SpecialValue{
				Type:        ValueTypeKeyVault,
				Original:    value,
				ParsedValue: kvRef,
			}, nil
		} else {
			// If it's Microsoft.KeyVault format but invalid, return an error
			return nil, fmt.Errorf("invalid Key Vault reference: %w", err)
		}
	}

	// Check for direct Key Vault URI format (must contain vault.azure.net)
	if strings.HasPrefix(value, "https://") && strings.Contains(value, "vault.azure.net") {
		kvRef, err := v.parseKeyVaultReference(value)
		if err == nil {
			// Additional validation for Key Vault reference
			if validationErr := v.validateKeyVaultReference(kvRef); validationErr != nil {
				return nil, fmt.Errorf("Key Vault validation error: %w", validationErr)
			}
			return &SpecialValue{
				Type:        ValueTypeKeyVault,
				Original:    value,
				ParsedValue: kvRef,
			}, nil
		} else {
			// If it's supposed to be a Key Vault URI but invalid, return an error
			return nil, fmt.Errorf("invalid Key Vault reference: %w", err)
		}
	}

	// Check for feature flag
	if ff, err := v.parseFeatureFlag(key, value); err == nil {
		// Additional validation for feature flag
		if validationErr := v.validateFeatureFlag(ff); validationErr != nil {
			return nil, fmt.Errorf("feature flag validation error: %w", validationErr)
		}
		return &SpecialValue{
			Type:        ValueTypeFeatureFlag,
			Original:    value,
			ParsedValue: ff,
		}, nil
	}

	// Regular value (including other https:// URLs that are not Key Vault references)
	return &SpecialValue{
		Type:        ValueTypeRegular,
		Original:    value,
		ParsedValue: value,
	}, nil
}

// parseKeyVaultReference parses a Key Vault secret reference
// Supports formats:
// - @Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/mysecret/version)
// - https://myvault.vault.azure.net/secrets/mysecret/version
func (v *Validator) parseKeyVaultReference(value string) (*KeyVaultReference, error) {
	// Check for Microsoft.KeyVault format
	if strings.HasPrefix(value, "@Microsoft.KeyVault(") && strings.HasSuffix(value, ")") {
		content := strings.TrimPrefix(strings.TrimSuffix(value, ")"), "@Microsoft.KeyVault(")
		params := make(map[string]string)

		// Parse parameters
		for _, param := range strings.Split(content, ";") {
			if parts := strings.SplitN(param, "=", 2); len(parts) == 2 {
				params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		secretURI, ok := params["SecretUri"]
		if !ok {
			return nil, fmt.Errorf("missing SecretUri in Key Vault reference")
		}

		return v.parseKeyVaultURI(secretURI)
	}

	// Check for direct URI format - must be https and contain vault.azure.net
	if strings.HasPrefix(value, "https://") && strings.Contains(value, "vault.azure.net") {
		return v.parseKeyVaultURI(value)
	}

	// If it's a URL but not a Key Vault URL, return an error
	if strings.HasPrefix(value, "https://") {
		return nil, fmt.Errorf("not a valid Key Vault URI")
	}

	return nil, fmt.Errorf("invalid Key Vault reference format")
}

// parseKeyVaultURI parses a Key Vault URI
func (v *Validator) parseKeyVaultURI(uri string) (*KeyVaultReference, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid Key Vault URI: %w", err)
	}

	// Validate it's a Key Vault URL
	if !strings.Contains(parsedURL.Host, "vault.azure.net") {
		return nil, fmt.Errorf("not a valid Key Vault URI")
	}

	// Parse path: /secrets/{secret-name}/{secret-version}
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] != "secrets" {
		return nil, fmt.Errorf("invalid Key Vault secret path")
	}

	ref := &KeyVaultReference{
		VaultURL:   fmt.Sprintf("https://%s", parsedURL.Host),
		SecretName: pathParts[1],
	}

	// Optional version
	if len(pathParts) > 2 {
		ref.SecretVersion = pathParts[2]
	}

	// Validate secret name
	if !v.isValidSecretName(ref.SecretName) {
		return nil, fmt.Errorf("invalid secret name: %s", ref.SecretName)
	}

	return ref, nil
}

// parseFeatureFlag parses a feature flag value
func (v *Validator) parseFeatureFlag(key, value string) (*FeatureFlag, error) {
	// Feature flags are typically boolean values with special naming
	if v.isFeatureFlagKey(key) {
		// Parse boolean value
		var enabled bool
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on", "enabled":
			enabled = true
		case "false", "0", "no", "off", "disabled":
			enabled = false
		default:
			return nil, fmt.Errorf("invalid feature flag value: %s", value)
		}

		return &FeatureFlag{
			Description: v.extractFeatureDescription(key),
			Enabled:     enabled,
		}, nil
	}

	return nil, fmt.Errorf("not a feature flag key")
}

// isFeatureFlagKey determines if a key represents a feature flag
func (v *Validator) isFeatureFlagKey(key string) bool {
	// Common feature flag patterns
	patterns := []string{
		`^feature\.`,
		`^flag\.`,
		`\.enabled$`,
		`\.disabled$`,
		`^enable\.`,
		`^disable\.`,
		`\.feature$`,
		`\.flag$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, strings.ToLower(key)); matched {
			return true
		}
	}

	return false
}

// extractFeatureDescription extracts a human-readable description from a feature flag key
func (v *Validator) extractFeatureDescription(key string) string {
	// Convert key to readable description
	description := strings.ReplaceAll(key, ".", " ")
	description = strings.ReplaceAll(description, "_", " ")
	description = strings.Title(description)
	return description
}

// isValidSecretName validates a Key Vault secret name
func (v *Validator) isValidSecretName(name string) bool {
	// Key Vault secret names must be 1-127 characters, alphanumeric and some special chars
	if len(name) < 1 || len(name) > 127 {
		return false
	}

	// Only alphanumeric, hyphen, and underscore allowed
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			 (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	return true
}

// ValidateConfiguration validates an entire configuration map
func (v *Validator) ValidateConfiguration(config map[string]string) ([]ValidationError, error) {
	var errors []ValidationError

	for key, value := range config {
		specialValue, err := v.ValidateAndParseValue(key, value)
		if err != nil {
			errors = append(errors, ValidationError{
				Key:     key,
				Value:   value,
				Message: err.Error(),
				Type:    "parsing_error",
			})
			continue
		}

		// Additional validation based on type
		switch specialValue.Type {
		case ValueTypeKeyVault:
			if kvRef, ok := specialValue.ParsedValue.(*KeyVaultReference); ok {
				if err := v.validateKeyVaultReference(kvRef); err != nil {
					errors = append(errors, ValidationError{
						Key:     key,
						Value:   value,
						Message: fmt.Sprintf("Key Vault validation error: %s", err.Error()),
						Type:    "keyvault_error",
					})
				}
			}
		case ValueTypeFeatureFlag:
			if ff, ok := specialValue.ParsedValue.(*FeatureFlag); ok {
				if err := v.validateFeatureFlag(ff); err != nil {
					errors = append(errors, ValidationError{
						Key:     key,
						Value:   value,
						Message: fmt.Sprintf("Feature flag validation error: %s", err.Error()),
						Type:    "feature_flag_error",
					})
				}
			}
		}
	}

	return errors, nil
}

// validateKeyVaultReference performs additional validation on Key Vault references
func (v *Validator) validateKeyVaultReference(ref *KeyVaultReference) error {
	// Validate vault URL format
	if !strings.HasPrefix(ref.VaultURL, "https://") || !strings.Contains(ref.VaultURL, "vault.azure.net") {
		return fmt.Errorf("invalid vault URL format")
	}

	// Validate secret name is not empty
	if ref.SecretName == "" {
		return fmt.Errorf("secret name cannot be empty")
	}

	return nil
}

// validateFeatureFlag performs additional validation on feature flags
func (v *Validator) validateFeatureFlag(ff *FeatureFlag) error {
	// Feature flags should have a description
	if ff.Description == "" {
		return fmt.Errorf("feature flag should have a description")
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Key     string
	Value   string
	Message string
	Type    string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Key, e.Message)
}
