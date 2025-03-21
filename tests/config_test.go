package config_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/centraunit/gonfig"
	configContracts "github.com/centraunit/gonfig/contracts"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite is the test suite for the config package
type ConfigTestSuite struct {
	suite.Suite
	registry configContracts.ConfigRegistry
}

// SetupTest sets up the test environment before each test
func (suite *ConfigTestSuite) SetupTest() {
	registry, err := gonfig.GetConfigRegistry("testing")
	suite.NoError(err)
	suite.registry = registry
	suite.registry.Register("testget", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return map[string]interface{}{
			"value": "testget",
		}
	})

	// Register a test config section with deep nesting
	suite.registry.Register("test", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return map[string]interface{}{
			"string_value":     "test",
			"int_value":        42,
			"bool_value":       true,
			"float_value":      3.14,
			"array_value":      []string{"one", "two", "three"},
			"string_for_array": "one,two,three",
			"nested": map[string]interface{}{
				"key": "value",
				"deep": map[string]interface{}{
					"deeper": map[string]interface{}{
						"deepest": "found",
						"numbers": []int{1, 2, 3},
						"config": map[string]interface{}{
							"enabled": true,
							"rate":    0.75,
							"tags":    []string{"test", "deep", "nesting"},
						},
					},
				},
			},
		}
	})
}

// TearDownTest cleans up the test environment after each test
func (suite *ConfigTestSuite) TearDownTest() {
	os.Clearenv()
}

// TestGetString tests retrieving a string value from the configuration map
func (suite *ConfigTestSuite) TestGetString() {
	// Test direct value
	value, err := suite.registry.GetString("test.string_value")
	suite.NoError(err)
	suite.Equal("test", value)

	// Test with default value
	value, err = suite.registry.GetString("test.nonexistent", "default")
	suite.NoError(err)
	suite.Equal("default", value)

	// Test invalid path
	_, err = suite.registry.GetString("invalid.path")
	suite.Error(err)
}

// TestGetInt tests retrieving an integer value from the configuration map
func (suite *ConfigTestSuite) TestGetInt() {
	// Test direct value
	value, err := suite.registry.GetInt("test.int_value")
	suite.NoError(err)
	suite.Equal(42, value)

	// Test with default value
	value, err = suite.registry.GetInt("test.nonexistent", 5000)
	suite.NoError(err)
	suite.Equal(5000, value)

	// Test invalid value
	_, err = suite.registry.GetInt("test.string_value")
	suite.Error(err)
}

// TestGetBool tests retrieving a boolean value from the configuration map
func (suite *ConfigTestSuite) TestGetBool() {
	// Test direct value
	value, err := suite.registry.GetBool("test.bool_value")
	suite.NoError(err)
	suite.Equal(true, value)

	// Test with default value
	value, err = suite.registry.GetBool("test.nonexistent", true)
	suite.NoError(err)
	suite.Equal(true, value)

	// Test invalid value
	_, err = suite.registry.GetBool("test.string_value")
	suite.Error(err)
}

// TestGetFloat tests retrieving a float value from the configuration map
func (suite *ConfigTestSuite) TestGetFloat() {
	// Test direct value
	value, err := suite.registry.GetFloat("test.float_value")
	suite.NoError(err)
	suite.Equal(3.14, value)

	// Test int to float conversion
	value, err = suite.registry.GetFloat("test.int_value")
	suite.NoError(err)
	suite.Equal(42.0, value)

	// Test with default value
	value, err = suite.registry.GetFloat("test.nonexistent", 2.718)
	suite.NoError(err)
	suite.Equal(2.718, value)

	// Test invalid value
	_, err = suite.registry.GetFloat("test.string_value")
	suite.Error(err)
}

// TestGetStringArray tests retrieving a string array from the configuration map
func (suite *ConfigTestSuite) TestGetStringArray() {
	// Test direct array value
	value, err := suite.registry.GetStringArray("test.array_value")
	suite.NoError(err)
	suite.Equal([]string{"one", "two", "three"}, value)

	// Test comma-separated string value
	value, err = suite.registry.GetStringArray("test.string_for_array")
	suite.NoError(err)
	suite.Equal([]string{"one", "two", "three"}, value)

	// Test with default value for nonexistent path
	value, err = suite.registry.GetStringArray("test.nonexistent", []string{"default"})
	suite.NoError(err)
	suite.Equal([]string{"default"}, value)

	// Test invalid type - should return error when trying to convert non-string/non-array value
	_, err = suite.registry.GetStringArray("test.int_value")
	suite.Error(err, "Expected error when converting int to string array")
}

// TestSet tests setting a value in the configuration map
func (suite *ConfigTestSuite) TestSet() {
	// Test setting new value
	err := suite.registry.Set("test.new_value", "test")
	suite.NoError(err)
	value, err := suite.registry.GetString("test.new_value")
	suite.NoError(err)
	suite.Equal("test", value)

	// Test overwriting existing value
	err = suite.registry.Set("test.string_value", "updated")
	suite.NoError(err)
	value, err = suite.registry.GetString("test.string_value")
	suite.NoError(err)
	suite.Equal("updated", value)

	// Test invalid path
	err = suite.registry.Set("invalid", "value")
	suite.Error(err)
}

// TestDotNotation tests retrieving values from the configuration map using dot notation
func (suite *ConfigTestSuite) TestDotNotation() {
	// Test deep nested string
	value, err := suite.registry.GetString("test.nested.deep.deeper.deepest")
	suite.NoError(err)
	suite.Equal("found", value)

	// Test deep nested bool
	boolVal, err := suite.registry.GetBool("test.nested.deep.deeper.config.enabled")
	suite.NoError(err)
	suite.Equal(true, boolVal)

	// Test deep nested float
	floatVal, err := suite.registry.GetFloat("test.nested.deep.deeper.config.rate")
	suite.NoError(err)
	suite.Equal(0.75, floatVal)

	// Test deep nested array
	arrayVal, err := suite.registry.GetStringArray("test.nested.deep.deeper.config.tags")
	suite.NoError(err)
	suite.Equal([]string{"test", "deep", "nesting"}, arrayVal)

	// Test invalid deep path
	_, err = suite.registry.GetString("test.nested.deep.invalid.path")
	suite.Error(err)
	suite.Contains(err.Error(), "key not found: 'nested.deep.invalid' in path 'test.nested.deep.invalid.path'")
}

// TestRefresh tests refreshing the configuration map
func (suite *ConfigTestSuite) TestRefresh() {
	// Get initial value
	value, err := suite.registry.GetString("test.string_value")
	suite.NoError(err)
	suite.Equal("test", value)

	// Register new config that returns different value
	suite.registry.Register("test", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return map[string]interface{}{
			"string_value": "updated",
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"value": "also_updated",
				},
			},
		}
	})

	// Refresh configs
	suite.registry.Refresh()

	// Value should be updated
	value, err = suite.registry.GetString("test.string_value")
	suite.NoError(err)
	suite.Equal("updated", value)

	// Deep nested value should also be updated
	value, err = suite.registry.GetString("test.nested.deep.value")
	suite.NoError(err)
	suite.Equal("also_updated", value)
}

// TestConfigSuite runs the test suite
func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// Test Get to get full config without dot notation
func (suite *ConfigTestSuite) TestGet() {
	value, err := suite.registry.Get("testget")
	suite.NoError(err)
	suite.Equal(map[string]interface{}{
		"value": "testget",
	}, value)
}

// TestErrorHandling tests various error scenarios
func (suite *ConfigTestSuite) TestErrorHandling() {
	// Test invalid path format
	_, err := suite.registry.Get("invalid")
	suite.Error(err)
	suite.Contains(err.Error(), "config section not found: 'invalid' in path 'invalid'")

	// Test nonexistent section
	_, err = suite.registry.Get("nonexistent.key")
	suite.Error(err)
	suite.Contains(err.Error(), "config section not found: 'nonexistent' in path 'nonexistent.key'")

	// Test nonexistent nested key
	_, err = suite.registry.Get("test.nonexistent.key")
	suite.Error(err)
	suite.Contains(err.Error(), "key not found: 'nonexistent' in path 'test.nonexistent.key'")

	// Test type conversion errors
	_, err = suite.registry.GetInt("test.string_value")
	suite.Error(err)
	suite.Contains(err.Error(), "cannot convert value 'test' at path 'test.string_value' to int")

	_, err = suite.registry.GetBool("test.int_value")
	suite.Error(err)
	suite.Contains(err.Error(), "cannot convert value at path 'test.int_value' to bool: found type int")

	_, err = suite.registry.GetFloat("test.bool_value")
	suite.Error(err)
	suite.Contains(err.Error(), "cannot convert value at path 'test.bool_value' to float64: found type bool")

	// Test array conversion errors
	_, err = suite.registry.GetStringArray("test.int_value")
	suite.Error(err)
	suite.Contains(err.Error(), "cannot convert value at path 'test.int_value' to string array: found type int")

	// Test mixed type array error
	suite.registry.Register("test_arrays", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return map[string]interface{}{
			"mixed_array": []interface{}{"string", 123, true},
		}
	})
	_, err = suite.registry.GetStringArray("test_arrays.mixed_array")
	suite.Error(err)
	suite.Contains(err.Error(), "cannot convert item at index 1 in path 'test_arrays.mixed_array' to string: found type int")
}

// TestDefaultValues tests default value handling
func (suite *ConfigTestSuite) TestDefaultValues() {
	// Test string default
	value, err := suite.registry.GetString("test.nonexistent", "default")
	suite.NoError(err)
	suite.Equal("default", value)

	// Test int default
	intVal, err := suite.registry.GetInt("test.nonexistent", 42)
	suite.NoError(err)
	suite.Equal(42, intVal)

	// Test bool default
	boolVal, err := suite.registry.GetBool("test.nonexistent", true)
	suite.NoError(err)
	suite.Equal(true, boolVal)

	// Test float default
	floatVal, err := suite.registry.GetFloat("test.nonexistent", 3.14)
	suite.NoError(err)
	suite.Equal(3.14, floatVal)

	// Test string array default
	arrayVal, err := suite.registry.GetStringArray("test.nonexistent", []string{"default"})
	suite.NoError(err)
	suite.Equal([]string{"default"}, arrayVal)

	// Test no default provided
	_, err = suite.registry.GetString("test.nonexistent")
	suite.Error(err)
	suite.Contains(err.Error(), "key not found: 'nonexistent' in path 'test.nonexistent'")
}

// TestSetErrors tests error cases for Set operation
func (suite *ConfigTestSuite) TestSetErrors() {
	// Test invalid path
	err := suite.registry.Set("invalid", "value")
	suite.Error(err)
	suite.Contains(err.Error(), "invalid config path")

	// Test nonexistent section
	err = suite.registry.Set("nonexistent.key", "value")
	suite.Error(err)
	suite.Contains(err.Error(), "config section not found")

	// Test setting value in nonexistent nested path
	err = suite.registry.Set("test.very.deep.nested.key", "value")
	suite.NoError(err) // Should create intermediate maps

	// Verify the deep path was created
	value, err := suite.registry.GetString("test.very.deep.nested.key")
	suite.NoError(err)
	suite.Equal("value", value)
}

// TestRefreshErrors tests error cases for Refresh operation
func (suite *ConfigTestSuite) TestRefreshErrors() {
	// Register a loader that returns nil
	suite.registry.Register("test_nil", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return nil
	})

	// Refresh should handle nil configs
	suite.registry.Refresh()

	// Attempting to get value from nil config should error
	_, err := suite.registry.Get("test_nil.key")
	suite.Error(err)

	// Register a loader that panics
	suite.registry.Register("test_panic", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		panic("test panic")
	})

	// Refresh should recover from panics
	suite.NotPanics(func() {
		suite.registry.Refresh()
	})
}

// TestUnmarshal tests deserializing a configuration section into a struct
func (suite *ConfigTestSuite) TestUnmarshal() {
	type DatabaseConfig struct {
		Host     string `config:"host" required:"true"`
		Port     int    `config:"port" required:"true"`
		Username string `config:"username"`
		Password string `config:"password"`
		Options  struct {
			MaxConn int  `config:"max_connections"`
			Debug   bool `config:"debug_mode"`
		} `config:"options"`
	}

	suite.registry.Register("database", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "admin",
			"options": map[string]interface{}{
				"max_connections": 100,
				"debug_mode":      true,
			},
		}
	})

	var config DatabaseConfig
	err := suite.registry.Unmarshal("database", &config)
	suite.NoError(err)
	suite.Equal("localhost", config.Host)
	suite.Equal(5432, config.Port)
	suite.Equal("admin", config.Username)
	suite.Equal("", config.Password) // Optional field
	suite.Equal(100, config.Options.MaxConn)
	suite.Equal(true, config.Options.Debug)
}

// TestSchemaValidation tests the schema validation functionality
func (suite *ConfigTestSuite) TestSchemaValidation() {
	schema := gonfig.NewConfigSchema()

	// Add schema fields
	schema.AddField("test.string_value", configContracts.ConfigSchemaField{
		Type:     reflect.String,
		Required: true,
	})

	schema.AddField("test.int_value", configContracts.ConfigSchemaField{
		Type:     reflect.Int,
		Required: true,
	})

	schema.AddField("test.optional_value", configContracts.ConfigSchemaField{
		Type:    reflect.String,
		Default: "default",
	})

	schema.AddField("test.validated_value", configContracts.ConfigSchemaField{
		Type:     reflect.Int,
		Required: true,
		Validator: func(v interface{}) error {
			if val, ok := v.(int); ok {
				if val < 0 || val > 100 {
					return fmt.Errorf("value must be between 0 and 100")
				}
			}
			return nil
		},
	})

	// Test valid configuration
	validConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"string_value":    "test",
			"int_value":       42,
			"validated_value": 50,
		},
	}
	err := schema.Validate(validConfig)
	suite.NoError(err)

	// Test wrong type
	wrongTypeConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"string_value":    123, // should be string
			"int_value":       42,
			"validated_value": 50,
		},
	}
	err = schema.Validate(wrongTypeConfig)
	suite.Error(err)
	suite.Contains(err.Error(), "validation failed for test.string_value: expected type string")

	// Test missing required field
	invalidConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"string_value": "test",
			"int_value":    42,
			// missing validated_value
		},
	}
	err = schema.Validate(invalidConfig)
	suite.Error(err)
	suite.Contains(err.Error(), "required field missing: test.validated_value")

	// Test validation function
	invalidValueConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"string_value":    "test",
			"int_value":       42,
			"validated_value": 200, // should be between 0 and 100
		},
	}
	err = schema.Validate(invalidValueConfig)
	suite.Error(err)
	suite.Contains(err.Error(), "validation failed for test.validated_value: value must be between 0 and 100")

	// Test default value
	defaultConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"string_value":    "test",
			"int_value":       42,
			"validated_value": 50,
		},
	}
	err = schema.Validate(defaultConfig)
	suite.NoError(err)

	// Verify default value was set
	value, exists := defaultConfig["test"].(map[string]interface{})["optional_value"]
	suite.True(exists)
	suite.Equal("default", value)
}

// TestSchemaFieldValidation tests individual schema field validation
func (suite *ConfigTestSuite) TestSchemaFieldValidation() {
	schema := gonfig.NewConfigSchema()

	// Add a field with validation
	schema.AddField("test.validated", configContracts.ConfigSchemaField{
		Type:     reflect.String,
		Required: true,
		Validator: func(v interface{}) error {
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("value is not a string")
			}
			if len(str) < 3 {
				return fmt.Errorf("string length must be at least 3")
			}
			return nil
		},
	})

	// Test valid value
	validConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"validated": "test",
		},
	}
	err := schema.Validate(validConfig)
	suite.NoError(err)

	// Test nil value for required field
	nilConfig := map[string]interface{}{
		"test": map[string]interface{}{},
	}
	err = schema.Validate(nilConfig)
	suite.Error(err)
	suite.Contains(err.Error(), "required field missing")

	// Test wrong type
	wrongTypeConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"validated": 123,
		},
	}
	err = schema.Validate(wrongTypeConfig)
	suite.Error(err)
	suite.Contains(err.Error(), "expected type string")

	// Test validation function
	invalidLengthConfig := map[string]interface{}{
		"test": map[string]interface{}{
			"validated": "ab",
		},
	}
	err = schema.Validate(invalidLengthConfig)
	suite.Error(err)
	suite.Contains(err.Error(), "string length must be at least 3")
}

// TestSingletonBehavior tests the singleton behavior of the config registry
func (suite *ConfigTestSuite) TestSingletonBehavior() {
	// Get first instance
	registry1, err := gonfig.GetConfigRegistry("testing")
	suite.NoError(err)
	suite.NotNil(registry1)

	// Get second instance
	registry2, err := gonfig.GetConfigRegistry("testing")
	suite.NoError(err)
	suite.NotNil(registry2)

	// Verify they're the same instance
	suite.Equal(fmt.Sprintf("%p", registry1), fmt.Sprintf("%p", registry2))

	// Verify changes in one affect the other
	err = registry1.Set("test.value", "changed")
	suite.NoError(err)

	value, err := registry2.GetString("test.value")
	suite.NoError(err)
	suite.Equal("changed", value)
}
