package commands

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp uninstall' command.
//

type UninstallCmd struct {
	Repo string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
	Keep bool   `short:"k" help:"Keep the addon files in res://addons/ but remove from tracking."`
}

func (cmd *UninstallCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	// Load the manifest.
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	fmt.Printf("Attempting to uninstall %s\n", cmd.Repo)

	// Remove the addon frfr
	m.RemoveAddon(cmd.Repo, cmd.Keep)
	manifest.SaveManifest(m)

	return nil
}
