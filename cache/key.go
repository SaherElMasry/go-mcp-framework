package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// KeyGenerator generates deterministic cache keys
// CRITICAL: Keys must be deterministic - same logical input = same key
// This prevents cache misses due to Go's random map iteration order
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// Generate generates a cache key from tool name and arguments
// Uses SHA-256 hash of canonicalized JSON
//
// CRITICAL FIX: Arguments are normalized before hashing to ensure
// deterministic keys regardless of map iteration order
//
// Example:
//
//	args1 := {"b": 2, "a": 1}  // Random order
//	args2 := {"a": 1, "b": 2}  // Different order
//	Generate("tool", args1) == Generate("tool", args2)  // ✅ Same key!
func (kg *KeyGenerator) Generate(toolName string, args map[string]interface{}) (string, error) {
	// CRITICAL: Normalize arguments for deterministic hashing
	normalized := kg.normalize(args)

	// Create a deterministic representation
	data := struct {
		Tool string      `json:"tool"`
		Args interface{} `json:"args"`
	}{
		Tool: toolName,
		Args: normalized,
	}

	// Serialize to JSON (now deterministic due to normalization)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cache key: %w", err)
	}

	// Hash the JSON using SHA-256
	hash := sha256.Sum256(jsonData)
	key := hex.EncodeToString(hash[:])

	return key, nil
}

// normalize ensures deterministic JSON serialization by sorting map keys
// This is CRITICAL for cache correctness
//
// Problem: Go maps have random iteration order
//
//	{"b": 2, "a": 1} might serialize as {"b":2,"a":1} or {"a":1,"b":2}
//	This causes different cache keys for the same logical input!
//
// Solution: Sort keys before serialization
//
//	{"b": 2, "a": 1} → {"a": 1, "b": 2} → always {"a":1,"b":2}
//
// This function recursively normalizes:
// - Maps: Sort keys alphabetically
// - Arrays: Keep order (order matters in arrays)
// - Primitives: Return as-is
func (kg *KeyGenerator) normalize(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Sort map keys for deterministic order
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Build normalized map with sorted keys
		normalized := make(map[string]interface{}, len(val))
		for _, k := range keys {
			normalized[k] = kg.normalize(val[k]) // Recursively normalize values
		}
		return normalized

	case []interface{}:
		// Recursively normalize array elements
		// NOTE: Array order is preserved (order matters!)
		normalized := make([]interface{}, len(val))
		for i, item := range val {
			normalized[i] = kg.normalize(item)
		}
		return normalized

	default:
		// Primitive types (string, int, float, bool, nil) are already deterministic
		return v
	}
}

// GenerateSimple generates a simple cache key without arguments
// Useful for testing or tools with no parameters
func (kg *KeyGenerator) GenerateSimple(toolName string) string {
	return fmt.Sprintf("tool:%s", toolName)
}
