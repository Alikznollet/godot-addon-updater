package util

import (
	"io"
	"os"
	"path/filepath"
)

// CopyDir recursively copies a directory tree from src to dst.
// It returns the path to the first top-level directory copied (the addon folder itself).
func CopyDir(src string, dst string) (string, error) {
	var copiedAddonPath string

	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the source directory root
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip the root source directory itself (we don't want to copy "addons" into "addons")
		if relPath == "." {
			return nil
		}

		// Determine the destination path for this specific item
		targetPath := filepath.Join(dst, relPath)

		// If the relative path has no parent directory (meaning it's a direct child of src)
		// and it is a directory, this is our actual addon folder (e.g., "my_plugin").
		if copiedAddonPath == "" && d.IsDir() && filepath.Dir(relPath) == "." {
			copiedAddonPath = relPath
		}

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

	if err != nil {
		return "", err
	}

	return copiedAddonPath, nil
}

// copyFile remains exactly the same as your code...
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}
