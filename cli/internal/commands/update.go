package commands

import (
	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp update' command.
//

type UpdateCmd struct {
	RequiresManifestCmd
	Repos []string `arg:"" optional:"" name:"repos" help:"List of all specific addons to update."`
	Yes   bool     `short:"y" help:"Automatically confirm each update without user interaction."`
}

func (cmd *UpdateCmd) Run() error {
	// Filter target repos
	var targets []manifest.Addon

	if len(cmd.Repos) == 0 {
		for _, addon := range cmd.Manifest.Addons {
			if !addon.Untracked {
				targets = append(targets, addon)
			}
		}
	} else {
		for _, repoName := range cmd.Repos {
			_, addon, isTracked := cmd.Manifest.FindByRepo(repoName)
			if !isTracked {
				util.Warn("Addon '%s' is not tracked by Wisp, skipping...", repoName)
				continue
			}
			targets = append(targets, addon)
		}
	}

	if len(targets) == 0 {
		util.Info("No tracked addons found to update.")
		return nil
	}

	util.Info("Checking %d addon(s) for updates...", len(targets))
	updatedCount := 0

	// Check and update repos that need it.
	for _, addon := range targets {
		isUpToDate, ref, err := cmd.Manifest.CheckAddon(addon.Repo)
		if err != nil {
			util.Warn("Failed to check %s: %v", addon.Repo, err)
			continue
		}

		if isUpToDate || ref == nil {
			util.Success("%s is up to date.", addon.Repo)
			continue
		}

		util.Info("Update found for %s (%s -> %s)", addon.Repo, addon.GetCurrentVersion(), ref.GetVersion())

		// User confirmation
		if !cmd.Yes {
			if !util.Confirm(true, "Do you want to download and apply this update?") {
				util.Info("Skipping %s...", addon.Repo)
				continue
			}
		}

		// Download and apply.
		util.Info("Applying update...")
		loc, err := github.DownloadAndExtract(ref.GetZipballUrl())
		if err != nil {
			util.Error("Failed to download %s: %v", addon.Repo, err)
			continue
		}

		// Call the correct addition function.
		if addon.Type == manifest.Release {
			cmd.Manifest.AddRelease(loc, addon.Repo, ref.GetVersion())
		} else {
			cmd.Manifest.AddBranch(loc, addon.Repo, addon.GetCurrentBranch(), ref.GetVersion())
		}

		updatedCount++
	}

	// Only save when something was modified.
	if updatedCount > 0 {
		if err := manifest.SaveManifest(cmd.Manifest); err != nil {
			return err
		}
		util.Success("Successfully updated %d addon(s)!", updatedCount)
	} else {
		util.Success("All checked addons are up to date.")
	}

	return nil
}
