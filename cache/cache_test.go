package cache_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/cache"
)

// Test: IsExpired
func TestEntry_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired - future expiry",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "expired - past expiry",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "just expired - 1ms ago",
			expiresAt: time.Now().Add(-1 * time.Millisecond),
			want:      true,
		},
		{
			name:      "about to expire - 1ms future",
			expiresAt: time.Now().Add(1 * time.Millisecond),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &cache.Entry{
				ExpiresAt: tt.expiresAt,
			}

			got := entry.IsExpired()
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test: TTL
func TestEntry_TTL(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		wantZero  bool
	}{
		{
			name:      "future expiry - should have TTL",
			expiresAt: time.Now().Add(1 * time.Hour),
			wantZero:  false,
		},
		{
			name:      "past expiry - should be zero",
			expiresAt: time.Now().Add(-1 * time.Hour),
			wantZero:  true,
		},
		{
			name:      "just expired - should be zero",
			expiresAt: time.Now().Add(-1 * time.Millisecond),
			wantZero:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &cache.Entry{
				ExpiresAt: tt.expiresAt,
			}

			ttl := entry.TTL()

			if tt.wantZero {
				if ttl != 0 {
					t.Errorf("TTL() = %v, want 0 (expired)", ttl)
				}
			} else {
				if ttl <= 0 {
					t.Errorf("TTL() = %v, want > 0 (not expired)", ttl)
				}
			}
		})
	}
}

// Test: Unmarshal
func TestEntry_Unmarshal(t *testing.T) {
	t.Run("unmarshal simple struct", func(t *testing.T) {
		type TestData struct {
			Message string `json:"message"`
			Count   int    `json:"count"`
		}

		original := TestData{
			Message: "hello",
			Count:   42,
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal test data: %v", err)
		}

		// Create entry
		entry := &cache.Entry{
			Value: jsonData,
		}

		// Unmarshal
		var result TestData
		if err := entry.Unmarshal(&result); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		// Verify
		if result.Message != original.Message {
			t.Errorf("Message = %v, want %v", result.Message, original.Message)
		}
		if result.Count != original.Count {
			t.Errorf("Count = %v, want %v", result.Count, original.Count)
		}
	})

	t.Run("unmarshal map", func(t *testing.T) {
		original := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		}

		jsonData, _ := json.Marshal(original)
		entry := &cache.Entry{
			Value: jsonData,
		}

		var result map[string]interface{}
		if err := entry.Unmarshal(&result); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if result["key1"] != "value1" {
			t.Errorf("key1 = %v, want value1", result["key1"])
		}
		if result["key2"] != float64(123) { // JSON numbers are float64
			t.Errorf("key2 = %v, want 123", result["key2"])
		}
	})

	t.Run("unmarshal array", func(t *testing.T) {
		original := []string{"a", "b", "c"}

		jsonData, _ := json.Marshal(original)
		entry := &cache.Entry{
			Value: jsonData,
		}

		var result []string
		if err := entry.Unmarshal(&result); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
		if result[0] != "a" {
			t.Errorf("result[0] = %v, want a", result[0])
		}
	})

	t.Run("unmarshal invalid JSON - should fail", func(t *testing.T) {
		entry := &cache.Entry{
			Value: json.RawMessage(`{invalid json}`),
		}

		var result map[string]interface{}
		err := entry.Unmarshal(&result)
		if err == nil {
			t.Error("Unmarshal() should fail for invalid JSON")
		}
	})
}

// Test: Age
func TestEntry_Age(t *testing.T) {
	t.Run("entry created 5 minutes ago", func(t *testing.T) {
		createdAt := time.Now().Add(-5 * time.Minute)

		entry := &cache.Entry{
			CreatedAt: createdAt,
		}

		age := entry.Age()

		// Age should be approximately 5 minutes (allow 1 second drift)
		expected := 5 * time.Minute
		if age < expected-time.Second || age > expected+time.Second {
			t.Errorf("Age() = %v, want ~%v", age, expected)
		}
	})

	t.Run("entry created just now", func(t *testing.T) {
		entry := &cache.Entry{
			CreatedAt: time.Now(),
		}

		age := entry.Age()

		// Age should be very small (< 100ms)
		if age > 100*time.Millisecond {
			t.Errorf("Age() = %v, want < 100ms", age)
		}
	})

	t.Run("entry created 1 hour ago", func(t *testing.T) {
		createdAt := time.Now().Add(-1 * time.Hour)

		entry := &cache.Entry{
			CreatedAt: createdAt,
		}

		age := entry.Age()

		// Age should be approximately 1 hour
		expected := 1 * time.Hour
		if age < expected-time.Second || age > expected+time.Second {
			t.Errorf("Age() = %v, want ~%v", age, expected)
		}
	})
}

// Test: Entry complete lifecycle
func TestEntry_Lifecycle(t *testing.T) {
	// Create entry
	data := map[string]interface{}{
		"result": "success",
		"count":  10,
	}
	jsonData, _ := json.Marshal(data)

	entry := &cache.Entry{
		Key:       "test-key-123",
		Value:     jsonData,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Hits:      0,
	}

	// Verify not expired
	if entry.IsExpired() {
		t.Error("entry should not be expired")
	}

	// Verify TTL
	ttl := entry.TTL()
	if ttl <= 0 || ttl > time.Hour {
		t.Errorf("TTL = %v, want ~1 hour", ttl)
	}

	// Verify Age
	age := entry.Age()
	if age > 100*time.Millisecond {
		t.Errorf("Age = %v, want < 100ms", age)
	}

	// Simulate hits
	entry.Hits = 5
	if entry.Hits != 5 {
		t.Errorf("Hits = %d, want 5", entry.Hits)
	}

	// Unmarshal
	var result map[string]interface{}
	if err := entry.Unmarshal(&result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if result["result"] != "success" {
		t.Errorf("result = %v, want success", result["result"])
	}
}

// Benchmark: Unmarshal
func BenchmarkEntry_Unmarshal(b *testing.B) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}
	jsonData, _ := json.Marshal(data)

	entry := &cache.Entry{
		Value: jsonData,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		entry.Unmarshal(&result)
	}
}
