package commands

import (
	"bufio"
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
