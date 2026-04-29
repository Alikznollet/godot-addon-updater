package commands

// TODO: Perform refactor!

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
)

//
// This file holds the 'wisp list' command.
//

type ListCmd struct {
	RequiresManifestCmd
}

func (cmd *ListCmd) Run() error {
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
		if addon, exists := cmd.Manifest.Addons[folderName]; exists {
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
