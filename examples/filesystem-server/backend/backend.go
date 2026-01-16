package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// FilesystemBackend implements filesystem operations
type FilesystemBackend struct {
	*backend.BaseBackend
	security *SecurityManager
}

// NewFilesystemBackend creates a new filesystem backend
func NewFilesystemBackend() *FilesystemBackend {
	b := &FilesystemBackend{
		BaseBackend: backend.NewBaseBackend("Filesystem Backend"),
	}

	b.registerTools()
	return b
}

// Initialize initializes the backend with configuration
func (b *FilesystemBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Get workspace root
	workspaceRoot := "./workspace"
	if root, ok := config["workspace_root"].(string); ok && root != "" {
		workspaceRoot = root
	}

	// Expand home directory
	if workspaceRoot[:2] == "~/" {
		home, _ := os.UserHomeDir()
		workspaceRoot = filepath.Join(home, workspaceRoot[2:])
	}

	// Get security settings
	secConfig := SecurityConfig{
		WorkspaceRoot:  workspaceRoot,
		MaxFileSize:    10 * 1024 * 1024, // 10MB
		MaxFilesPerDir: 1000,
		ReadOnly:       false,
	}

	if maxSize, ok := config["max_file_size"].(float64); ok {
		secConfig.MaxFileSize = int64(maxSize)
	}

	if readOnly, ok := config["read_only"].(bool); ok {
		secConfig.ReadOnly = readOnly
	}

	if allowedExts, ok := config["allowed_extensions"].([]interface{}); ok {
		secConfig.AllowedExts = make([]string, len(allowedExts))
		for i, ext := range allowedExts {
			secConfig.AllowedExts[i] = ext.(string)
		}
	}

	if blockedExts, ok := config["blocked_extensions"].([]interface{}); ok {
		secConfig.BlockedExts = make([]string, len(blockedExts))
		for i, ext := range blockedExts {
			secConfig.BlockedExts[i] = ext.(string)
		}
	}

	// Create security manager
	b.security = NewSecurityManager(secConfig)

	// Ensure workspace exists
	if err := b.security.EnsureWorkspace(); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

// registerTools registers all filesystem tools
func (b *FilesystemBackend) registerTools() {
	// File operations
	b.RegisterTool(
		backend.NewTool("file_create").
			Description("Create a new file with optional content").
			StringParam("path", "Path to the new file", true).
			StringParam("content", "Initial file content (optional)", false).
			Build(),
		b.handleFileCreate,
	)

	b.RegisterTool(
		backend.NewTool("file_read").
			Description("Read the contents of a file").
			StringParam("path", "Path to the file", true).
			Build(),
		b.handleFileRead,
	)

	b.RegisterTool(
		backend.NewTool("file_write").
			Description("Write or overwrite file content").
			StringParam("path", "Path to the file", true).
			StringParam("content", "Content to write", true).
			Build(),
		b.handleFileWrite,
	)

	b.RegisterTool(
		backend.NewTool("file_update").
			Description("Append content to an existing file").
			StringParam("path", "Path to the file", true).
			StringParam("content", "Content to append", true).
			Build(),
		b.handleFileUpdate,
	)

	b.RegisterTool(
		backend.NewTool("file_delete").
			Description("Delete a file").
			StringParam("path", "Path to the file", true).
			Build(),
		b.handleFileDelete,
	)

	b.RegisterTool(
		backend.NewTool("file_copy").
			Description("Copy a file to a new location").
			StringParam("source", "Source file path", true).
			StringParam("destination", "Destination file path", true).
			Build(),
		b.handleFileCopy,
	)

	b.RegisterTool(
		backend.NewTool("file_search").
			Description("Search for text in files").
			StringParam("path", "Directory to search in", true).
			StringParam("query", "Text to search for", true).
			BoolParam("case_sensitive", "Case sensitive search", false, boolPtr(false)).
			Build(),
		b.handleFileSearch,
	)

	b.RegisterTool(
		backend.NewTool("file_show_content").
			Description("Show file content with metadata").
			StringParam("path", "Path to the file", true).
			Build(),
		b.handleFileShowContent,
	)

	// Folder operations
	b.RegisterTool(
		backend.NewTool("folder_create").
			Description("Create a new directory").
			StringParam("path", "Path to the new directory", true).
			Build(),
		b.handleFolderCreate,
	)

	b.RegisterTool(
		backend.NewTool("folder_delete").
			Description("Delete a directory").
			StringParam("path", "Path to the directory", true).
			BoolParam("recursive", "Delete recursively", false, boolPtr(false)).
			Build(),
		b.handleFolderDelete,
	)

	b.RegisterTool(
		backend.NewTool("folder_rename").
			Description("Rename a directory").
			StringParam("old_path", "Current directory path", true).
			StringParam("new_path", "New directory path", true).
			Build(),
		b.handleFolderRename,
	)

	b.RegisterTool(
		backend.NewTool("folder_copy").
			Description("Copy a directory recursively").
			StringParam("source", "Source directory path", true).
			StringParam("destination", "Destination directory path", true).
			Build(),
		b.handleFolderCopy,
	)

	b.RegisterTool(
		backend.NewTool("folder_move").
			Description("Move a directory").
			StringParam("source", "Source directory path", true).
			StringParam("destination", "Destination directory path", true).
			Build(),
		b.handleFolderMove,
	)

	b.RegisterTool(
		backend.NewTool("folder_list").
			Description("List contents of a directory").
			StringParam("path", "Directory path", true).
			BoolParam("recursive", "List recursively", false, boolPtr(false)).
			Build(),
		b.handleFolderList,
	)
}

func boolPtr(b bool) *bool {
	return &b
}
