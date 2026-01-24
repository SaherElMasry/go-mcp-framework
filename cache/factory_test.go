package cache_test

import (
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/cache"
)

// Test: Factory with disabled cache
func TestNew_Disabled(t *testing.T) {
	config := cache.DefaultConfig()
	config.Enabled = false

	c, err := cache.New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should return NoOpCache
	if _, ok := c.(*cache.NoOpCache); !ok {
		t.Error("New() should return NoOpCache when disabled")
	}
}

// Test: Factory with short cache
func TestNew_ShortCache(t *testing.T) {
	config := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     60,
		MaxSize: 1000,
		Enabled: true,
	}

	c, err := cache.New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should return MemoryCache
	if _, ok := c.(*cache.MemoryCache); !ok {
		t.Error("New() should return MemoryCache for TypeShort")
	}
}

// Test: Factory with long cache (not implemented yet)
func TestNew_LongCache(t *testing.T) {
	config := &cache.Config{
		Type:      cache.TypeLong,
		TTL:       30,
		Directory: ".cache",
		Enabled:   true,
	}

	_, err := cache.New(config)
	if err == nil {
		t.Error("New() should return error for TypeLong (not implemented yet)")
	}
}

// Test: Factory with invalid config
func TestNew_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *cache.Config
	}{
		{
			name: "invalid type",
			config: &cache.Config{
				Type:    "invalid",
				TTL:     60,
				Enabled: true,
			},
		},
		{
			name: "negative TTL",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     -1,
				MaxSize: 1000,
				Enabled: true,
			},
		},
		{
			name: "zero max size for short cache",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     60,
				MaxSize: 0,
				Enabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cache.New(tt.config)
			if err == nil {
				t.Error("New() should return error for invalid config")
			}
		})
	}
}

// Test: MustNew panics on error
func TestMustNew_Panics(t *testing.T) {
	config := &cache.Config{
		Type:    "invalid",
		TTL:     60,
		Enabled: true,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew() should panic on error")
		}
	}()

	cache.MustNew(config)
}

// Test: MustNew succeeds
func TestMustNew_Success(t *testing.T) {
	config := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     60,
		MaxSize: 1000,
		Enabled: true,
	}

	c := cache.MustNew(config)
	if c == nil {
		t.Error("MustNew() returned nil")
	}

	if _, ok := c.(*cache.MemoryCache); !ok {
		t.Error("MustNew() should return MemoryCache")
	}
}

// Test: Default config creates NoOp
func TestNew_DefaultConfig(t *testing.T) {
	config := cache.DefaultConfig()

	// Default should be disabled
	if config.Enabled {
		t.Error("DefaultConfig() should have Enabled = false")
	}

	c, err := cache.New(config)
	if err != nil {
		t.Fatalf("New() with DefaultConfig error = %v", err)
	}

	if _, ok := c.(*cache.NoOpCache); !ok {
		t.Error("New() with DefaultConfig should return NoOpCache")
	}
}

// Test: Factory with custom TTL
func TestNew_CustomTTL(t *testing.T) {
	config := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     120, // 2 minutes
		MaxSize: 500,
		Enabled: true,
	}

	c, err := cache.New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	mc, ok := c.(*cache.MemoryCache)
	if !ok {
		t.Fatal("expected MemoryCache")
	}

	// Verify cache was created with correct TTL
	// (We can't directly test TTL, but we can verify it doesn't error)
	if mc == nil {
		t.Error("MemoryCache is nil")
	}
}

// Test: Factory with per-tool TTL
func TestNew_PerToolTTL(t *testing.T) {
	config := &cache.Config{
		Type:    cache.TypeShort,
		TTL:     60,
		MaxSize: 1000,
		Enabled: true,
		ToolTTL: map[string]time.Duration{
			"search":     30 * time.Second,
			"list_files": 5 * time.Minute,
		},
	}

	c, err := cache.New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if c == nil {
		t.Error("New() returned nil")
	}

	// Verify per-tool TTL is stored in config
	if config.GetToolTTL("search") != 30*time.Second {
		t.Error("per-tool TTL not preserved")
	}
}

// Test: Real-world scenarios
func TestNew_RealWorldScenarios(t *testing.T) {
	t.Run("production cache - enabled with monitoring", func(t *testing.T) {
		config := &cache.Config{
			Type:    cache.TypeShort,
			TTL:     300, // 5 minutes
			MaxSize: 10000,
			Enabled: true,
		}

		c, err := cache.New(config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if _, ok := c.(*cache.MemoryCache); !ok {
			t.Error("should return MemoryCache for production")
		}
	})

	t.Run("development - cache disabled", func(t *testing.T) {
		config := cache.DefaultConfig()
		config.Enabled = false

		c, err := cache.New(config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if _, ok := c.(*cache.NoOpCache); !ok {
			t.Error("should return NoOpCache when disabled")
		}
	})

	t.Run("testing - small cache", func(t *testing.T) {
		config := &cache.Config{
			Type:    cache.TypeShort,
			TTL:     10,
			MaxSize: 10,
			Enabled: true,
		}

		c, err := cache.New(config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if c == nil {
			t.Error("New() returned nil")
		}
	})
}
