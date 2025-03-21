package contracts

import "reflect"

// ConfigLoader is a function type that returns configuration values
type ConfigLoader func(registry ConfigRegistry) map[string]interface{}

// ConfigRegistry defines the interface for configuration management
type ConfigRegistry interface {
	// Core operations
	Get(path string) (interface{}, error)
	GetString(path string, defaultValue ...string) (string, error)
	GetInt(path string, defaultValue ...int) (int, error)
	GetBool(path string, defaultValue ...bool) (bool, error)
	GetFloat(path string, defaultValue ...float64) (float64, error)
	GetStringArray(path string, defaultValue ...[]string) ([]string, error)
	Set(path string, value interface{}) error
	Register(name string, loader ConfigLoader)
	Refresh()
	Unmarshal(section string, v interface{}) error
	UnmarshalKey(path string, v interface{}) error
	GetEnvString(key string, defaultValue string) string
	GetEnvInt(key string, defaultValue int) int
	GetEnvBool(key string, defaultValue bool) bool
	GetEnvStringArray(key string, defaultValue []string) []string
}

// Schema defines the interface for configuration validation
type ConfigSchema interface {
	AddField(path string, field ConfigSchemaField)
	Validate(config map[string]interface{}) error
}

// SchemaField represents a field in the configuration schema
type ConfigSchemaField struct {
	Type      reflect.Kind
	Required  bool
	Default   interface{}
	Validator func(interface{}) error
}

// PathCache defines the interface for path caching operations
type ConfigPathCache interface {
	Get(path string) []string
}
