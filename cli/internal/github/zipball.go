package github

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// Downloads and extracts a Zipball to it's destination in res://addons/
// Will return the folder name if the extraction succeeds.
func DownloadAndExtract(zipUrl string) (string, error) {
	// Download the zip.
	resp, err := http.Get(zipUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download zip: %v", err)
	}
	defer resp.Body.Close()

	// Create temp file to hold the zip.
	tmpFile, err := os.CreateTemp("", "godot-addon-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Downloading loading bar.
	bar := util.NewDownloadBar(resp.ContentLength, "Downloading archive")

	// Stream the download into the temp file
	_, err = io.Copy(io.MultiWriter(tmpFile, bar), resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}

	bar.Finish()
	tmpFile.Close()

	// Read the ZIP
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read zip file: %v", err)
	}
	defer zipReader.Close()

	util.Info("Scouting Zip...")

	var discoveredFolderName string
	var prefixToRemove string

	for _, file := range zipReader.File {
		parts := strings.Split(file.Name, "/")

		for i, part := range parts {
			// If we found addons and there's files deeper still
			if part == "addons" && i+1 < len(parts) {
				potentialFolder := parts[i+1]

				// Ignore empty strings or files directly in addons
				if potentialFolder != "" && !strings.Contains(potentialFolder, ".") {
					discoveredFolderName = potentialFolder
					prefixToRemove = strings.Join(parts[:i+1], "/") + "/"
					break
				}
			}
		}

		if discoveredFolderName != "" {
			break // We found the addon!
		}
	}

	if discoveredFolderName == "" {
		return "", fmt.Errorf("could not find a valid 'addons/' folder inside zip")
	}

	util.Info("Discovered addon: %s", discoveredFolderName)

	// Remove the old files.
	targetDir := filepath.Join("addons", discoveredFolderName)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		util.Warn("Removing old version...")
		os.RemoveAll(targetDir)
	}

	extBar := util.NewItemBar(len(zipReader.File), "Extracting files")

	// Extraction.
	for _, file := range zipReader.File {
		// Only extract files that are inside the exact addon folder.
		if !strings.HasPrefix(file.Name, prefixToRemove+discoveredFolderName) {
			continue
		}

		cleanRelativePath := strings.TrimPrefix(file.Name, prefixToRemove)
		targetPath := filepath.Join("addons", cleanRelativePath)

		// Check for vulnerabilities
		cleanPath := filepath.Clean(targetPath)
		if strings.HasPrefix(cleanPath, "..") {
			return "", fmt.Errorf("illegal file path detected: %s", cleanPath)
		}

		// Create directories or extract files
		if file.FileInfo().IsDir() {
			os.MkdirAll(cleanPath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(cleanPath), os.ModePerm); err != nil {
			return "", err
		}
		if err := extractSingleFile(file, cleanPath); err != nil {
			return "", err
		}

		extBar.Add(1)
	}

	// Finish the progress bar.
	extBar.Finish()
	util.Success("Extracted to %s!", targetDir)

	return discoveredFolderName, nil
}

// Helper function that extracts a single file to disk.
func extractSingleFile(file *zip.File, targetPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create an empty file
	dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the bytes from the zip to the file.
	_, err = io.Copy(dst, src)
	return err
}
