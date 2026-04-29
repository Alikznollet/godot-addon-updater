package commands

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/godot"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp uninstall' command.
//

type UninstallCmd struct {
	RequiresManifestCmd
	Repo string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
}

func (cmd *UninstallCmd) Run() error {
	folderName, _, isTracked := cmd.Manifest.FindByRepo(cmd.Repo)
	if !isTracked {
		return fmt.Errorf("addon '%s' is not tracked", cmd.Repo)
	}

	util.Info("Uninstalling %s...", cmd.Repo)

	remove := util.Confirm(false, "Do you want to remove the associated folder '%s' and disable the addon?", folderName)
	if remove {
		if err := godot.DisableAddon(folderName); err != nil {
			util.Warn("Could not modify '%s': %v", godot.ProjectFile, err)
		} else {
			util.Info("Disabled plugin in '%s'.", godot.ProjectFile)
		}
	} else {
		util.Info("Untracking %s (keeping files on disk)...", cmd.Repo)
	}

	// Remove the addon.
	if err := cmd.Manifest.RemoveAddon(cmd.Repo, !remove); err != nil {
		return err
	}

	if err := manifest.SaveManifest(cmd.Manifest); err != nil {
		return err
	}

	util.Success("Successfully uninstalled '%s'!", cmd.Repo)
	return nil
}
