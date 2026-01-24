package cache_test

import (
	"strings"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/cache"
)

// Test: DefaultConfig
func TestDefaultConfig(t *testing.T) {
	cfg := cache.DefaultConfig()

	// ⚠️ CRITICAL: Verify cache is DISABLED by default (safe default)
	if cfg.Enabled {
		t.Error("default config should have caching DISABLED for safety")
	}

	if cfg.Type != cache.TypeShort {
		t.Errorf("Type = %v, want %v", cfg.Type, cache.TypeShort)
	}

	if cfg.TTL != 60 {
		t.Errorf("TTL = %v, want 60", cfg.TTL)
	}

	if cfg.MaxSize != 1000 {
		t.Errorf("MaxSize = %v, want 1000", cfg.MaxSize)
	}

	if cfg.Directory != ".mcp-cache" {
		t.Errorf("Directory = %v, want .mcp-cache", cfg.Directory)
	}

	if cfg.ToolTTL == nil {
		t.Error("ToolTTL should be initialized")
	}

	if len(cfg.ToolTTL) != 0 {
		t.Errorf("ToolTTL should be empty initially, got %d entries", len(cfg.ToolTTL))
	}
}

// Test: Validate
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *cache.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid short cache",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     60,
				MaxSize: 1000,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "valid long cache",
			config: &cache.Config{
				Type:      cache.TypeLong,
				TTL:       30,
				Directory: "/tmp/cache",
				Enabled:   true,
			},
			wantErr: false,
		},
		{
			name: "disabled cache - no validation (safe default)",
			config: &cache.Config{
				Type:    "invalid",
				TTL:     -1,
				Enabled: false, // ⚠️ Disabled = no validation
			},
			wantErr: false, // Should NOT error because disabled
		},
		{
			name: "invalid type",
			config: &cache.Config{
				Type:    "invalid",
				TTL:     60,
				Enabled: true,
			},
			wantErr: true,
			errMsg:  "invalid cache type",
		},
		{
			name: "negative TTL",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     -1,
				MaxSize: 1000,
				Enabled: true,
			},
			wantErr: true,
			errMsg:  "TTL must be positive",
		},
		{
			name: "zero TTL",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     0,
				MaxSize: 1000,
				Enabled: true,
			},
			wantErr: true,
			errMsg:  "TTL must be positive",
		},
		{
			name: "negative max size for short cache",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     60,
				MaxSize: -1,
				Enabled: true,
			},
			wantErr: true,
			errMsg:  "max_size must be positive",
		},
		{
			name: "zero max size for short cache",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     60,
				MaxSize: 0,
				Enabled: true,
			},
			wantErr: true,
			errMsg:  "max_size must be positive",
		},
		{
			name: "missing directory for long cache",
			config: &cache.Config{
				Type:      cache.TypeLong,
				TTL:       30,
				Directory: "",
				Enabled:   true,
			},
			wantErr: true,
			errMsg:  "directory is required",
		},
		{
			name: "very large TTL - should be valid",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     86400, // 1 day in seconds
				MaxSize: 1000,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "very small TTL - should be valid",
			config: &cache.Config{
				Type:    cache.TypeShort,
				TTL:     1, // 1 second
				MaxSize: 1000,
				Enabled: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Errorf("Validate() error = nil, want error containing %q", tt.errMsg)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

// Test: GetTTLDuration
func TestConfig_GetTTLDuration(t *testing.T) {
	tests := []struct {
		name string
		cfg  *cache.Config
		want time.Duration
	}{
		{
			name: "short cache - seconds",
			cfg: &cache.Config{
				Type: cache.TypeShort,
				TTL:  60,
			},
			want: 60 * time.Second,
		},
		{
			name: "long cache - minutes",
			cfg: &cache.Config{
				Type: cache.TypeLong,
				TTL:  30,
			},
			want: 30 * time.Minute,
		},
		{
			name: "short cache - 1 second",
			cfg: &cache.Config{
				Type: cache.TypeShort,
				TTL:  1,
			},
			want: 1 * time.Second,
		},
		{
			name: "long cache - 1 minute",
			cfg: &cache.Config{
				Type: cache.TypeLong,
				TTL:  1,
			},
			want: 1 * time.Minute,
		},
		{
			name: "short cache - 1 hour in seconds",
			cfg: &cache.Config{
				Type: cache.TypeShort,
				TTL:  3600,
			},
			want: 3600 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetTTLDuration()
			if got != tt.want {
				t.Errorf("GetTTLDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test: GetToolTTL
func TestConfig_GetToolTTL(t *testing.T) {
	cfg := &cache.Config{
		Type: cache.TypeShort,
		TTL:  60,
		ToolTTL: map[string]time.Duration{
			"search":     30 * time.Second,
			"list_files": 5 * time.Minute,
		},
	}

	tests := []struct {
		name     string
		toolName string
		want     time.Duration
	}{
		{
			name:     "tool with override - search",
			toolName: "search",
			want:     30 * time.Second,
		},
		{
			name:     "tool with override - list_files",
			toolName: "list_files",
			want:     5 * time.Minute,
		},
		{
			name:     "tool without override - uses default",
			toolName: "read_file",
			want:     60 * time.Second, // Default from config
		},
		{
			name:     "non-existent tool - uses default",
			toolName: "unknown_tool",
			want:     60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetToolTTL(tt.toolName)
			if got != tt.want {
				t.Errorf("GetToolTTL(%s) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

// Test: SetToolTTL
func TestConfig_SetToolTTL(t *testing.T) {
	t.Run("set TTL on empty map", func(t *testing.T) {
		cfg := cache.DefaultConfig()

		toolName := "search"
		ttl := 45 * time.Second

		cfg.SetToolTTL(toolName, ttl)

		got := cfg.GetToolTTL(toolName)
		if got != ttl {
			t.Errorf("After SetToolTTL, GetToolTTL(%s) = %v, want %v", toolName, got, ttl)
		}
	})

	t.Run("set multiple TTLs", func(t *testing.T) {
		cfg := cache.DefaultConfig()

		cfg.SetToolTTL("tool1", 10*time.Second)
		cfg.SetToolTTL("tool2", 20*time.Second)
		cfg.SetToolTTL("tool3", 30*time.Second)

		if cfg.GetToolTTL("tool1") != 10*time.Second {
			t.Error("tool1 TTL not set correctly")
		}
		if cfg.GetToolTTL("tool2") != 20*time.Second {
			t.Error("tool2 TTL not set correctly")
		}
		if cfg.GetToolTTL("tool3") != 30*time.Second {
			t.Error("tool3 TTL not set correctly")
		}
	})

	t.Run("override existing TTL", func(t *testing.T) {
		cfg := cache.DefaultConfig()

		cfg.SetToolTTL("tool", 10*time.Second)
		cfg.SetToolTTL("tool", 20*time.Second) // Override

		got := cfg.GetToolTTL("tool")
		if got != 20*time.Second {
			t.Errorf("GetToolTTL(tool) = %v, want 20s (should be overridden)", got)
		}
	})
}

// Test: Enable/Disable
func TestConfig_EnableDisable(t *testing.T) {
	cfg := cache.DefaultConfig()

	// Verify disabled by default
	if cfg.IsEnabled() {
		t.Error("cache should be disabled by default")
	}

	// Enable
	cfg.Enable()
	if !cfg.IsEnabled() {
		t.Error("cache should be enabled after Enable()")
	}

	// Disable
	cfg.Disable()
	if cfg.IsEnabled() {
		t.Error("cache should be disabled after Disable()")
	}
}

// Test: Type constants
func TestType_Constants(t *testing.T) {
	if cache.TypeShort != "short" {
		t.Errorf("TypeShort = %v, want 'short'", cache.TypeShort)
	}

	if cache.TypeLong != "long" {
		t.Errorf("TypeLong = %v, want 'long'", cache.TypeLong)
	}
}

// Test: Config edge cases
func TestConfig_EdgeCases(t *testing.T) {
	t.Run("nil ToolTTL map - SetToolTTL initializes it", func(t *testing.T) {
		cfg := &cache.Config{
			ToolTTL: nil, // Explicitly nil
		}

		cfg.SetToolTTL("tool", 10*time.Second)

		if cfg.ToolTTL == nil {
			t.Error("SetToolTTL should initialize nil ToolTTL map")
		}

		if cfg.GetToolTTL("tool") != 10*time.Second {
			t.Error("SetToolTTL should work even with initially nil map")
		}
	})

	t.Run("empty string tool name", func(t *testing.T) {
		cfg := cache.DefaultConfig()

		cfg.SetToolTTL("", 10*time.Second)
		got := cfg.GetToolTTL("")

		if got != 10*time.Second {
			t.Error("should handle empty string tool name")
		}
	})
}

// Benchmark: GetToolTTL
func BenchmarkConfig_GetToolTTL(b *testing.B) {
	cfg := &cache.Config{
		Type: cache.TypeShort,
		TTL:  60,
		ToolTTL: map[string]time.Duration{
			"search": 30 * time.Second,
		},
	}

	b.Run("with override", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetToolTTL("search")
		}
	})

	b.Run("without override", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg.GetToolTTL("unknown_tool")
		}
	})
}
