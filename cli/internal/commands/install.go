package commands

import (
	"fmt"
	"strings"

	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp install' command.
//

type InstallCmd struct {
	Repo    string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
	Version string `short:"v" xor:"target" help:"Specific version tag to install (e.g. v1.0.0)."`
	Branch  string `short:"b" xor:"target" help:"Branch to track instead of tracking releases (e.g. main)."`
}

func (cmd *InstallCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	// Grab the manifest and versions
	targetVersion := cmd.Version
	targetBranch := cmd.Branch
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	// Split the repo name
	parts := strings.Split(cmd.Repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("Invalid repository format. Must be 'owner/repo'.")
	}

	owner := parts[0]
	repo := parts[1]

	// If nothing specified we default to the latest release.
	if targetVersion == "" && targetBranch == "" {
		targetVersion = "latest"
	}

	if targetBranch != "" {
		fmt.Printf("Installing %s (Tracking Branch: %s)\n", cmd.Repo, targetBranch)

		// Fetch the latest commit from the target branch
		branchData, err := github.GetAddonRef(owner, repo, targetBranch, true)
		if err != nil {
			return err
		}

		// Extract the commit for fetching.
		commitHash := branchData.GetVersion()
		_, addon, isTracked := m.FindByRepo(cmd.Repo)

		if isTracked {
			if addon.Commit != "" {
				if addon.Commit == commitHash {
					fmt.Printf("%s is already up to date on branch %s (%s)\n", cmd.Repo, targetBranch, commitHash)
				}
				fmt.Printf("Updating %s from %s -> %s...\n", cmd.Repo, addon.Commit, commitHash)
			} else {
				fmt.Printf("Switching to branch tracking for %s on %s (%s).\n", cmd.Repo, targetBranch, commitHash)
			}
		} else {
			fmt.Printf("Tracking branch %s (%s)...\n", targetBranch, commitHash)
		}

		// Build the URL and download/extract the files.
		zipUrl := branchData.GetZipballUrl()
		loc, err := github.DownloadAndExtract(zipUrl)
		if err != nil {
			return err
		}

		// Make sure to pass the full repo name and branch+commit.
		m.AddBranch(loc, cmd.Repo, targetBranch, commitHash)
	} else {
		fmt.Printf("Installing %s (Release Version: %s)\n", cmd.Repo, targetVersion)

		// Fetch the target release from github.
		release, err := github.GetAddonRef(owner, repo, targetVersion, false)
		if err != nil {
			return err
		}

		_, addon, isTracked := m.FindByRepo(cmd.Repo)

		if isTracked {
			if release.GetVersion() == addon.Version {
				fmt.Printf("%s is already up to date (%s)\n", cmd.Repo, release.GetVersion())
				return nil
			}
			fmt.Printf("Updating %s from %s -> %s...\n", cmd.Repo, addon.Version, release.GetVersion())
		} else {
			fmt.Printf("Tracking releases (%s)...\n", release.GetVersion())
		}

		loc, err := github.DownloadAndExtract(release.GetZipballUrl())
		if err != nil {
			return err
		}

		// Make sure to pass the full repo name to the Addon.
		m.AddRelease(loc, cmd.Repo, release.GetVersion())
	}

	// TODO: Automatically enable addons when installed?
	manifest.SaveManifest(m) // Make sure to save.
	fmt.Println("Addon installed successfully!")

	return nil
}
