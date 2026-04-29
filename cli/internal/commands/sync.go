package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp sync' command.
//

type SyncCmd struct {
	RequiresManifestCmd
}

func (cmd *SyncCmd) Run() error {
	folderNames, err := manifest.GetAddonFolderContents()
	if err != nil {
		return err
	}

	// If no addons print a message and exit.
	if len(folderNames) == 0 {
		util.Info("There are currently no addons installed or tracked in this project.")
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
	for folderName := range cmd.Manifest.Addons {
		if !physicalFolders[folderName] {
			util.Warn("Folder `%s` no longer exists. Removing from tracking.", folderName)
			delete(cmd.Manifest.Addons, folderName)
			manifestChanged = true
		}
	}

	// Now we Prompt for unknowns.

	unresolvedCount := 0

	for _, folderName := range folderNames {
		if _, exists := cmd.Manifest.Addons[folderName]; exists {
			continue // Already tracked
		}

		// Delegates all ui to helpers
		changed := cmd.handleUnknown(folderName)
		if changed {
			manifestChanged = true
		} else {
			unresolvedCount++
		}
	}

	// Save changes to the manifest if a change happened
	if manifestChanged {
		if err := manifest.SaveManifest(cmd.Manifest); err != nil {
			return err
		}
		util.Success("Sync complete! '%s' has been updated.", manifest.ManifestName)

		if unresolvedCount > 0 {
			util.Warn("%d folder(s) still remain untracked.", unresolvedCount)
		}
	} else {
		if unresolvedCount > 0 {
			util.Warn("Sync finished. %d folder(s) still remain untracked.", unresolvedCount)
		} else {
			util.Success("Everything is already perfectly in sync!")
		}
	}

	return nil
}

// Displays the menu for an unknown folder
// Returns true if the manifest was modified.
func (cmd *SyncCmd) handleUnknown(folderName string) bool {
	util.Info("Found unknown addon folder: %s", util.Cyan(folderName))

	fmt.Println()

	// Print the menu cleanly.
	util.Info("What do you want to do with the unknown folder?")
	fmt.Printf("  [%s] Link to a GitHub repository\n", util.Cyan("1"))
	fmt.Printf("  [%s] Mark as Local (ignore in future syncs)\n", util.Cyan("2"))
	fmt.Printf("  [%s] Skip for now\n", util.Cyan("3"))

	fmt.Println()

	choice := util.Prompt("3", "Choice (1/2/3)")

	switch choice {
	case "3":
		util.Info("Skipping...")
		return false
	case "2":
		cmd.Manifest.Addons[folderName] = manifest.Addon{Untracked: true}
		util.Success("Marked '%s' as a local/untracked addon.", folderName)
		return true
	case "1":
		return cmd.linkToGitHub(folderName)
	default:
		util.Warn("Invalid choice. Skipping...")
		return false
	}
}

// Displays the menu for linking to github and handles the fresh installation
// Returns true if the manifest was modified.
func (cmd *SyncCmd) linkToGitHub(folderName string) bool {
	repo := util.Prompt("", "Enter GitHub repository (e.g. ramokz/phantom-camera)")
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		util.Warn("Invalid format. Must be 'owner/repo'. Skipping...")
		return false
	}

	// Ask for the tracking type
	typeChoice := util.Prompt("1", "Track [%s] Releases or [%s] a specific Branch? (1/2)", util.Cyan("1"), util.Cyan("2"))
	isBranch := (typeChoice == "2")

	// Ask for the specifics
	var target string
	if isBranch {
		target = util.Prompt("main", "Enter branch name")
		util.Info("Verifying branch '%s'...", target)
	} else {
		target = util.Prompt("latest", "Enter release tag")
		util.Info("Verifying release '%s'...", target)
	}

	// Fetch from GH
	ref, err := github.GetAddonRef(parts[0], parts[1], target, isBranch)
	if err != nil {
		util.Error("Could not verify with GitHub: %v", err)
		return false
	}

	// Fresh install logic
	if util.Confirm(false, "Fresh install files from GitHub? (Deletes current folder)") {
		addonPath := filepath.Join("addons", folderName)
		if err := os.RemoveAll(addonPath); err != nil {
			util.Error("Could not remove old folder: %v", err)
			return false
		}

		loc, err := github.DownloadAndExtract(ref.GetZipballUrl())
		if err != nil {
			util.Error("Could not perform fresh install: %v", err)
			return false
		}
		folderName = loc // Update folder	name to the newly extracted one.
	}

	// Map it to a manifest
	if isBranch {
		cmd.Manifest.AddBranch(folderName, repo, target, ref.GetVersion())
		util.Success("Linked '%s' to branch '%s'!", folderName, target)
	} else {
		cmd.Manifest.AddRelease(folderName, repo, ref.GetVersion())
		util.Success("Linked '%s' to release '%s'!", folderName, ref.GetVersion())
	}

	return true
}
