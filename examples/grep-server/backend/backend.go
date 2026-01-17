package backend

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
func (gb *GrepBackend) handleGrepHTML(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
	// Extract arguments
	filePath := args["file_path"].(string)
	pattern := "href="
	if p, ok := args["pattern"].(string); ok && p != "" {
		pattern = p
	}

	// Resolve path
	fullPath := gb.resolvePath(filePath)

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w (tried: %s)", filePath, err, fullPath)
	}
	defer file.Close()

	// Count total lines for progress
	totalLines := countLines(fullPath)

	// Scan file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 0
	matchCount := 0

	for scanner.Scan() {
		// Check for cancellation
		select {
		case <-emit.Context().Done():
			return ctx.Err()
		default:
		}

		lineNum++
		line := scanner.Text()

		// Emit progress every 10 lines
		if lineNum%10 == 0 {
			emit.EmitProgress(
				int64(lineNum),
				int64(totalLines),
				fmt.Sprintf("Scanned %d/%d lines, found %d matches", lineNum, totalLines, matchCount),
			)
		}

		// Check if line contains pattern
		if strings.Contains(line, pattern) {
			matchCount++

			// Extract URL if searching for href
			url := extractURL(line, pattern)

			// Emit match
			emit.EmitData(map[string]interface{}{
				"line_number": lineNum,
				"line_text":   strings.TrimSpace(line),
				"url":         url,
				"match_count": matchCount,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Final progress
	emit.EmitProgress(
		int64(totalLines),
		int64(totalLines),
		fmt.Sprintf("Complete! Found %d matches in %d lines", matchCount, totalLines),
	)

	return nil
}

// handleSearchCSV searches CSV files by field
func (gb *GrepBackend) handleSearchCSV(ctx context.Context, args map[string]interface{}, emit backend.StreamingEmitter) error {
	filePath := args["file_path"].(string)
	searchType := args["search_type"].(string)
	searchValue := args["search_value"].(string)

	// Resolve path
	fullPath := gb.resolvePath(filePath)

	// Open CSV
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV %s: %w (tried: %s)", filePath, err, fullPath)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Find column index
	columnIndex := findColumnIndex(header, searchType)
	if columnIndex == -1 {
		return fmt.Errorf("unknown search type: %s (valid: name, email, age, salary, department)", searchType)
	}

	// Read all records for total count
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	totalRecords := len(records)
	matchCount := 0

	// Search records
	for i, record := range records {
		// Check cancellation
		select {
		case <-emit.Context().Done():
			return ctx.Err()
		default:
		}

		// Progress update every 5 records
		if i%5 == 0 {
			emit.EmitProgress(
				int64(i+1),
				int64(totalRecords),
				fmt.Sprintf("Searched %d/%d records, found %d matches", i+1, totalRecords, matchCount),
			)
		}

		// Check if matches
		if matchesSearch(record[columnIndex], searchValue, searchType) {
			matchCount++

			// Parse record
			user := parseUserRecord(header, record)

			// Emit match
			emit.EmitData(map[string]interface{}{
				"record_number": i + 1,
				"match_count":   matchCount,
				"user":          user,
				"matched_field": searchType,
				"matched_value": record[columnIndex],
			})
		}
	}

	// Final progress
	emit.EmitProgress(
		int64(totalRecords),
		int64(totalRecords),
		fmt.Sprintf("Search complete! Found %d matches out of %d records", matchCount, totalRecords),
	)

	return nil
}

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

func matchesSearch(value, searchValue, searchType string) bool {
	switch searchType {
	case "name", "email", "department", "status":
		// Case-insensitive substring match
		return strings.Contains(strings.ToLower(value), strings.ToLower(searchValue))

	case "age", "salary":
		// Numeric comparison - can search for exact or range
		recordNum, err := strconv.Atoi(value)
		if err != nil {
			return false
		}

		// Support ">X", "<X", or exact match
		if strings.HasPrefix(searchValue, ">") {
			threshold, _ := strconv.Atoi(strings.TrimPrefix(searchValue, ">"))
			return recordNum > threshold
		}
		if strings.HasPrefix(searchValue, "<") {
			threshold, _ := strconv.Atoi(strings.TrimPrefix(searchValue, "<"))
			return recordNum < threshold
		}

		searchNum, err := strconv.Atoi(searchValue)
		if err != nil {
			return false
		}
		return recordNum == searchNum

	default:
		return strings.Contains(strings.ToLower(value), strings.ToLower(searchValue))
	}
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
