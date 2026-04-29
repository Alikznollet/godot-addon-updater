package commands

import (
	"os"

	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp init' command.
//

// Initialization Command structure used by Kong.
type InitCmd struct {
	Force bool `short:"f" help:"Overwrites existing addons.json."`
}

// Code Ran by the initialization.
func (cmd *InitCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	util.Info("Initializing '%s'...", manifest.ManifestName)

	// Initialize the manifest.
	err := manifest.InitManifest(cmd.Force)
	if err != nil {
		return err
	}

	// Grab the project path
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	util.Success("Initialized '%s' for project %s\n", manifest.ManifestName, path)

	return nil
}
