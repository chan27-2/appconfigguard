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
