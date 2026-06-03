package commands

import (
	"github.com/alikznollet/godot-wisp/cli/internal/git"
	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/godot"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp install' command.
//

type InstallCmd struct {
	RequiresManifestCmd
	Repo   string `arg:"" name:"repo" help:"The repository formatted as a url or just the owner and repo name (e.g. ramokz/phantom-camera or https://github.com/ramokz/phantom-camera)."`
	Tag    string `short:"t" xor:"target" help:"Specific version tag to install (e.g. v1.0.0)."`
	Branch string `short:"b" xor:"target" help:"Branch to track instead of tracking releases (e.g. main)."`
}

func (cmd *InstallCmd) Run() error {
	// Split the repo name
	repoInfo, err := git.ParseRepoString(cmd.Repo)
	if err != nil {
		return err
	}

	var updated bool
	if cmd.Branch != "" {
		updated, err = cmd.installBranch(repoInfo)
	} else {
		version := cmd.Tag
		if version == "" {
			version = "latest"
		}
		updated, err = cmd.installRelease(repoInfo, version)
	}

	if err != nil {
		return err
	}

	// Only run updating procedure when a new file arrived.
	if updated {
		folderName, _, _ := cmd.Manifest.FindByRepo(cmd.Repo)

		if util.Confirm(false, "Enable '%s' in '%s'?", cmd.Repo, godot.ProjectFile) {
			if err := godot.EnableAddon(folderName); err != nil {
				util.Warn("Failed to auto-enable addon: %v", err)
			} else {
				util.Success("Addon enabled!")
			}
		}

		// Save the manifest.
		if err := manifest.SaveManifest(cmd.Manifest); err != nil {
			return err
		}
		util.Success("Addon '%s' installed successfully!", cmd.Repo)
	}
	return nil
}

// Install a branch from github.
func (cmd *InstallCmd) installBranch(repoInfo git.RepoInfo) (bool, error) {
	util.Info("Installing %s (Branch: %s)", cmd.Repo, cmd.Branch)

	// Fetch the latest commit from the target branch
	branchData, err := github.GetAddonRef(repoInfo, cmd.Branch, true)
	if err != nil {
		return false, err
	}

	// Extract the commit for fetching.
	commitHash := branchData.GetVersion()
	_, addon, isTracked := cmd.Manifest.FindByRepo(cmd.Repo)

	if isTracked {
		if addon.Commit != "" {
			if addon.Commit == commitHash {
				util.Success("Addon is already up to date!")
				return false, nil
			}
			util.Info("Found a newer version...")
		} else {
			util.Info("Switching tracking from release mode to branch mode for %s...", cmd.Repo)
		}
	} else {
		util.Info("Tracking branch '%s'...", cmd.Branch)
	}

	// Build the URL and download/extract the files.
	loc, err := git.GitDownload(repoInfo.BuildRepoUrl(), branchData.GetVersion())
	if err != nil {
		return false, err
	}

	// Make sure to pass the full repo name and branch+commit.
	cmd.Manifest.AddBranch(loc, repoInfo, cmd.Branch, commitHash)
	return true, nil
}

// Install a Release from github.
func (cmd *InstallCmd) installRelease(repoInfo git.RepoInfo, version string) (bool, error) {
	util.Info("Installing %s (Release: %s)", cmd.Repo, version)

	// Fetch the target release from github.
	release, err := github.GetAddonRef(repoInfo, version, false)
	if err != nil {
		return false, err
	}

	_, addon, isTracked := cmd.Manifest.FindByRepo(cmd.Repo)

	if isTracked {
		if release.GetVersion() == addon.Version {
			util.Success("Addon is already up to date!")
			return false, nil
		} else if addon.Commit != "" {
			util.Info("Switching tracking from branch mode to release mode for %s...", cmd.Repo)
		} else {
			util.Info("Found a newer version...")
		}
	} else {
		util.Info("Tracking release...")
	}

	loc, err := git.GitDownload(repoInfo.BuildRepoUrl(), release.GetVersion())
	if err != nil {
		return false, err
	}

	// Make sure to pass the full repo name to the Addon.
	cmd.Manifest.AddRelease(loc, repoInfo, release.GetVersion())
	return true, nil
}
