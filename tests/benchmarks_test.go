package config_test

import (
	"os"
	"testing"

	"github.com/centraunit/gonfig"
	configContracts "github.com/centraunit/gonfig/contracts"
)

// BenchmarkConfigRegistry benchmarks all operations of the config package
func BenchmarkConfigRegistry(b *testing.B) {
	registry, err := gonfig.GetConfigRegistry("testing")
	if err != nil {
		b.Fatalf("error creating config registry: %s", err)
	}
	registry.Register("test", func(registry configContracts.ConfigRegistry) map[string]interface{} {
		return map[string]interface{}{
			"string_value": "test",
			"int_value":    42,
			"bool_value":   true,
			"float_value":  3.14,
			"array_value":  []string{"one", "two", "three"},
			"nested": map[string]interface{}{
				"key": "value",
				"deep": map[string]interface{}{
					"deeper": map[string]interface{}{
						"deepest": "found",
					},
				},
			},
		}
	})

	b.Run("Get/Simple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.Get("test.string_value")
		}
	})

	b.Run("Get/Deep", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.Get("test.nested.deep.deeper.deepest")
		}
	})

	b.Run("GetString/Simple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetString("test.string_value")
		}
	})

	b.Run("GetString/WithDefault", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetString("test.nonexistent", "default")
		}
	})

	b.Run("GetInt/Direct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetInt("test.int_value")
		}
	})

	b.Run("GetInt/FromString", func(b *testing.B) {
		registry.Set("test.string_int", "42")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetInt("test.string_int")
		}
	})

	b.Run("GetBool/Direct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetBool("test.bool_value")
		}
	})

	b.Run("GetBool/FromString", func(b *testing.B) {
		registry.Set("test.string_bool", "true")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetBool("test.string_bool")
		}
	})

	b.Run("GetFloat/Direct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetFloat("test.float_value")
		}
	})

	b.Run("GetFloat/FromInt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetFloat("test.int_value")
		}
	})

	b.Run("GetStringArray/Direct", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetStringArray("test.array_value")
		}
	})

	b.Run("GetStringArray/FromString", func(b *testing.B) {
		registry.Set("test.string_array", "one,two,three")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetStringArray("test.string_array")
		}
	})

	b.Run("Set/Simple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.Set("test.benchmark_value", i)
		}
	})

	b.Run("Set/Deep", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.Set("test.nested.deep.deeper.benchmark", i)
		}
	})

	b.Run("Refresh", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			registry.Refresh()
		}
	})
}

// BenchmarkEnvironmentVariables benchmarks environment variable operations
func BenchmarkEnvironmentVariables(b *testing.B) {
	registry, err := gonfig.GetConfigRegistry("testing")
	if err != nil {
		b.Fatalf("error creating config registry: %s", err)
	}
	os.Setenv("TEST_STRING", "value")
	os.Setenv("TEST_INT", "42")
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_ARRAY", "one,two,three")
	defer func() {
		os.Unsetenv("TEST_STRING")
		os.Unsetenv("TEST_INT")
		os.Unsetenv("TEST_BOOL")
		os.Unsetenv("TEST_ARRAY")
	}()

	b.Run("GetEnvString/Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvString("TEST_STRING", "default")
		}
	})

	b.Run("GetEnvString/Default", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvString("TEST_NONEXISTENT", "default")
		}
	})

	b.Run("GetEnvInt/Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvInt("TEST_INT", 0)
		}
	})

	b.Run("GetEnvInt/Default", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvInt("TEST_NONEXISTENT", 0)
		}
	})

	b.Run("GetEnvBool/Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvBool("TEST_BOOL", false)
		}
	})

	b.Run("GetEnvBool/Default", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvBool("TEST_NONEXISTENT", false)
		}
	})

	b.Run("GetEnvStringArray/Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvStringArray("TEST_ARRAY", nil)
		}
	})

	b.Run("GetEnvStringArray/Default", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetEnvStringArray("TEST_NONEXISTENT", []string{"default"})
		}
	})
}
