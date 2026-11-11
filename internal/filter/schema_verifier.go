package filter

import (
	"encoding/json"
	"fmt"
	"os"
)

// SchemaVerifier validates that a response string is valid JSON and adheres to a schema.
type SchemaVerifier struct {
	hasSchema  bool
	schemaData map[string]any
}

// NewSchemaVerifier loads an optional JSON Schema from disk.
func NewSchemaVerifier(schemaPath string) (*SchemaVerifier, error) {
	if schemaPath == "" {
		return &SchemaVerifier{hasSchema: false}, nil
	}

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %q: %w", schemaPath, err)
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(data, &schemaMap); err != nil {
		return nil, fmt.Errorf("invalid json schema syntax: %w", err)
	}

	return &SchemaVerifier{
		hasSchema:  true,
		schemaData: schemaMap,
	}, nil
}

// Validate checks whether the response is valid JSON and verifies required fields if defined in schema.
func (v *SchemaVerifier) Validate(response string) (bool, error) {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return false, fmt.Errorf("response is not valid JSON: %w", err)
	}

	if !v.hasSchema {
		return true, nil
	}

	// Validate required fields if specified in schema
	if reqFields, ok := v.schemaData["required"].([]any); ok {
		for _, rf := range reqFields {
			fieldName, isStr := rf.(string)
			if !isStr {
				continue
			}
			if _, exists := parsed[fieldName]; !exists {
				return false, fmt.Errorf("missing required schema field: %q", fieldName)
			}
		}
	}

	return true, nil
}
