package util

import (
	"io"
	"os"
	"path/filepath"
)

// Recursively copies a directory tree from src to dst.
func CopyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the source directory root
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Determine the destination path for this specific item
		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			// Get the permissions of the source directory
			info, err := d.Info()
			if err != nil {
				return err
			}
			// Create the directory if it doesn't exist
			return os.MkdirAll(targetPath, info.Mode())
		}

		// If it's a file, copy its contents
		return copyFile(path, targetPath)
	})
}

// Copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Ensure the parent directory exists just in case
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Create or truncate the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the bytes
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Sync to disk and match permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}
