package cache_test

import (
	"fmt"
	"testing"

	"github.com/SaherElMasry/go-mcp-framework/cache"
)

// Test: Basic Generation
func TestKeyGenerator_Generate(t *testing.T) {
	kg := cache.NewKeyGenerator()

	t.Run("basic generation", func(t *testing.T) {
		args := map[string]interface{}{
			"path": "/tmp/test.txt",
		}

		key, err := kg.Generate("read_file", args)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if key == "" {
			t.Error("Generate() returned empty key")
		}

		// Key should be SHA-256 hex (64 characters)
		if len(key) != 64 {
			t.Errorf("key length = %d, want 64 (SHA-256 hex)", len(key))
		}
	})

	t.Run("different tools produce different keys", func(t *testing.T) {
		args := map[string]interface{}{
			"path": "/tmp/test.txt",
		}

		key1, _ := kg.Generate("read_file", args)
		key2, _ := kg.Generate("write_file", args)

		if key1 == key2 {
			t.Error("different tools should produce different keys")
		}
	})

	t.Run("different args produce different keys", func(t *testing.T) {
		args1 := map[string]interface{}{
			"path": "/tmp/test1.txt",
		}

		args2 := map[string]interface{}{
			"path": "/tmp/test2.txt",
		}

		key1, _ := kg.Generate("read_file", args1)
		key2, _ := kg.Generate("read_file", args2)

		if key1 == key2 {
			t.Error("different args should produce different keys")
		}
	})

	t.Run("empty args", func(t *testing.T) {
		args := map[string]interface{}{}

		key, err := kg.Generate("tool", args)
		if err != nil {
			t.Fatalf("Generate() with empty args error = %v", err)
		}

		if len(key) != 64 {
			t.Errorf("key length = %d, want 64", len(key))
		}
	})

	t.Run("nil args - should handle gracefully", func(t *testing.T) {
		key, err := kg.Generate("tool", nil)
		if err != nil {
			t.Fatalf("Generate() with nil args error = %v", err)
		}

		if len(key) != 64 {
			t.Errorf("key length = %d, want 64", len(key))
		}
	})
}

// ⚠️ CRITICAL TEST: Deterministic Key Generation
// This is THE most important test for cache correctness!
func TestKeyGenerator_Deterministic(t *testing.T) {
	kg := cache.NewKeyGenerator()

	t.Run("same args in different order produce same key", func(t *testing.T) {
		// This is THE critical test - map order should not matter
		args1 := map[string]interface{}{
			"b": 2,
			"a": 1,
			"c": 3,
		}

		args2 := map[string]interface{}{
			"c": 3,
			"a": 1,
			"b": 2,
		}

		args3 := map[string]interface{}{
			"a": 1,
			"b": 2,
			"c": 3,
		}

		key1, err1 := kg.Generate("tool", args1)
		key2, err2 := kg.Generate("tool", args2)
		key3, err3 := kg.Generate("tool", args3)

		if err1 != nil || err2 != nil || err3 != nil {
			t.Fatalf("Generate() errors: %v, %v, %v", err1, err2, err3)
		}

		if key1 != key2 {
			t.Errorf("keys should match for same logical args\nkey1: %s\nkey2: %s", key1, key2)
		}

		if key1 != key3 {
			t.Errorf("keys should match for same logical args\nkey1: %s\nkey3: %s", key1, key3)
		}
	})

	t.Run("nested maps are normalized", func(t *testing.T) {
		args1 := map[string]interface{}{
			"filter": map[string]interface{}{
				"z": 3,
				"a": 1,
			},
		}

		args2 := map[string]interface{}{
			"filter": map[string]interface{}{
				"a": 1,
				"z": 3,
			},
		}

		key1, _ := kg.Generate("tool", args1)
		key2, _ := kg.Generate("tool", args2)

		if key1 != key2 {
			t.Error("nested maps should produce same key regardless of order")
		}
	})

	t.Run("deeply nested maps are normalized", func(t *testing.T) {
		args1 := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"z": 3,
						"a": 1,
					},
				},
			},
		}

		args2 := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"a": 1,
						"z": 3,
					},
				},
			},
		}

		key1, _ := kg.Generate("tool", args1)
		key2, _ := kg.Generate("tool", args2)

		if key1 != key2 {
			t.Error("deeply nested maps should produce same key")
		}
	})

	t.Run("arrays maintain order - order matters!", func(t *testing.T) {
		args1 := map[string]interface{}{
			"items": []interface{}{1, 2, 3},
		}

		args2 := map[string]interface{}{
			"items": []interface{}{3, 2, 1},
		}

		key1, _ := kg.Generate("tool", args1)
		key2, _ := kg.Generate("tool", args2)

		// Arrays should NOT be reordered - order matters in arrays!
		if key1 == key2 {
			t.Error("different array orders should produce different keys")
		}
	})

	t.Run("same array order produces same key", func(t *testing.T) {
		args1 := map[string]interface{}{
			"items": []interface{}{1, 2, 3},
		}

		args2 := map[string]interface{}{
			"items": []interface{}{1, 2, 3},
		}

		key1, _ := kg.Generate("tool", args1)
		key2, _ := kg.Generate("tool", args2)

		if key1 != key2 {
			t.Error("same array order should produce same key")
		}
	})

	t.Run("repeated generation produces same key", func(t *testing.T) {
		args := map[string]interface{}{
			"path": "/tmp/test.txt",
			"mode": "read",
		}

		// Generate key 10 times
		keys := make([]string, 10)
		for i := 0; i < 10; i++ {
			key, err := kg.Generate("read_file", args)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			keys[i] = key
		}

		// All keys should be identical
		for i := 1; i < len(keys); i++ {
			if keys[i] != keys[0] {
				t.Errorf("key %d (%s) differs from key 0 (%s)", i, keys[i], keys[0])
			}
		}
	})

	t.Run("map with arrays of maps", func(t *testing.T) {
		args1 := map[string]interface{}{
			"filters": []interface{}{
				map[string]interface{}{"z": 1, "a": 2},
				map[string]interface{}{"b": 3, "c": 4},
			},
		}

		args2 := map[string]interface{}{
			"filters": []interface{}{
				map[string]interface{}{"a": 2, "z": 1}, // Different order in map
				map[string]interface{}{"c": 4, "b": 3}, // Different order in map
			},
		}

		key1, _ := kg.Generate("tool", args1)
		key2, _ := kg.Generate("tool", args2)

		// Maps inside arrays should be normalized
		if key1 != key2 {
			t.Error("maps inside arrays should be normalized")
		}
	})
}

// Test: Complex Types
func TestKeyGenerator_ComplexTypes(t *testing.T) {
	kg := cache.NewKeyGenerator()

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "nested objects",
			args: map[string]interface{}{
				"config": map[string]interface{}{
					"nested": map[string]interface{}{
						"deep": map[string]interface{}{
							"value": 42,
						},
					},
				},
			},
		},
		{
			name: "mixed types",
			args: map[string]interface{}{
				"string": "hello",
				"number": 42,
				"float":  3.14,
				"bool":   true,
				"null":   nil,
			},
		},
		{
			name: "arrays of maps",
			args: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": 1, "name": "first"},
					map[string]interface{}{"id": 2, "name": "second"},
				},
			},
		},
		{
			name: "empty values",
			args: map[string]interface{}{
				"empty_string": "",
				"empty_array":  []interface{}{},
				"empty_map":    map[string]interface{}{},
			},
		},
		{
			name: "unicode strings",
			args: map[string]interface{}{
				"emoji":   "🚀",
				"chinese": "你好",
				"arabic":  "مرحبا",
			},
		},
		{
			name: "special characters",
			args: map[string]interface{}{
				"path":  "/tmp/test\nfile.txt",
				"quote": `He said "hello"`,
				"slash": "back\\slash",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := kg.Generate("tool", tt.args)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
			}

			if key == "" {
				t.Error("Generate() returned empty key")
			}

			if len(key) != 64 {
				t.Errorf("key length = %d, want 64", len(key))
			}

			// Verify determinism - generate again
			key2, _ := kg.Generate("tool", tt.args)
			if key != key2 {
				t.Error("repeated generation should produce same key")
			}
		})
	}
}

// Test: Edge Cases
func TestKeyGenerator_EdgeCases(t *testing.T) {
	kg := cache.NewKeyGenerator()

	t.Run("very large args", func(t *testing.T) {
		// Create large args
		args := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			args[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}

		key, err := kg.Generate("tool", args)
		if err != nil {
			t.Fatalf("Generate() with large args error = %v", err)
		}

		if len(key) != 64 {
			t.Errorf("key length = %d, want 64", len(key))
		}
	})

	t.Run("numeric keys as strings", func(t *testing.T) {
		args := map[string]interface{}{
			"1": "one",
			"2": "two",
			"3": "three",
		}

		key, err := kg.Generate("tool", args)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(key) != 64 {
			t.Errorf("key length = %d, want 64", len(key))
		}
	})

	t.Run("negative numbers", func(t *testing.T) {
		args := map[string]interface{}{
			"negative": -42,
			"positive": 42,
		}

		key1, _ := kg.Generate("tool", args)

		args2 := map[string]interface{}{
			"positive": 42,
			"negative": -42,
		}

		key2, _ := kg.Generate("tool", args2)

		if key1 != key2 {
			t.Error("keys should match regardless of order")
		}
	})

	t.Run("floating point numbers", func(t *testing.T) {
		args := map[string]interface{}{
			"pi": 3.14159,
		}

		key, err := kg.Generate("tool", args)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(key) != 64 {
			t.Errorf("key length = %d, want 64", len(key))
		}
	})
}

// Test: GenerateSimple
func TestKeyGenerator_GenerateSimple(t *testing.T) {
	kg := cache.NewKeyGenerator()

	tests := []struct {
		name     string
		toolName string
		want     string
	}{
		{
			name:     "simple tool name",
			toolName: "test_tool",
			want:     "tool:test_tool",
		},
		{
			name:     "tool with underscores",
			toolName: "read_file",
			want:     "tool:read_file",
		},
		{
			name:     "empty tool name",
			toolName: "",
			want:     "tool:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kg.GenerateSimple(tt.toolName)
			if got != tt.want {
				t.Errorf("GenerateSimple() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test: Real-world scenarios
func TestKeyGenerator_RealWorldScenarios(t *testing.T) {
	kg := cache.NewKeyGenerator()

	t.Run("file system tool - read_file", func(t *testing.T) {
		args1 := map[string]interface{}{
			"path": "/home/user/document.txt",
		}

		args2 := map[string]interface{}{
			"path": "/home/user/document.txt",
		}

		key1, _ := kg.Generate("read_file", args1)
		key2, _ := kg.Generate("read_file", args2)

		if key1 != key2 {
			t.Error("same file path should produce same key")
		}
	})

	t.Run("search tool with complex filters", func(t *testing.T) {
		args1 := map[string]interface{}{
			"query": "golang cache",
			"filters": map[string]interface{}{
				"language": "go",
				"stars":    ">100",
			},
			"sort": "relevance",
		}

		args2 := map[string]interface{}{
			"sort":  "relevance",
			"query": "golang cache",
			"filters": map[string]interface{}{
				"stars":    ">100",
				"language": "go",
			},
		}

		key1, _ := kg.Generate("search", args1)
		key2, _ := kg.Generate("search", args2)

		if key1 != key2 {
			t.Error("search with same logical filters should produce same key")
		}
	})

	t.Run("pagination - different pages different keys", func(t *testing.T) {
		args1 := map[string]interface{}{
			"page": 1,
		}

		args2 := map[string]interface{}{
			"page": 2,
		}

		key1, _ := kg.Generate("list", args1)
		key2, _ := kg.Generate("list", args2)

		if key1 == key2 {
			t.Error("different pages should produce different keys")
		}
	})
}

// Benchmarks
func BenchmarkKeyGenerator_Generate(b *testing.B) {
	kg := cache.NewKeyGenerator()

	args := map[string]interface{}{
		"path":   "/tmp/test.txt",
		"mode":   "read",
		"offset": 0,
		"limit":  100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kg.Generate("read_file", args)
	}
}

func BenchmarkKeyGenerator_GenerateComplex(b *testing.B) {
	kg := cache.NewKeyGenerator()

	args := map[string]interface{}{
		"filter": map[string]interface{}{
			"status":   "active",
			"priority": "high",
			"tags":     []interface{}{"urgent", "bug", "customer"},
		},
		"sort": map[string]interface{}{
			"by":    "created_at",
			"order": "desc",
		},
		"pagination": map[string]interface{}{
			"page":     1,
			"per_page": 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kg.Generate("search", args)
	}
}

func BenchmarkKeyGenerator_GenerateLargeArgs(b *testing.B) {
	kg := cache.NewKeyGenerator()

	// Create large args with 100 keys
	args := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		args[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kg.Generate("tool", args)
	}
}
