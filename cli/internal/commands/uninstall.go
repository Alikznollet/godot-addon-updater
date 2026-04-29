package commands

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
)

//
// This file holds the 'wisp uninstall' command.
//

type UninstallCmd struct {
	RequiresManifestCmd
	Repo string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
	Keep bool   `short:"k" help:"Keep the addon files in res://addons/ but remove from tracking."`
}

func (cmd *UninstallCmd) Run() error {
	fmt.Printf("Attempting to uninstall %s\n", cmd.Repo)

	// Remove the addon frfr
	cmd.Manifest.RemoveAddon(cmd.Repo, cmd.Keep)
	manifest.SaveManifest(cmd.Manifest)

	return nil
}
