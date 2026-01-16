package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// handleFileCreate creates a new file
func (b *FilesystemBackend) handleFileCreate(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)
	content := ""
	if c, ok := args["content"].(string); ok {
		content = c
	}

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "write"); err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileSize(int64(len(content))); err != nil {
		return nil, err
	}

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return nil, fmt.Errorf("file already exists: %s", path)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"success": true,
		"path":    relPath,
		"size":    len(content),
		"message": fmt.Sprintf("File created: %s", relPath),
	}, nil
}

// handleFileRead reads file content
func (b *FilesystemBackend) handleFileRead(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "read"); err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"path":        relPath,
		"content":     string(content),
		"size":        len(content),
		"modified":    info.ModTime().Format(time.RFC3339),
		"permissions": info.Mode().String(),
	}, nil
}

// handleFileWrite writes/overwrites file content
func (b *FilesystemBackend) handleFileWrite(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)
	content := args["content"].(string)

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "write"); err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileSize(int64(len(content))); err != nil {
		return nil, err
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"success": true,
		"path":    relPath,
		"size":    len(content),
		"message": fmt.Sprintf("File written: %s", relPath),
	}, nil
}

// handleFileUpdate appends content to file
func (b *FilesystemBackend) handleFileUpdate(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)
	content := args["content"].(string)

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "write"); err != nil {
		return nil, err
	}

	// Read existing content
	existing, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read existing file: %w", err)
	}

	newContent := string(existing) + content

	if err := b.security.ValidateFileSize(int64(len(newContent))); err != nil {
		return nil, err
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"success":  true,
		"path":     relPath,
		"size":     len(newContent),
		"appended": len(content),
		"message":  fmt.Sprintf("File updated: %s", relPath),
	}, nil
}

// handleFileDelete deletes a file
func (b *FilesystemBackend) handleFileDelete(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "delete"); err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, use folder_delete: %s", path)
	}

	if err := os.Remove(fullPath); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"success": true,
		"path":    relPath,
		"message": fmt.Sprintf("File deleted: %s", relPath),
	}, nil
}

// handleFileCopy copies a file
func (b *FilesystemBackend) handleFileCopy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	srcPath := args["source"].(string)
	dstPath := args["destination"].(string)

	srcFull, err := b.security.ValidatePath(srcPath)
	if err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}

	dstFull, err := b.security.ValidatePath(dstPath)
	if err != nil {
		return nil, fmt.Errorf("invalid destination path: %w", err)
	}

	if err := b.security.ValidateFileOperation(srcPath, "read"); err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(dstPath, "write"); err != nil {
		return nil, err
	}

	// Open source
	src, err := os.Open(srcFull)
	if err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	// Create destination
	if err := os.MkdirAll(filepath.Dir(dstFull), 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	dst, err := os.Create(dstFull)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination: %w", err)
	}
	defer dst.Close()

	// Copy
	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	srcRel, _ := b.security.GetRelativePath(srcFull)
	dstRel, _ := b.security.GetRelativePath(dstFull)

	return map[string]interface{}{
		"success":     true,
		"source":      srcRel,
		"destination": dstRel,
		"size":        written,
		"message":     fmt.Sprintf("File copied: %s → %s", srcRel, dstRel),
	}, nil
}

// handleFileSearch searches for text in files
func (b *FilesystemBackend) handleFileSearch(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	searchPath := args["path"].(string)
	query := args["query"].(string)

	caseSensitive := false
	if cs, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	fullPath, err := b.security.ValidatePath(searchPath)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(searchPath, "read"); err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		text := string(content)
		searchIn := text
		searchFor := query

		if !caseSensitive {
			searchIn = strings.ToLower(text)
			searchFor = strings.ToLower(query)
		}

		if strings.Contains(searchIn, searchFor) {
			relPath, _ := b.security.GetRelativePath(path)

			// Count occurrences
			count := strings.Count(searchIn, searchFor)

			results = append(results, map[string]interface{}{
				"path":     relPath,
				"matches":  count,
				"size":     len(content),
				"modified": info.ModTime().Format(time.RFC3339),
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	}, nil
}

// handleFileShowContent shows file content with metadata
func (b *FilesystemBackend) handleFileShowContent(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "read"); err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", path)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	// Count lines
	lines := strings.Count(string(content), "\n") + 1

	// Detect if binary
	isBinary := false
	for _, b := range content[:min(512, len(content))] {
		if b == 0 {
			isBinary = true
			break
		}
	}

	result := map[string]interface{}{
		"path":        relPath,
		"size":        len(content),
		"lines":       lines,
		"modified":    info.ModTime().Format(time.RFC3339),
		"permissions": info.Mode().String(),
		"is_binary":   isBinary,
	}

	if !isBinary {
		result["content"] = string(content)
	} else {
		result["message"] = "Binary file (content not displayed)"
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
