// Copyright 2026 Alikznollet
// GNU GPL

package main

import (
	"encoding/json"
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
	Repos []string `arg:"" optional:"" name:"repos" help:"List of all specific addons to update."`
	Yes   bool     `short:"k" help:"Automatically confirm each update without user interaction."`
}

func (cmd *UpdateCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	// Load the manifest.
	_, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	if len(cmd.Repos) == 0 {
		fmt.Println("Attempting to update ALL addons...")
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

	// Load the manifest.
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	var outdated []manifest.OutdatedAddon

	for folderName, addon := range m.Addons {
		isUpToDate, ref, err := m.CheckAddon(addon.Repo)
		if err != nil {
			fmt.Printf("Something went wrong while checking %s, moving to next addon...: %v\n", addon.Repo, err)
			continue
		}
		if !isUpToDate && ref != nil {
			fmt.Printf("A newer version of %s is available (%s -> %s).\n", addon.Repo, addon.Version, ref.GetVersion())

			// We need to HACK a bit here...
			currentVersion := addon.Commit
			currentBranch := addon.Version
			if addon.Type == manifest.Release {
				currentVersion = addon.Version
				currentBranch = ""
			}

			// Add an OutdatedAddon object to the list for later.
			outdated = append(outdated, manifest.OutdatedAddon{
				Folder:  folderName,
				Repo:    addon.Repo,
				Current: currentVersion,
				Latest:  ref.GetVersion(),
				Type:    addon.Type,
				Branch:  currentBranch,
			})
		}
	}

	// Handle output based on request.
	if cmd.Json {
		jsonData, err := json.MarshalIndent(outdated, "", "  ")
		if err != nil {
			return err
		}
		// Print straight to stdout so the editor plugin can catch it.
		fmt.Println(string(jsonData))
		return nil
	}

	// Now for non json
	if len(outdated) == 0 {
		fmt.Println("All addons are up to date!")
		return nil
	}

	for _, o := range outdated {
		fmt.Printf("%s can be updated from %s to %s.", o.Repo, o.Current, o.Latest)
	}

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
