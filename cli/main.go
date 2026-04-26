// Copyright 2026 Alikznollet
// GNU GPL

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/alikznollet/godot-addon-updater/internal/github"
	"github.com/alikznollet/godot-addon-updater/internal/manifest"
	"github.com/alikznollet/godot-addon-updater/internal/util"
)

// -- Command Structs -- //

// Initialization Command structure used by Kong.
type InitCmd struct {
	Force bool `short:"f" help:"Overwrites existing addons.json."`
}

// Code Ran by the initialization.
func (cmd *InitCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	fmt.Println("Initializing addons.json...")
	fmt.Printf("Force Overwrite: %v\n", cmd.Force)

	if _, err := os.Stat("addons.json"); err == nil {
		if !cmd.Force {
			return errors.New("addons.json already exists. Use --force to overwrite.")
		}
		fmt.Println("Overwriting existing addons.json...")
	} else if !os.IsNotExist(err) {
		return err
	}

	// Create and save an empty manifest.
	m := manifest.AddonManifest{
		Addons: make(map[string]manifest.Addon),
	}
	err := manifest.SaveManifest(m)

	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("Initialized addons.json file at %s\n", path)

	return nil
}

// Install

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
		branchData, err := github.GetBranch(owner, repo, targetBranch)
		if err != nil {
			return err
		}

		// Extract the commit for fetching.
		commitHash := branchData.Commit.Sha
		_, addon, isTracked := m.FindByRepo(cmd.Repo)

		if isTracked {
			if addon.Commit != "" {
				if addon.Commit == commitHash {
					fmt.Printf("%s is already up to date on branch %s (%s)\n", cmd.Repo, targetBranch, commitHash)
				}
				fmt.Printf("Updating %s from %s -> %s...\n", cmd.Repo, addon.Commit, commitHash)
			} else {
				fmt.Printf("Switching to branch tracking for %s on %s (%s).", cmd.Repo, targetBranch, commitHash)
			}
		} else {
			fmt.Printf("Tracking branch %s (%s)...\n", targetBranch, commitHash)
		}

		// Build the URL and download/extract the files.
		zipUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", owner, repo, commitHash)
		loc, err := github.DownloadAndExtract(zipUrl)
		if err != nil {
			return err
		}

		// Make sure to pass the full repo name and branch+commit.
		m.AddBranch(loc, cmd.Repo, targetBranch, commitHash)
	} else {
		fmt.Printf("Installing %s (Release Version: %s)\n", cmd.Repo, targetVersion)

		// Fetch the target release from github.
		release, err := github.GetRelease(owner, repo, targetVersion)
		if err != nil {
			return err
		}

		_, addon, isTracked := m.FindByRepo(cmd.Repo)

		if isTracked {
			if release.TagName == addon.Version {
				fmt.Printf("%s is already up to date (%s)\n", cmd.Repo, release.TagName)
				return nil
			}
			fmt.Printf("Updating %s from %s -> %s...\n", cmd.Repo, addon.Version, release.TagName)
		} else {
			fmt.Printf("Tracking releases (%s)...\n", release.TagName)
		}

		loc, err := github.DownloadAndExtract(release.ZipballUrl)
		if err != nil {
			return err
		}

		// Make sure to pass the full repo name to the Addon.
		m.AddRelease(loc, cmd.Repo, release.TagName)
	}

	manifest.SaveManifest(m) // Make sure to save.
	fmt.Println("Addon installed successfully!")

	return nil
}

// Uninstall

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

// Update

type UpdateCmd struct {
	Name string `short:"n" default:"all" help:"Specific addon to update."`
}

func (cmd *UpdateCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	if cmd.Name == "all" {
		fmt.Println("Attempting to update all installed addons...")
	} else {
		fmt.Printf("Attempting to update %s\n", cmd.Name)
	}

	return nil
}

// Sync

type SyncCmd struct{}

func (cmd *SyncCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	fmt.Println("Looking for untracked addons...")
	return nil
}

// Check

type CheckCmd struct {
	Json bool `short:"j" help:"Return a structured JSON object instead of CLI output."`
}

func (cmd *CheckCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	fmt.Println("Checking for updates...")

	return nil
}

// CLI

var cli struct {
	Init      InitCmd      `cmd:"" help:"Initialize a new addons.json file."`
	Install   InstallCmd   `cmd:"" help:"Install a new addon from GitHub."`
	Uninstall UninstallCmd `cmd:"" help:"Uninstall an addon from the project."`
	Update    UpdateCmd    `cmd:"" help:"Check for updates for all installed addons."`
	Sync      SyncCmd      `cmd:"" help:"Synchronize any untracked addons in the project."`
	Check     CheckCmd     `cmd:"" help:"Check for updates without directly installing them."`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name("gau"),
		kong.Description("A CLI tool to manage Godot addons from GitHub."),
		kong.UsageOnError(),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
