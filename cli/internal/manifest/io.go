package manifest

// Defines the IO functionality for addons.json.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Returns a list of all folder names inside of res://addons
func GetAddonFolderContents() ([]string, error) {
	entries, err := os.ReadDir("addons")
	if err != nil {
		// If the folder doesn't exist yet there is nothing to scan.
		if os.IsNotExist(err) {
			return make([]string, 0), nil
		}
		return nil, err
	}

	folderNames := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip if the entry isn't a directory.
		}

		folderName := entry.Name()
		folderNames = append(folderNames, folderName)
	}

	return folderNames, nil
}

// Saves an AddonManifest object to addons.json.
func SaveManifest(manifest AddonManifest) error {
	// Serialize the object into the JSON format.
	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Write the manifest to the file.
	err = os.WriteFile("addons.json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write addons.json to disk: %w", err)
	}

	return nil
}

// Retrieves an AddonManifest object from addons.json.
func LoadManifest() (AddonManifest, error) {
	var manifest AddonManifest

	// Read the entire file.
	data, err := os.ReadFile("addons.json")
	if err != nil {
		// Return a clear error if the file doesn't exist
		if os.IsNotExist(err) {
			return manifest, errors.New("addons.json not found. Run 'wisp init' first.")
		}
		return manifest, fmt.Errorf("failed to read addons.json: %w", err)
	}

	// Parse the raw JSON to an AddonManifest object.
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("failed to parse addons.json: %w", err)
	}

	// If no addons are registered .Addons will come back as nil.
	// Fill it with an empty map if so.
	if manifest.Addons == nil {
		manifest.Addons = make(map[string]Addon)
	}

	return manifest, nil
}

// Removes an addon folder from the addons/ folder in the project.
// Based on the folderName passed.
func deleteAddonFolder(folderName string) error {
	// Check for any malicious paths.
	cleanPath := filepath.Clean(folderName)
	if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "/") || strings.Contains(cleanPath, "\\") {
		return fmt.Errorf("invalid folder name: %s", cleanPath)
	}

	// Actually remove the folder.
	targetDir := filepath.Join("addons", cleanPath)
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to delete folder %s: %v", targetDir, err)
	}

	return nil
}
