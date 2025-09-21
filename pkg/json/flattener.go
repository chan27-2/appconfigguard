package json

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/chan27-2/appconfigguard/pkg/validator"
)

// Flattener handles JSON flattening and unflattening operations
type Flattener struct{
	validator *validator.Validator
}

// NewFlattener creates a new JSON flattener instance
func NewFlattener() *Flattener {
	return &Flattener{
		validator: validator.NewValidator(),
	}
}

// Flatten converts nested JSON into flat key/value pairs using dot notation
func (f *Flattener) Flatten(data interface{}) (map[string]string, error) {
	result := make(map[string]string)
	err := f.flattenRecursive(data, "", result)
	return result, err
}

// FlattenAndValidate converts nested JSON into flat key/value pairs with validation
func (f *Flattener) FlattenAndValidate(data interface{}) (map[string]string, []validator.ValidationError, error) {
	result := make(map[string]string)
	err := f.flattenRecursive(data, "", result)
	if err != nil {
		return nil, nil, err
	}

	// Validate the flattened configuration
	errors, validateErr := f.validator.ValidateConfiguration(result)
	return result, errors, validateErr
}

// ValidateConfiguration validates a flattened configuration
func (f *Flattener) ValidateConfiguration(config map[string]string) ([]validator.ValidationError, error) {
	return f.validator.ValidateConfiguration(config)
}

// flattenRecursive recursively flattens nested structures
func (f *Flattener) flattenRecursive(data interface{}, prefix string, result map[string]string) error {
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Map:
		return f.flattenMap(v, prefix, result)
	case reflect.Slice, reflect.Array:
		return f.flattenSlice(v, prefix, result)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		result[prefix] = f.formatValue(v)
		return nil
	default:
		// For complex types, try to marshal to JSON string
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal complex type: %w", err)
		}
		result[prefix] = string(jsonBytes)
		return nil
	}
}

// flattenMap flattens a map structure
func (f *Flattener) flattenMap(v reflect.Value, prefix string, result map[string]string) error {
	for _, key := range v.MapKeys() {
		keyStr := f.formatKey(key)
		newPrefix := f.joinKeys(prefix, keyStr)

		value := v.MapIndex(key)
		if !value.IsValid() {
			continue
		}

		err := f.flattenRecursive(value.Interface(), newPrefix, result)
		if err != nil {
			return err
		}
	}
	return nil
}

// flattenSlice flattens an array/slice structure
func (f *Flattener) flattenSlice(v reflect.Value, prefix string, result map[string]string) error {
	for i := 0; i < v.Len(); i++ {
		newPrefix := f.joinKeys(prefix, strconv.Itoa(i))
		value := v.Index(i)

		err := f.flattenRecursive(value.Interface(), newPrefix, result)
		if err != nil {
			return err
		}
	}
	return nil
}

// formatKey formats a key for flattening
func (f *Flattener) formatKey(key reflect.Value) string {
	switch key.Kind() {
	case reflect.String:
		return key.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(key.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(key.Uint(), 10)
	default:
		return fmt.Sprintf("%v", key.Interface())
	}
}

// formatValue formats a value for storage
func (f *Flattener) formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.String:
		return v.String()
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// joinKeys joins key parts with dot notation
func (f *Flattener) joinKeys(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// Unflatten converts flat key/value pairs back into nested JSON
func (f *Flattener) Unflatten(flat map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range flat {
		parts := strings.Split(key, ".")
		err := f.unflattenRecursive(result, parts, value)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// unflattenRecursive recursively builds nested structure from flat keys
func (f *Flattener) unflattenRecursive(current map[string]interface{}, parts []string, value string) error {
	if len(parts) == 0 {
		return nil
	}

	key := parts[0]
	remaining := parts[1:]

	if len(remaining) == 0 {
		// This is a leaf node
		parsedValue, err := f.parseValue(value)
		if err != nil {
			return err
		}
		current[key] = parsedValue
		return nil
	}

	// Check if this should be an array or object
	if f.isArrayIndex(remaining[0]) {
		// This should be an array
		if _, exists := current[key]; !exists {
			current[key] = make([]interface{}, 0)
		}

		arr, ok := current[key].([]interface{})
		if !ok {
			return fmt.Errorf("type conflict at key %s", key)
		}

		// Extend array if necessary
		index, _ := strconv.Atoi(remaining[0])
		for len(arr) <= index {
			arr = append(arr, nil)
		}

		if arr[index] == nil {
			arr[index] = make(map[string]interface{})
		}

		child, ok := arr[index].(map[string]interface{})
		if !ok {
			child = make(map[string]interface{})
			arr[index] = child
		}

		current[key] = arr
		return f.unflattenRecursive(child, remaining[1:], value)
	} else {
		// This should be an object
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}

		child, ok := current[key].(map[string]interface{})
		if !ok {
			return fmt.Errorf("type conflict at key %s", key)
		}

		return f.unflattenRecursive(child, remaining, value)
	}
}

// isArrayIndex checks if a string represents an array index
func (f *Flattener) isArrayIndex(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// parseValue attempts to parse a string value into its appropriate type
func (f *Flattener) parseValue(value string) (interface{}, error) {
	// For Azure App Configuration, we want to keep values as strings
	// since that's how they're stored. Only parse complex JSON structures.
	// Check if it's a JSON object/array first
	if (strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) ||
		(strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")) {
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			return jsonValue, nil
		}
	}

	// Default to string for all other values
	return value, nil
}
