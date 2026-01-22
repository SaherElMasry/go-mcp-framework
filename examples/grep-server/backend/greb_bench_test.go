package backend

import (
	"context"
	"testing"
)

// MockEmitter implements the backend.StreamingEmitter interface correctly
type MockEmitter struct {
	ctx context.Context
}

// EmitData matches the interface
func (m *MockEmitter) EmitData(data interface{}) error {
	return nil
}

// EmitProgress NOW returns an error to satisfy the interface
func (m *MockEmitter) EmitProgress(current, total int64, message string) error {
	return nil
}

// Context returns the context
func (m *MockEmitter) Context() context.Context {
	return m.ctx
}

func BenchmarkHandleSearchCSV(b *testing.B) {
	gb := NewGrepBackend()
	ctx := context.Background()
	emitter := &MockEmitter{ctx: ctx}

	// Make sure this file actually exists in the path relative to this test file
	args := map[string]interface{}{
		"file_path":    "../demo-data/info-records.csv",
		"search_type":  "department",
		"search_value": "Engineering",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := gb.handleSearchCSV(ctx, args, emitter)
		if err != nil {
			b.Fatalf("Search failed at iteration %d: %v", i, err)
		}
	}
}
