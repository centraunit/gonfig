package config_test

import (
	"strings"
	"testing"

	"github.com/centraunit/gonfig"
	"github.com/stretchr/testify/assert"
)

func TestPathCache(t *testing.T) {
	pc := gonfig.NewPathCache()

	// Test basic path splitting
	path := "database.connections.mysql.host"
	expected := []string{"database", "connections", "mysql", "host"}
	result := pc.Get(path)
	assert.Equal(t, expected, result)

	// Test cache hit (same result, should be cached)
	result2 := pc.Get(path)
	assert.Equal(t, expected, result2)
	assert.Equal(t, &result[0], &result2[0], "Should return same slice from cache")
}

func BenchmarkPathCache(b *testing.B) {
	pc := gonfig.NewPathCache()
	paths := []string{
		"database.connections.mysql.host",
		"app.name",
		"logging.channels.stack.drivers",
		"mail.mailers.smtp.encryption",
	}

	b.Run("WithCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pc.Get(paths[i%len(paths)])
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = strings.Split(paths[i%len(paths)], ".")
		}
	})
}
