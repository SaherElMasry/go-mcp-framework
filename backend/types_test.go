package backend_test

import (
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// Test: IsCacheable
func TestToolDefinition_IsCacheable(t *testing.T) {
	tests := []struct {
		name string
		tool backend.ToolDefinition
		want bool
	}{
		{
			name: "cacheable tool",
			tool: backend.ToolDefinition{
				Name: "read_file",
				Cache: backend.ToolCacheConfig{
					Cacheable: true,
				},
			},
			want: true,
		},
		{
			name: "non-cacheable tool",
			tool: backend.ToolDefinition{
				Name: "create_file",
				Cache: backend.ToolCacheConfig{
					Cacheable: false,
				},
			},
			want: false,
		},
		{
			name: "default tool (no cache config)",
			tool: backend.ToolDefinition{
				Name:  "tool",
				Cache: backend.ToolCacheConfig{}, // Default: false
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.IsCacheable()
			if got != tt.want {
				t.Errorf("IsCacheable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test: GetCacheTTL
func TestToolDefinition_GetCacheTTL(t *testing.T) {
	fiveMin := 5 * time.Minute
	tenMin := 10 * time.Minute

	tests := []struct {
		name       string
		tool       backend.ToolDefinition
		defaultTTL time.Duration
		want       time.Duration
	}{
		{
			name: "tool with custom TTL",
			tool: backend.ToolDefinition{
				Name: "search",
				Cache: backend.ToolCacheConfig{
					Cacheable: true,
					TTL:       &fiveMin,
				},
			},
			defaultTTL: tenMin,
			want:       fiveMin, // Use tool TTL
		},
		{
			name: "tool without custom TTL",
			tool: backend.ToolDefinition{
				Name: "list_files",
				Cache: backend.ToolCacheConfig{
					Cacheable: true,
					TTL:       nil,
				},
			},
			defaultTTL: tenMin,
			want:       tenMin, // Use default
		},
		{
			name: "non-cacheable tool",
			tool: backend.ToolDefinition{
				Name: "create_file",
				Cache: backend.ToolCacheConfig{
					Cacheable: false,
				},
			},
			defaultTTL: tenMin,
			want:       tenMin, // Falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.GetCacheTTL(tt.defaultTTL)
			if got != tt.want {
				t.Errorf("GetCacheTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test: HasCacheTags
func TestToolDefinition_HasCacheTags(t *testing.T) {
	tool := backend.ToolDefinition{
		Name: "search",
		Cache: backend.ToolCacheConfig{
			Cacheable: true,
			Tags:      []string{"read-only", "fast", "public"},
		},
	}

	tests := []struct {
		name string
		tags []string
		want bool
	}{
		{
			name: "has single tag",
			tags: []string{"read-only"},
			want: true,
		},
		{
			name: "has multiple tags",
			tags: []string{"read-only", "fast"},
			want: true,
		},
		{
			name: "missing tag",
			tags: []string{"read-only", "slow"},
			want: false,
		},
		{
			name: "no tags specified",
			tags: []string{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tool.HasCacheTags(tt.tags...)
			if got != tt.want {
				t.Errorf("HasCacheTags(%v) = %v, want %v", tt.tags, got, tt.want)
			}
		})
	}
}

// Test: Helper Functions
func TestMakeCacheable(t *testing.T) {
	ttl := 5 * time.Minute
	config := backend.MakeCacheable(ttl)

	if !config.Cacheable {
		t.Error("MakeCacheable() should set Cacheable to true")
	}

	if config.TTL == nil {
		t.Error("MakeCacheable() should set TTL")
	}

	if *config.TTL != ttl {
		t.Errorf("TTL = %v, want %v", *config.TTL, ttl)
	}
}

func TestMakeNonCacheable(t *testing.T) {
	config := backend.MakeNonCacheable()

	if config.Cacheable {
		t.Error("MakeNonCacheable() should set Cacheable to false")
	}

	if config.TTL != nil {
		t.Error("MakeNonCacheable() should not set TTL")
	}
}

func TestDefaultCacheConfig(t *testing.T) {
	config := backend.DefaultCacheConfig()

	if config.Cacheable {
		t.Error("DefaultCacheConfig() should be non-cacheable (safe default)")
	}
}

// Test: ToolBuilder with Cache
func TestToolBuilder_WithCache(t *testing.T) {
	ttl := 5 * time.Minute
	tool := backend.NewTool("read_file").
		Description("Reads a file").
		StringParam("path", "File path", true).
		WithCache(true, ttl).
		Build()

	if !tool.IsCacheable() {
		t.Error("tool should be cacheable")
	}

	if tool.GetCacheTTL(0) != ttl {
		t.Errorf("TTL = %v, want %v", tool.GetCacheTTL(0), ttl)
	}
}

func TestToolBuilder_Cacheable(t *testing.T) {
	tool := backend.NewTool("list_files").
		Cacheable().
		Build()

	if !tool.IsCacheable() {
		t.Error("tool should be cacheable")
	}
}

func TestToolBuilder_NonCacheable(t *testing.T) {
	tool := backend.NewTool("create_file").
		NonCacheable().
		Build()

	if tool.IsCacheable() {
		t.Error("tool should not be cacheable")
	}
}

func TestToolBuilder_CacheTags(t *testing.T) {
	tool := backend.NewTool("search").
		WithCache(true, time.Minute).
		CacheTags("read-only", "fast").
		Build()

	if !tool.HasCacheTags("read-only") {
		t.Error("tool should have 'read-only' tag")
	}

	if !tool.HasCacheTags("read-only", "fast") {
		t.Error("tool should have both tags")
	}
}

// Test: Default Cache Config on NewTool
func TestToolBuilder_DefaultNonCacheable(t *testing.T) {
	tool := backend.NewTool("some_tool").Build()

	if tool.IsCacheable() {
		t.Error("tools should be non-cacheable by default (safe default)")
	}
}

// Test: Real-world Scenarios
func TestToolDefinition_RealWorldScenarios(t *testing.T) {
	t.Run("file system read tool - cacheable", func(t *testing.T) {
		tool := backend.NewTool("read_file").
			Description("Reads file content").
			StringParam("path", "File path", true).
			WithCache(true, 5*time.Minute).
			Build()

		if !tool.IsCacheable() {
			t.Error("read_file should be cacheable")
		}
	})

	t.Run("file system write tool - non-cacheable", func(t *testing.T) {
		tool := backend.NewTool("create_file").
			Description("Creates a new file").
			StringParam("path", "File path", true).
			StringParam("content", "File content", true).
			NonCacheable().
			Build()

		if tool.IsCacheable() {
			t.Error("create_file should NOT be cacheable")
		}
	})

	t.Run("search tool - cacheable with short TTL", func(t *testing.T) {
		tool := backend.NewTool("search").
			Description("Searches for items").
			StringParam("query", "Search query", true).
			WithCache(true, 30*time.Second).
			CacheTags("read-only", "volatile").
			Build()

		if !tool.IsCacheable() {
			t.Error("search should be cacheable")
		}

		if tool.GetCacheTTL(0) != 30*time.Second {
			t.Error("search should have 30s TTL")
		}
	})
}
