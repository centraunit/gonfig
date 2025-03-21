package gonfig

import (
	"fmt"
	"reflect"
	"strings"

	configContracts "github.com/centraunit/gonfig/contracts"
)

// ConfigSchemaField represents a field in the configuration schema
type ConfigSchemaField struct {
	Type      reflect.Kind
	Required  bool
	Default   interface{}
	Validator func(interface{}) error
}

// Schema defines the structure and validation rules for configuration
type ConfigSchema struct {
	Fields map[string]configContracts.ConfigSchemaField
}

// NewConfigSchema creates a new schema instance
func NewConfigSchema() configContracts.ConfigSchema {
	return &ConfigSchema{
		Fields: make(map[string]configContracts.ConfigSchemaField),
	}
}

// AddField adds a field to the schema
func (s *ConfigSchema) AddField(path string, field configContracts.ConfigSchemaField) {
	s.Fields[path] = field

}

// Validate checks if a configuration matches the schema
func (s *ConfigSchema) Validate(config map[string]interface{}) error {
	for path, field := range s.Fields {
		parts := strings.Split(path, ".")
		value, err := traverse(config, parts, path)
		if err != nil {
			if field.Required {
				return fmt.Errorf("required field missing: %s", path)
			}
			if field.Default != nil {
				if err := setValue(config, parts, field.Default); err != nil {
					return fmt.Errorf("failed to set default value for %s: %w", path, err)
				}
			}
			continue
		}

		if err := validateValue(value, field); err != nil {
			return fmt.Errorf("validation failed for %s: %w", path, err)
		}
	}
	return nil
}

// validateValue checks if a value matches the schema field requirements
func validateValue(value interface{}, field configContracts.ConfigSchemaField) error {
	if value == nil {
		if field.Required {
			return fmt.Errorf("required field is nil")
		}
		return nil
	}

	valueType := reflect.TypeOf(value).Kind()
	if valueType != field.Type {
		return fmt.Errorf("expected type %v, got %v", field.Type, valueType)
	}

	if field.Validator != nil {
		if err := field.Validator(value); err != nil {
			return err
		}
	}

	return nil
}
