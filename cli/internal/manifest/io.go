package manifest

// Defines the IO functionality for addons.json.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// Name of the manifest file.
const ManifestName = "wisp.json"

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

// Initialize a new manifest or return an error message if it fails.
func InitManifest(force bool) error {
	if _, err := os.Stat(ManifestName); err == nil {
		if !force {
			return fmt.Errorf("'%s' already exists, use --force to overwrite", ManifestName)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	m := AddonManifest{
		Addons:        make(map[string]Addon),
		SchemaVersion: CurrentSchemaVersion, // Pin the current schema version.
	}
	return SaveManifest(&m)
}

// Saves an AddonManifest object to addons.json.
func SaveManifest(manifest *AddonManifest) error {
	// Serialize the object into the JSON format.
	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Write the manifest to a temp file.
	tmpFile := fmt.Sprintf("%s.tmp", ManifestName)
	if err = os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary manifest: %v", err)
	}

	// Atomically swap it.
	if err := os.Rename(tmpFile, ManifestName); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to update wisp.json: %v", err)
	}

	return nil
}

// Retrieves an AddonManifest object from addons.json.
func LoadManifest() (*AddonManifest, error) {
	var manifest *AddonManifest

	// Read the entire file.
	data, err := os.ReadFile(ManifestName)
	if err != nil {
		// Return a clear error if the file doesn't exist
		if os.IsNotExist(err) {
			return manifest, fmt.Errorf("'%s' not found. Run 'wisp init' first.", ManifestName)
		}
		return manifest, fmt.Errorf("failed to read '%s': %w", ManifestName, err)
	}

	// Parse the raw JSON to an AddonManifest object.
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("failed to parse '%s': %w", ManifestName, err)
	}

	// Check for new version of the manifest.
	// ! Any migrations from older to newer version can be prompted here.
	if manifest.SchemaVersion > CurrentSchemaVersion {
		return manifest, fmt.Errorf("this project uses a newer wisp.json format. Please update your Wisp CLI.")
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

// Fetches outdated addons from the addons.json file.
func FetchOutdatedAddons(ctx context.Context, m *AddonManifest) ([]OutdatedAddon, error) {
	var outdated []OutdatedAddon = make([]OutdatedAddon, 0)

	util.Info("Checking for updates...")

	for folderName, addon := range m.Addons {
		isUpToDate, ref, err := m.CheckAddon(ctx, addon.RepoInfo)
		if err != nil {
			util.Warn("Failed to check %s: %v", addon.RepoInfo.Repo, err)
			continue
		}

		if !isUpToDate && ref != nil {
			util.Info("Update found for %s (%s -> %s)", addon.RepoInfo.Repo, addon.GetCurrentVersion(), ref.GetVersion())

			outdated = append(outdated, OutdatedAddon{
				Folder:   folderName,
				RepoInfo: addon.RepoInfo,
				Current:  addon.GetCurrentVersion(),
				Latest:   ref.GetVersion(),
				Type:     addon.Type,
				Branch:   addon.GetCurrentBranch(),
			})
		}
	}

	return outdated, nil
}
