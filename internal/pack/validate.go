package pack

import (
	"fmt"
	"strings"
)

func (p *Pack) ValidateMetadata(label string, metadata map[string]any) error {
	if p == nil || label == "" || len(p.MetadataSchemas) == 0 {
		return nil
	}
	schema, ok := p.MetadataSchemas[label]
	if !ok {
		return nil
	}
	if metadata == nil {
		metadata = map[string]any{}
	}

	if required, ok := schema["required"].([]any); ok {
		for _, field := range required {
			name, _ := field.(string)
			if name == "" {
				continue
			}
			if _, exists := metadata[name]; !exists {
				return fmt.Errorf("metadata.%s is required for label %q", name, label)
			}
		}
	}

	props, _ := schema["properties"].(map[string]any)
	for key, value := range metadata {
		propSchema, ok := props[key].(map[string]any)
		if !ok {
			continue
		}
		if err := validateProperty(key, value, propSchema); err != nil {
			return fmt.Errorf("metadata for label %q: %w", label, err)
		}
	}
	return nil
}

func validateProperty(name string, value any, schema map[string]any) error {
	expectedType, _ := schema["type"].(string)
	switch expectedType {
	case "string":
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", name)
		}
		if enumValues, ok := schema["enum"].([]any); ok {
			if !enumContains(enumValues, s) {
				return fmt.Errorf("%s must be one of %v", name, enumValues)
			}
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int64:
		default:
			return fmt.Errorf("%s must be a number", name)
		}
	case "array":
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s must be an array", name)
		}
		if itemSchema, ok := schema["items"].(map[string]any); ok {
			for i, item := range items {
				if err := validateProperty(fmt.Sprintf("%s[%d]", name, i), item, itemSchema); err != nil {
					return err
				}
			}
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("%s must be an object", name)
		}
	}
	return nil
}

func enumContains(values []any, target string) bool {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}
