package lockfile

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alikznollet/godot-addon-updater/internal/types"
)

// Saves an AddonManifest object to addons.json.
// Will fail if the folder where the command tool was
// executed from does not contain a project.godot file.
func SaveManifest(manifest types.AddonManifest) error {
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
