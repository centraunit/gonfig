package gonfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	configContracts "github.com/centraunit/gonfig/contracts"
	"github.com/joho/godotenv"
)

var (
	globalConfigRegistry     configContracts.ConfigRegistry
	globalConfigRegistryOnce sync.Once
)

// ConfigRegistry provides a thread-safe registry for managing configuration values.
// It supports dot notation access, type conversion, and dynamic reloading of configurations.
type ConfigRegistry struct {
	configs map[string]map[string]interface{}
	loaders map[string]configContracts.ConfigLoader
	mu      sync.RWMutex
}

// GetConfigRegistry creates a new instance of ConfigRegistry.
// It initializes the internal maps for storing configurations and their loaders.
func GetConfigRegistry(env string) (configContracts.ConfigRegistry, error) {
	var initErr error
	globalConfigRegistryOnce.Do(func() {
		if env == "" {
			initErr = fmt.Errorf("env is required when initializing config registry")
			return
		}

		// Load appropriate env file
		if env == "development" || env == "staging" || env == "production" {
			if err := godotenv.Load(".env"); err != nil {
				initErr = fmt.Errorf("error loading .env file: %w", err)
				return
			}
		} else if env == "testing" {
			if err := godotenv.Load(".env.testing"); err != nil {
				initErr = fmt.Errorf("error loading .env.testing file: %w", err)
				return
			}
		} else {
			initErr = fmt.Errorf("invalid env: %s", env)
			return
		}

		globalConfigRegistry = &ConfigRegistry{
			configs: make(map[string]map[string]interface{}),
			loaders: make(map[string]configContracts.ConfigLoader),
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return globalConfigRegistry, nil
}

// Register adds a new configuration section with its loader function.
// The loader function will be called immediately to populate the initial configuration,
// and can be called again during Refresh operations.
func (r *ConfigRegistry) Register(name string, loader configContracts.ConfigLoader) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.loaders[name] = loader

	// Recover from panics in loader
	defer func() {
		if rec := recover(); rec != nil {
			r.configs[name] = make(map[string]interface{})
		}
	}()

	r.configs[name] = loader(r)
}

// Refresh reloads all configurations using their registered loader functions.
// This is useful when configuration sources (like environment variables) have changed.
func (r *ConfigRegistry) Refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, loader := range r.loaders {
		// Recover from panics for each loader
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					if _, exists := r.configs[name]; !exists {
						r.configs[name] = make(map[string]interface{})
					}
				}
			}()
			r.configs[name] = loader(r)
		}()
	}

}

// Get retrieves a value from the configuration using dot notation.
// Returns an error if the path is invalid or the value doesn't exist.
// Example: Get("database.connections.mysql.host")
func (r *ConfigRegistry) Get(path string) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Normal lookup
	value, err := r.lookup(path)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// lookup performs the actual configuration lookup
func (r *ConfigRegistry) lookup(path string) (interface{}, error) {
	parts := strings.Split(path, ".")

	section := parts[0]
	config, ok := r.configs[section]
	if !ok {
		return nil, fmt.Errorf("config section not found: '%s' in path '%s'", section, path)
	}

	if config == nil {
		return nil, fmt.Errorf("config section is nil: '%s' in path '%s'", section, path)
	}
	if len(parts) == 1 {
		return config, nil
	}
	return traverse(config, parts[1:], path)
}

// Set updates a configuration value using dot notation.
// Returns an error if the path is invalid or the section doesn't exist.
// Example: Set("app.name", "MyApp")
func (r *ConfigRegistry) Set(path string, value interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid config path: %s", path)
	}

	section := parts[0]
	config, ok := r.configs[section]
	if !ok {
		return fmt.Errorf("config section not found: %s", section)
	}

	return setValue(config, parts[1:], value)
}

// GetString retrieves a string value from the configuration.
// Accepts optional default value to be returned if the path doesn't exist.
// Returns an error if the value cannot be converted to string.
func (r *ConfigRegistry) GetString(path string, defaultValue ...string) (string, error) {
	value, err := r.Get(path)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value at %s is not a string", path)
	}

	return str, nil
}

// GetInt retrieves an integer value from the configuration.
// Accepts optional default value to be returned if the path doesn't exist.
// Supports conversion from string and float64 values.
// Returns an error if the value cannot be converted to int.
func (r *ConfigRegistry) GetInt(path string, defaultValue ...int) (int, error) {
	value, err := r.Get(path)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("cannot convert value '%v' at path '%s' to int: %v", v, path, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert value at path '%s' to int: found type %T", path, value)
	}
}

// GetBool retrieves a boolean value from the configuration.
// Accepts optional default value to be returned if the path doesn't exist.
// Supports conversion from string values ("true"/"false").
// Returns an error if the value cannot be converted to bool.
func (r *ConfigRegistry) GetBool(path string, defaultValue ...bool) (bool, error) {
	value, err := r.Get(path)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("cannot convert value '%v' at path '%s' to bool: %v", v, path, err)
		}
		return b, nil
	default:
		return false, fmt.Errorf("cannot convert value at path '%s' to bool: found type %T", path, value)
	}
}

// GetFloat retrieves a float64 value from the configuration.
// Accepts optional default value to be returned if the path doesn't exist.
// Supports conversion from string and int values.
// Returns an error if the value cannot be converted to float64.
func (r *ConfigRegistry) GetFloat(path string, defaultValue ...float64) (float64, error) {
	value, err := r.Get(path)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert value '%v' at path '%s' to float64: %v", v, path, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert value at path '%s' to float64: found type %T", path, value)
	}
}

// GetStringArray retrieves a string array from the configuration.
// Accepts optional default value to be returned if the path doesn't exist.
// Supports conversion from comma-separated strings and []interface{} values.
// Returns an error if the value cannot be converted to []string.
func (r *ConfigRegistry) GetStringArray(path string, defaultValue ...[]string) ([]string, error) {
	value, err := r.Get(path)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return nil, err
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case string:
		if v == "" {
			return []string{}, nil
		}
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("cannot convert item at index %d in path '%s' to string: found type %T", i, path, item)
			}
			result[i] = str
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert value at path '%s' to string array: found type %T", path, value)
	}
}

// GetEnvString retrieves a string value from environment variables.
// Returns the default value if the environment variable doesn't exist.
func (r *ConfigRegistry) GetEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an integer value from environment variables.
// Returns the default value if the environment variable doesn't exist or cannot be converted.
func (r *ConfigRegistry) GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvBool retrieves a boolean value from environment variables.
// Returns the default value if the environment variable doesn't exist.
// The value "true" (case-insensitive) is considered true, all other values are false.
func (r *ConfigRegistry) GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// GetEnvStringArray retrieves a string array from environment variables.
// Returns the default value if the environment variable doesn't exist.
// The value is split on commas and each part is trimmed of whitespace.
func (r *ConfigRegistry) GetEnvStringArray(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		parts := strings.Split(value, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return defaultValue
}

// Internal helper functions

// traverse walks through a nested configuration map using the given path parts.
// It returns the value at the specified path or an error if the path is invalid.
// Example: traverse(config, []string{"database", "host"})
func traverse(config map[string]interface{}, parts []string, fullPath string) (interface{}, error) {
	current := config
	for i, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]interface{})
		if !ok {
			currentPath := strings.Join(parts[:i+1], ".")
			if _, exists := current[part]; !exists {
				return nil, fmt.Errorf("key not found: '%s' in path '%s'", currentPath, fullPath)
			}
			return nil, fmt.Errorf("value at '%s' in path '%s' is not a map, cannot traverse further", currentPath, fullPath)
		}
		current = next
	}

	lastPart := parts[len(parts)-1]
	value, ok := current[lastPart]
	if !ok {
		return nil, fmt.Errorf("key not found: '%s' in path '%s'", lastPart, fullPath)
	}

	return value, nil
}

// setValue updates a value in a nested configuration map using the given path parts.
// It creates intermediate maps if they don't exist.
// Example: setValue(config, []string{"database", "host"}, "localhost")
func setValue(config map[string]interface{}, parts []string, value interface{}) error {
	current := config
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			current[part] = next
		}
		current = next
	}

	current[parts[len(parts)-1]] = value
	return nil
}

// Unmarshal deserializes a configuration section into a struct
func (r *ConfigRegistry) Unmarshal(section string, v interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.configs[section]
	if !ok {
		return fmt.Errorf("config section not found: '%s'", section)
	}

	// Use reflection to map config values to struct fields
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshal target must be a non-nil pointer")
	}

	return unmarshalInto(config, val.Elem())
}

// UnmarshalKey deserializes a specific configuration key into a struct
func (r *ConfigRegistry) UnmarshalKey(path string, v interface{}) error {
	value, err := r.Get(path)
	if err != nil {
		return err
	}

	configMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("value at '%s' is not a map", path)
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshal target must be a non-nil pointer")
	}

	return unmarshalInto(configMap, val.Elem())
}

// Helper function to unmarshal config into a struct
func unmarshalInto(config map[string]interface{}, val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Get the config key from struct tag or field name
		key := field.Tag.Get("config")
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		if key == "-" {
			continue // Skip this field
		}

		value, ok := config[key]
		if !ok {
			// Check if field is required
			if field.Tag.Get("required") == "true" {
				return fmt.Errorf("required field '%s' not found in configuration", key)
			}
			continue
		}

		if err := setField(fieldVal, value); err != nil {
			return fmt.Errorf("error setting field '%s': %w", key, err)
		}
	}

	return nil
}

// setField sets a value to a struct field using reflection
func setField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		str, err := toString(value)
		if err != nil {
			return err
		}
		field.SetString(str)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := toInt64(value)
		if err != nil {
			return err
		}
		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := toUint64(value)
		if err != nil {
			return err
		}
		field.SetUint(i)

	case reflect.Float32, reflect.Float64:
		f, err := toFloat64(value)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := toBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			s, err := toStringSlice(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(s))
		} else {
			return fmt.Errorf("unsupported slice type: %v", field.Type())
		}

	case reflect.Struct:
		if m, ok := value.(map[string]interface{}); ok {
			return unmarshalInto(m, field)
		}
		return fmt.Errorf("cannot set struct field with value of type %T", value)

	default:
		return fmt.Errorf("unsupported field type: %v", field.Type())
	}

	return nil
}

// Helper functions for type conversion
func toString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func toUint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case uint:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative float64 to uint64")
		}
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", value)
	}
}

func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func toStringSlice(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case string:
		if v == "" {
			return []string{}, nil
		}
		return strings.Split(v, ","), nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			str, err := toString(item)
			if err != nil {
				return nil, fmt.Errorf("cannot convert item at index %d: %w", i, err)
			}
			result[i] = str
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []string", value)
	}
}
