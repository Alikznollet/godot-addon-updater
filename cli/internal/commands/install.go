package commands

import (
	"fmt"
	"strings"

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
	Repo   string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
	Tag    string `short:"t" xor:"target" help:"Specific version tag to install (e.g. v1.0.0)."`
	Branch string `short:"b" xor:"target" help:"Branch to track instead of tracking releases (e.g. main)."`
}

func (cmd *InstallCmd) Run() error {
	// Split the repo name
	parts := strings.Split(cmd.Repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository format. Must be 'owner/repo'")
	}
	owner, repo := parts[0], parts[1]

	var err error
	var updated bool
	if cmd.Branch != "" {
		updated, err = cmd.installBranch(owner, repo)
	} else {
		version := cmd.Tag
		if version == "" {
			version = "latest"
		}
		updated, err = cmd.installRelease(owner, repo, version)
	}

	if err != nil {
		return err
	}

	// Only run updating procedure when a new file arrived.
	if updated {
		folderName, _, _ := cmd.Manifest.FindByRepo(cmd.Repo)

		if util.Confirm(false, "Enable '%s' in '%s'?", repo, godot.ProjectFile) {
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
func (cmd *InstallCmd) installBranch(owner string, repo string) (bool, error) {
	util.Info("Installing %s (Branch: %s)", cmd.Repo, cmd.Branch)

	// Fetch the latest commit from the target branch
	branchData, err := github.GetAddonRef(owner, repo, cmd.Branch, true)
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
	zipUrl := branchData.GetZipballUrl()
	loc, err := github.DownloadAndExtract(zipUrl)
	if err != nil {
		return false, err
	}

	// Make sure to pass the full repo name and branch+commit.
	cmd.Manifest.AddBranch(loc, cmd.Repo, cmd.Branch, commitHash)
	return true, nil
}

// Install a Release from github.
func (cmd *InstallCmd) installRelease(owner string, repo string, version string) (bool, error) {
	util.Info("Installing %s (Release: %s)", cmd.Repo, version)

	// Fetch the target release from github.
	release, err := github.GetAddonRef(owner, repo, version, false)
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

	loc, err := github.DownloadAndExtract(release.GetZipballUrl())
	if err != nil {
		return false, err
	}

	// Make sure to pass the full repo name to the Addon.
	cmd.Manifest.AddRelease(loc, cmd.Repo, release.GetVersion())
	return true, nil
}
