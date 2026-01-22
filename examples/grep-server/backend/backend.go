package backend

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// GrepBackend implements grep-like search tools
type GrepBackend struct {
	*backend.BaseBackend
	baseDir string // Base directory for file operations
}

// NewGrepBackend creates a new grep backend
func NewGrepBackend() *GrepBackend {
	base := backend.NewBaseBackend("grep")

	// Get current working directory
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}

	gb := &GrepBackend{
		BaseBackend: base,
		baseDir:     baseDir,
	}

	// Register streaming tools
	gb.registerTools()

	return gb
}

func (gb *GrepBackend) registerTools() {
	// Tool 1: Grep HTML
	htmlTool := backend.NewTool("grep_html").
		Description("Search HTML file for patterns (e.g., href= to find URLs). Streams matches line-by-line.").
		StringParam("file_path", "Path to HTML file (relative or absolute)", true).
		StringParam("pattern", "Pattern to search for (default: href=)", false).
		Streaming(true).
		Build()

	gb.RegisterStreamingTool(htmlTool, gb.handleGrepHTML)

	// Tool 2: Search CSV
	csvTool := backend.NewTool("search_csv").
		Description("Search CSV file by field value. Returns matching records with details.").
		StringParam("file_path", "Path to CSV file (relative or absolute)", true).
		StringParam("search_type", "Field to search: name, email, age, salary, department", true).
		StringParam("search_value", "Value to search for (supports >, < for numbers)", true).
		Streaming(true).
		Build()

	gb.RegisterStreamingTool(csvTool, gb.handleSearchCSV)
}

// resolvePath resolves a file path (handles both relative and absolute)
func (gb *GrepBackend) resolvePath(path string) string {
	// If absolute, use as-is
	if filepath.IsAbs(path) {
		return path
	}

	// Try relative to base directory
	fullPath := filepath.Join(gb.baseDir, path)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	}

	// Try as-is (maybe it's relative to current working directory)
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Return the original path and let caller handle the error
	return fullPath
}

// handleGrepHTML searches HTML files for patterns
// func (gb *GrepBackend) handleGrepHTML(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
// 	// Extract arguments
// 	filePath := args["file_path"].(string)
// 	pattern := "href="
// 	if p, ok := args["pattern"].(string); ok && p != "" {
// 		pattern = p
// 	}

// 	// Resolve path
// 	fullPath := gb.resolvePath(filePath)

// 	// Open file
// 	file, err := os.Open(fullPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to open file %s: %w (tried: %s)", filePath, err, fullPath)
// 	}
// 	defer file.Close()

// 	// Count total lines for progress
// 	totalLines := countLines(fullPath)

// 	// Scan file line by line
// 	scanner := bufio.NewScanner(file)
// 	lineNum := 0
// 	matchCount := 0

// 	for scanner.Scan() {
// 		// Check for cancellation
// 		select {
// 		case <-emit.Context().Done():
// 			return ctx.Err()
// 		default:
// 		}

// 		lineNum++
// 		line := scanner.Text()

// 		// Emit progress every 10 lines
// 		if lineNum%10 == 0 {
// 			emit.EmitProgress(
// 				int64(lineNum),
// 				int64(totalLines),
// 				fmt.Sprintf("Scanned %d/%d lines, found %d matches", lineNum, totalLines, matchCount),
// 			)
// 		}

// 		// Check if line contains pattern
// 		if strings.Contains(line, pattern) {
// 			matchCount++

// 			// Extract URL if searching for href
// 			url := extractURL(line, pattern)

// 			// Emit match
// 			emit.EmitData(map[string]interface{}{
// 				"line_number": lineNum,
// 				"line_text":   strings.TrimSpace(line),
// 				"url":         url,
// 				"match_count": matchCount,
// 			})
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return fmt.Errorf("error reading file: %w", err)
// 	}

// 	// Final progress
// 	emit.EmitProgress(
// 		int64(totalLines),
// 		int64(totalLines),
// 		fmt.Sprintf("Complete! Found %d matches in %d lines", matchCount, totalLines),
// 	)

//		return nil
//	}
//
// // === Fast optimizedGrepHTML ===//
func (gb *GrepBackend) handleGrepHTML(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
	filePath := args["file_path"].(string)
	pattern := strings.ToLower(args["pattern"].(string)) // Pre-lowercase for speed

	fullPath := gb.resolvePath(filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open HTML: %w", err)
	}
	defer file.Close()

	// OPTIMIZATION: Use a large 64KB buffer for scanning
	// This prevents "token too long" errors on minified HTML files
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Can expand up to 1MB per line

	lineNum := 0
	matches := 0
	start := time.Now()

	// Replace the loop with this "Zero-Copy" version
	patternBytes := []byte(strings.ToLower(pattern))

	for scanner.Scan() {
		lineNum++
		// Get the raw bytes without creating a string
		lineBytes := scanner.Bytes()

		// Search directly in the bytes
		if bytes.Contains(bytes.ToLower(lineBytes), patternBytes) {
			matches++
			emit.EmitData(map[string]interface{}{
				"line_number": lineNum,
				"content":     string(lineBytes), // Only convert to string when we find a match!
			})
		}
	}
	duration := time.Since(start)
	emit.EmitProgress(int64(lineNum), int64(lineNum),
		fmt.Sprintf("✅ HTML Scan complete! Checked %d lines in %v. Found %d matches.", lineNum, duration, matches))

	return scanner.Err()
}

// === FastOptimization ===//
func (gb *GrepBackend) handleSearchCSV(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
	start := time.Now() // Add this at the very top

	filePath := args["file_path"].(string)
	searchType := args["search_type"].(string)
	searchLower := strings.ToLower(args["search_value"].(string))

	fullPath := gb.resolvePath(filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	// OPTIMIZATION 1: Buffered Reading (64KB chunks)
	// This reduces the "Syscall" overhead you saw in pprof
	bufferedFile := bufio.NewReaderSize(file, 64*1024)

	reader := csv.NewReader(bufferedFile)

	// OPTIMIZATION 2: Reuse Record Slice
	// This stops Go from creating a new slice for every single line
	reader.ReuseRecord = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// We must COPY the header because ReuseRecord might interfere later
	headerCopy := make([]string, len(header))
	copy(headerCopy, header)

	columnIndex := findColumnIndex(headerCopy, searchType)
	if columnIndex == -1 {
		return fmt.Errorf("unknown search type: %s", searchType)
	}

	matchCount := 0
	currentRow := 0

	for {
		record, err := reader.Read()
		if err != nil {
			break // EOF
		}
		currentRow++

		// Optimized search check
		if matchesSearchOptimized(record[columnIndex], searchLower, searchType) {
			matchCount++

			// parseUserRecord creates a map, which is fine for matches
			user := parseUserRecord(headerCopy, record)

			emit.EmitData(map[string]interface{}{
				"record_number": currentRow,
				"user":          user,
			})
		}

		// Progress update (every 50 rows for less overhead)
		if currentRow%50 == 0 {
			emit.EmitProgress(int64(currentRow), 0, "Scanning...")
		}
	}

	// At the very end, before 'return nil':
	duration := time.Since(start)
	emit.EmitProgress(
		int64(currentRow),
		int64(currentRow),
		fmt.Sprintf("✅ Search complete! Scanned %d rows in %v. Found %d matches.", currentRow, duration, matchCount),
	)

	return nil
}

// matchesSearchOptimized avoids memory allocations seen in pprof
func matchesSearchOptimized(value, searchLower, searchType string) bool {
	switch searchType {
	case "name", "email", "department", "status":
		// strings.Contains(strings.ToLower) is still needed for substring search,
		// but searchLower is now pre-computed outside the loop.
		return strings.Contains(strings.ToLower(value), searchLower)

	case "age", "salary":
		// Direct numeric logic to avoid extra function calls
		valNum, err := strconv.Atoi(value)
		if err != nil {
			return false
		}

		if strings.HasPrefix(searchLower, ">") {
			threshold, _ := strconv.Atoi(strings.TrimPrefix(searchLower, ">"))
			return valNum > threshold
		}
		if strings.HasPrefix(searchLower, "<") {
			threshold, _ := strconv.Atoi(strings.TrimPrefix(searchLower, "<"))
			return valNum < threshold
		}

		searchNum, err := strconv.Atoi(searchLower)
		return err == nil && valNum == searchNum

	default:
		return strings.EqualFold(value, searchLower)
	}
}

// /////////////////////////////////////////////

// Helper functions

func countLines(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}

func extractURL(line, pattern string) string {
	// Simple URL extraction for href=
	if !strings.Contains(pattern, "href") {
		return ""
	}

	// Find href="..."
	start := strings.Index(line, pattern)
	if start == -1 {
		return ""
	}

	start += len(pattern)

	// Find quote after href=
	if start >= len(line) {
		return ""
	}

	quote := line[start]
	if quote != '"' && quote != '\'' {
		return ""
	}

	start++
	end := strings.IndexByte(line[start:], quote)
	if end == -1 {
		return ""
	}

	return line[start : start+end]
}

func findColumnIndex(header []string, searchType string) int {
	for i, col := range header {
		if strings.EqualFold(col, searchType) {
			return i
		}
	}
	return -1
}

func parseUserRecord(header, record []string) map[string]string {
	user := make(map[string]string)
	for i, col := range header {
		if i < len(record) {
			user[col] = record[i]
		}
	}
	return user
}
