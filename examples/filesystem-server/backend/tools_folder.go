package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// handleFolderCreate creates a new directory
func (b *FilesystemBackend) handleFolderCreate(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "write"); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"success": true,
		"path":    relPath,
		"message": fmt.Sprintf("Directory created: %s", relPath),
	}, nil
}

// handleFolderDelete deletes a directory
func (b *FilesystemBackend) handleFolderDelete(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)

	recursive := false
	if r, ok := args["recursive"].(bool); ok {
		recursive = r
	}

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "delete"); err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("directory not found: %s", path)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	if recursive {
		err = os.RemoveAll(fullPath)
	} else {
		err = os.Remove(fullPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to delete directory: %w", err)
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"success":   true,
		"path":      relPath,
		"recursive": recursive,
		"message":   fmt.Sprintf("Directory deleted: %s", relPath),
	}, nil
}

// handleFolderRename renames a directory
func (b *FilesystemBackend) handleFolderRename(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	oldPath := args["old_path"].(string)
	newPath := args["new_path"].(string)

	oldFull, err := b.security.ValidatePath(oldPath)
	if err != nil {
		return nil, fmt.Errorf("invalid old path: %w", err)
	}

	newFull, err := b.security.ValidatePath(newPath)
	if err != nil {
		return nil, fmt.Errorf("invalid new path: %w", err)
	}

	if err := b.security.ValidateFileOperation(oldPath, "write"); err != nil {
		return nil, err
	}

	if err := os.Rename(oldFull, newFull); err != nil {
		return nil, fmt.Errorf("failed to rename directory: %w", err)
	}

	oldRel, _ := b.security.GetRelativePath(oldFull)
	newRel, _ := b.security.GetRelativePath(newFull)

	return map[string]interface{}{
		"success":  true,
		"old_path": oldRel,
		"new_path": newRel,
		"message":  fmt.Sprintf("Directory renamed: %s → %s", oldRel, newRel),
	}, nil
}

// handleFolderCopy copies a directory recursively
func (b *FilesystemBackend) handleFolderCopy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	filesCopied := 0
	bytesCopied := int64(0)

	err = filepath.Walk(srcFull, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcFull, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dstFull, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		written, err := io.Copy(dst, src)
		if err != nil {
			return err
		}

		filesCopied++
		bytesCopied += written

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to copy directory: %w", err)
	}

	srcRel, _ := b.security.GetRelativePath(srcFull)
	dstRel, _ := b.security.GetRelativePath(dstFull)

	return map[string]interface{}{
		"success":      true,
		"source":       srcRel,
		"destination":  dstRel,
		"files_copied": filesCopied,
		"bytes_copied": bytesCopied,
		"message":      fmt.Sprintf("Directory copied: %s → %s", srcRel, dstRel),
	}, nil
}

// handleFolderMove moves a directory
func (b *FilesystemBackend) handleFolderMove(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	if err := b.security.ValidateFileOperation(srcPath, "write"); err != nil {
		return nil, err
	}

	if err := os.Rename(srcFull, dstFull); err != nil {
		return nil, fmt.Errorf("failed to move directory: %w", err)
	}

	srcRel, _ := b.security.GetRelativePath(srcFull)
	dstRel, _ := b.security.GetRelativePath(dstFull)

	return map[string]interface{}{
		"success":     true,
		"source":      srcRel,
		"destination": dstRel,
		"message":     fmt.Sprintf("Directory moved: %s → %s", srcRel, dstRel),
	}, nil
}

// handleFolderList lists directory contents
func (b *FilesystemBackend) handleFolderList(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path := args["path"].(string)

	recursive := false
	if r, ok := args["recursive"].(bool); ok {
		recursive = r
	}

	fullPath, err := b.security.ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if err := b.security.ValidateFileOperation(path, "read"); err != nil {
		return nil, err
	}

	var entries []map[string]interface{}

	if recursive {
		err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if path == fullPath {
				return nil
			}

			relPath, _ := b.security.GetRelativePath(path)

			entry := map[string]interface{}{
				"name":        filepath.Base(path),
				"path":        relPath,
				"is_dir":      info.IsDir(),
				"size":        info.Size(),
				"modified":    info.ModTime().Format(time.RFC3339),
				"permissions": info.Mode().String(),
			}

			entries = append(entries, entry)
			return nil
		})
	} else {
		dirEntries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range dirEntries {
			info, _ := entry.Info()
			entryPath := filepath.Join(fullPath, entry.Name())
			relPath, _ := b.security.GetRelativePath(entryPath)

			e := map[string]interface{}{
				"name":        entry.Name(),
				"path":        relPath,
				"is_dir":      entry.IsDir(),
				"size":        info.Size(),
				"modified":    info.ModTime().Format(time.RFC3339),
				"permissions": info.Mode().String(),
			}

			entries = append(entries, e)
		}
	}

	relPath, _ := b.security.GetRelativePath(fullPath)

	return map[string]interface{}{
		"path":      relPath,
		"entries":   entries,
		"count":     len(entries),
		"recursive": recursive,
	}, nil
}
