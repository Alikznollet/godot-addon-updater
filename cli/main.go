// Copyright 2026 Alikznollet
// GNU GPL

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
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
	Yes   bool     `short:"y" help:"Automatically confirm each update without user interaction."`
}

func (cmd *UpdateCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	// Load the manifest.
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	var reposToGet []manifest.Addon
	if len(cmd.Repos) == 0 {
		reposToGet = make([]manifest.Addon, 0, len(m.Addons))
		for _, addon := range m.Addons {
			if !addon.Untracked {
				reposToGet = append(reposToGet, addon)
			}
		}
	} else {
		reposToGet = make([]manifest.Addon, 0, len(cmd.Repos))
		for _, repoName := range cmd.Repos {
			_, addon, isTracked := m.FindByRepo(repoName)
			if isTracked {
				reposToGet = append(reposToGet, addon)
			}
		}
	}

	fmt.Printf("Attempting to update %d tracked addons.\n", len(reposToGet))

	// Check and update repos that need it.
	for _, addon := range reposToGet {
		fmt.Printf("Attempting to update %s...\n", addon.Repo)

		isUpToDate, ref, err := m.CheckAddon(addon.Repo)
		if err != nil {
			fmt.Printf("Error while checking %s: %v\n", addon.Repo, err)
			continue
		}

		// Only try to update if the addon is not up to date and ref is there.
		if !isUpToDate && ref != nil {
			fmt.Printf("Update found for %s -> %s.\n", addon.Repo, ref.GetVersion())

			// Only ask for confirm if yes is not passed.
			if !cmd.Yes {
				confirmationString := fmt.Sprintf("Do you want to download and apply updates for %s?", addon.Repo)
				if !util.AskForConfirmation(confirmationString) {
					fmt.Println("Update cancelled.")
					continue
				}
			}

			fmt.Println("Applying updates...")
			loc, err := github.DownloadAndExtract(ref.GetZipballUrl())
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", addon.Repo, err)
				continue // Continue to the next update.
			}

			switch addon.Type {
			case manifest.Release:
				m.AddRelease(loc, addon.Repo, ref.GetVersion())
			case manifest.Branch:
				m.AddBranch(loc, addon.Repo, addon.Version, ref.GetVersion())
			default:
				continue
			}
		} else {
			fmt.Printf("%s is up to date.\n", addon.Repo)
		}
	}

	// Make sure to save the manifest after all of this.
	err = manifest.SaveManifest(m)
	if err != nil {
		return err
	}

	fmt.Println("Finished updating addons!")

	return nil
}

// Sync

type SyncCmd struct{}

func (cmd *SyncCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	// Load the manifest
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	// Also grab all of the folder names from addons.
	folderNames, err := manifest.GetAddonFolderContents()
	if err != nil {
		return err
	}

	// If no addons print a message and exit.
	if len(folderNames) == 0 {
		fmt.Println("There are currently no addons installed in this project.")
		return nil
	}

	manifestChanged := false

	// Firstly we PRUNE the addons.json file.

	// Quick lookup map for physical folders
	physicalFolders := make(map[string]bool)
	for _, f := range folderNames {
		physicalFolders[f] = true
	}

	// If a tracked addon is no longer in the folders remove it from addons.json
	for folderName := range m.Addons {
		if !physicalFolders[folderName] {
			fmt.Printf("Folder `%s` no longer exists. Removing the associated addon from addons.json.\n", folderName)
			delete(m.Addons, folderName)
			manifestChanged = true
		}
	}

	// Now we Prompt for unknowns.

	scanner := bufio.NewScanner(os.Stdin)

	for _, folderName := range folderNames {
		if _, exists := m.Addons[folderName]; exists {
			continue // If it exists in the manifest then we skip no matter what.
		}

		fmt.Printf("Found unknown addon folder: %s\n", folderName)
		fmt.Println("  [1] Link to a GitHub repository")
		fmt.Println("  [2] Mark as Local (ignore in future syncs)")
		fmt.Println("  [3] Skip for now")

		choice := util.AskInput("Choice (1/2/3): ", scanner)

		switch choice {
		case "3":
			fmt.Println("Skipping...")
		case "2":
			// We just mark it as untracked and leave it.
			m.Addons[folderName] = manifest.Addon{Untracked: true}
			fmt.Printf("Marked '%s' as a local/untracked addon.\n", folderName)
			manifestChanged = true
			continue
		case "1":
			// Get the repo
			repo := util.AskInput("Enter a GitHub repository (e.g ramokz/phantom-camera): ", scanner)
			parts := strings.Split(repo, "/")
			if len(parts) != 2 {
				fmt.Println("Invalid format. Skipping...")
				continue
			}

			// Get the type of tracking.
			typeChoice := util.AskInput("Do you want to track [1] Releases or [2] a specific Branch? (1/2): ", scanner)

			switch typeChoice {
			case "1":
				// If we are trying to link a release
				target := util.AskInput("Enter release tag (e.g., v1.0.0) or type 'latest': ", scanner)
				fmt.Printf("Verifying release '%s'...\n", target)

				ref, err := github.GetAddonRef(parts[0], parts[1], target, false)
				if err != nil {
					fmt.Printf("Could not verify release: %v\n", err)
					continue
				}

				// Check if the user wants to perform a fresh install.
				install := util.AskForConfirmation("Do you want to fresh install the files?")
				if install {

					// Remove the original file path (in case it doesn't match the downloaded one.)
					addonPath := filepath.Join("addons", folderName)
					if err := os.RemoveAll(addonPath); err != nil {
						fmt.Printf("Could not remove old folder before install: %v\n", err)
						continue
					}

					loc, err := github.DownloadAndExtract(ref.GetZipballUrl())
					if err != nil {
						fmt.Printf("Could not perform fresh install: %v", err)
						continue
					}
					folderName = loc
				}

				// Add the release to the manifest.
				m.AddRelease(folderName, repo, ref.GetVersion())
				fmt.Printf("Linked '%s' to release '%s'.\n", folderName, ref.GetVersion())
				manifestChanged = true
			case "2":
				// If we are trying to link a branch.
				branchName := util.AskInput("Enter branch name (e.g., main, develop): ", scanner)
				fmt.Printf("Verifying branch '%s'...\n", branchName)

				ref, err := github.GetAddonRef(parts[0], parts[1], branchName, true)
				if err != nil {
					fmt.Printf("Could not verify branch: %v\n", err)
					continue
				}

				// Check if the user wants to perform a fresh install.
				install := util.AskForConfirmation("Do you want to fresh install the files?")
				if install {

					// Remove the original file path (in case it doesn't match the downloaded one.)
					addonPath := filepath.Join("addons", folderName)
					if err := os.RemoveAll(addonPath); err != nil {
						fmt.Printf("Could not remove old folder before install: %v\n", err)
						continue
					}

					loc, err := github.DownloadAndExtract(ref.GetZipballUrl())
					if err != nil {
						fmt.Printf("Could not perform fresh install: %v", err)
						continue
					}
					folderName = loc
				}

				// Add the branch to the manifest.
				m.AddBranch(folderName, repo, branchName, ref.GetVersion())
				fmt.Printf("Linked '%s' to branch '%s'.\n", folderName, branchName)
				manifestChanged = true
			default:
				fmt.Println("Invalid choice. Skipping...")
			}
		default:
			fmt.Println("Invalid choice. Skipping...")
		}
	}

	// Save changes to the manifest.
	if manifestChanged {
		err := manifest.SaveManifest(m)
		if err != nil {
			return err
		}
		fmt.Println("Sync complete! Addons.json updated.")
	} else {
		fmt.Println("Everything is already perfectly in sync!")
	}

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

// List

type ListCmd struct{}

func (cmd *ListCmd) Run() error {
	if err := util.EnsureGodotProject(); err != nil {
		return err
	}

	// Load the manifest
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	// Also grab all of the folder names from addons.
	folderNames, err := manifest.GetAddonFolderContents()
	if err != nil {
		return err
	}

	// If no addons print a message and exit.
	if len(folderNames) == 0 {
		fmt.Println("There are currently no addons installed in this project.")
		return nil
	}

	// Loop over all of the folder names.
	for _, folderName := range folderNames {
		// If the addon exists in the manifest we can display it.
		if addon, exists := m.Addons[folderName]; exists {
			if addon.Untracked {
				fmt.Printf("%s (Status: Untracked)\n", folderName)
			} else {
				switch addon.Type {
				case manifest.Branch:
					fmt.Printf("%s (Branch: %s) (Commit: %s)\n", addon.Repo, addon.Version, addon.Commit)
				case manifest.Release:
					fmt.Printf("%s (Release: %s)\n", addon.Repo, addon.Version)
				}
			}
		} else {
			// Otherwise it means this addon is unknown.
			fmt.Printf("%s (Status: Unknown)\n", folderName)
		}
	}

	return nil
}

// CLI

// GoReleaser will inject the current version here on release.
var version = "dev"

var cli struct {
	Init      InitCmd          `cmd:"" help:"Initialize a new addons.json file."`
	Install   InstallCmd       `cmd:"" help:"Install a new addon from GitHub."`
	Uninstall UninstallCmd     `cmd:"" help:"Uninstall an addon from the project."`
	Update    UpdateCmd        `cmd:"" help:"Check for updates for all installed addons."`
	Sync      SyncCmd          `cmd:"" help:"Synchronize any untracked addons in the project."`
	Check     CheckCmd         `cmd:"" help:"Check for updates without directly installing them."`
	List      ListCmd          `cmd:"" help:"List all addons in the current project."`
	Version   kong.VersionFlag `short:"v" help:"Print the current version and exit."`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name("wisp"),
		kong.Description("The lightweight way to manage your Godot addons."),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
