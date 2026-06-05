package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/alikznollet/godot-wisp/cli/internal/git"
	"github.com/alikznollet/godot-wisp/cli/internal/godot"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp init' command.
//

// Initialization Command structure used by Kong.
type InitCmd struct {
	RequiresGodotProjectCmd
	Force bool `short:"f" help:"Overwrites existing addons.json."`
}

// Code Ran by the initialization.
func (cmd *InitCmd) Run(ctx context.Context) error {
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
	util.Success("Initialized '%s' for project '%s'", manifest.ManifestName, path)

	// Load the manifest.
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	util.Info("Wisp has a Godot editor plugin that lets you check for updates directly inside the engine!")
	install := util.Confirm(true, "Would you like to install it now?")

	if install {
		util.Info("Installing Wisp Godot Plugin...")

		// Firstly format repoInfo
		repoInfo, err := git.ParseRepoString("alikznollet/godot-wisp")
		if err != nil {
			return err
		}

		// Then get the release info.
		release, err := git.GetAddonRef(ctx, repoInfo, "latest", false)
		if err != nil {
			return err
		}

		loc, err := git.GitDownload(ctx, repoInfo.BuildRepoUrl(), release.GetVersion())
		if err != nil {
			return err
		}

		m.AddRelease(loc, repoInfo, release.GetVersion())

		// Save the manifest with wisp installed
		if err := manifest.SaveManifest(m); err != nil {
			return err
		}

		// Enable wisp in the project.godot
		if err := godot.EnableAddon(loc); err != nil {
			return fmt.Errorf("failed to enable addon in project.godot: %v", err)
		}

		util.Success("Wisp Godot Plugin successfully installed and enabled!")
	} else {
		util.Info("Skipping plugin installation. You can always install it later!")
	}

	return nil
}
