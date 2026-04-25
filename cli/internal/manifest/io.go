package manifest

// Defines the IO functionality for addons.json.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Saves an AddonManifest object to addons.json.
func SaveManifest(manifest AddonManifest) error {
	// Serialize the object into the JSON format.
	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to encode JSON: %w", err)
	}

	// Write the manifest to the file.
	err = os.WriteFile("addons.json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write addons.json to disk: %w", err)
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
			return manifest, errors.New("addons.json not found. Run 'gau init' first.")
		}
		return manifest, fmt.Errorf("Failed to read addons.json: %w", err)
	}

	// Parse the raw JSON to an AddonManifest object.
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("Failed to parse addons.json: %w", err)
	}

	// If no addons are registered .Addons will come back as nil.
	// Fill it with an empty map if so.
	if manifest.Addons == nil {
		manifest.Addons = make(map[string]Addon)
	}

	return manifest, nil
}
