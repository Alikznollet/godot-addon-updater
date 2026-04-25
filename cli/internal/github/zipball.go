package github

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Downloads and extracts a Zipball to it's destination in res://addons/
// Will return the folder name if the extraction succeeds.
func DownloadAndExtract(zipUrl string) (string, error) {
	fmt.Println("Downloading zip archive...")

	// Download the zip.
	resp, err := http.Get(zipUrl)
	if err != nil {
		return "", fmt.Errorf("Failed to download zip: %v", err)
	}
	defer resp.Body.Close()

	// Create temp file to hold the zip.
	tmpFile, err := os.CreateTemp("", "godot-addon-*.zip")
	if err != nil {
		return "", fmt.Errorf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Stream the download into the temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to write to temp file: %v", err)
	}

	// Read the ZIP
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("Failed to read zip file: %v", err)
	}
	defer zipReader.Close()

	fmt.Println("Scouting Zip...")

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
		return "", fmt.Errorf("Could not find a valid 'addons/' folder inside the zip.")
	}

	fmt.Printf("Discovered addon: %s\n", discoveredFolderName)
	fmt.Println("Extracting...")

	// Remove the old files.
	targetDir := filepath.Join("addons", discoveredFolderName)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		fmt.Printf("Removing old version...\n")
		os.RemoveAll(targetDir)
	}

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
			return "", fmt.Errorf("Illegal file path detected: %s", cleanPath)
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
	}

	fmt.Printf("Extracted to %s!\n", targetDir)

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
