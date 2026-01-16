package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SecurityConfig holds security settings
type SecurityConfig struct {
	WorkspaceRoot  string
	MaxFileSize    int64
	MaxFilesPerDir int
	AllowedExts    []string
	BlockedExts    []string
	ReadOnly       bool
	EnableSymlinks bool
}

// SecurityManager handles path validation and sandboxing
type SecurityManager struct {
	config SecurityConfig
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityConfig) *SecurityManager {
	// Set defaults
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 10MB
	}
	if config.MaxFilesPerDir == 0 {
		config.MaxFilesPerDir = 1000
	}

	return &SecurityManager{
		config: config,
	}
}

// ValidatePath validates and resolves a path within the workspace
func (sm *SecurityManager) ValidatePath(path string) (string, error) {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Prevent absolute paths
	if filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("absolute paths not allowed: %s", path)
	}

	// Prevent path traversal
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal attempt detected: %s", path)
	}

	// Join with workspace root
	fullPath := filepath.Join(sm.config.WorkspaceRoot, cleanPath)

	// Resolve symlinks
	if sm.config.EnableSymlinks {
		var err error
		fullPath, err = filepath.EvalSymlinks(fullPath)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to resolve path: %w", err)
		}
	}

	// Verify it's still within workspace
	absWorkspace, err := filepath.Abs(sm.config.WorkspaceRoot)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// For non-existent paths, check parent directory
	if os.IsNotExist(err) {
		absPath = filepath.Dir(absPath)
	}

	relPath, err := filepath.Rel(absWorkspace, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path outside workspace: %s", path)
	}

	return fullPath, nil
}

// ValidateFileOperation checks if a file operation is allowed
func (sm *SecurityManager) ValidateFileOperation(path string, operation string) error {
	if sm.config.ReadOnly && operation != "read" {
		return fmt.Errorf("read-only mode enabled, operation not allowed: %s", operation)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))

	// Check blocked extensions
	for _, blocked := range sm.config.BlockedExts {
		if ext == strings.ToLower(blocked) {
			return fmt.Errorf("file extension blocked: %s", ext)
		}
	}

	// Check allowed extensions (if whitelist exists)
	if len(sm.config.AllowedExts) > 0 {
		allowed := false
		for _, allowedExt := range sm.config.AllowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension not in whitelist: %s", ext)
		}
	}

	return nil
}

// ValidateFileSize checks if file size is within limits
func (sm *SecurityManager) ValidateFileSize(size int64) error {
	if size > sm.config.MaxFileSize {
		return fmt.Errorf("file size exceeds limit: %d > %d bytes", size, sm.config.MaxFileSize)
	}
	return nil
}

// ValidateDirectorySize checks number of files in directory
func (sm *SecurityManager) ValidateDirectorySize(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	if len(entries) > sm.config.MaxFilesPerDir {
		return fmt.Errorf("directory contains too many files: %d > %d", len(entries), sm.config.MaxFilesPerDir)
	}

	return nil
}

// GetRelativePath returns path relative to workspace
func (sm *SecurityManager) GetRelativePath(fullPath string) (string, error) {
	return filepath.Rel(sm.config.WorkspaceRoot, fullPath)
}

// EnsureWorkspace creates workspace directory if it doesn't exist
func (sm *SecurityManager) EnsureWorkspace() error {
	return os.MkdirAll(sm.config.WorkspaceRoot, 0755)
}
