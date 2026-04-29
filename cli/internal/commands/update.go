package commands

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp update' command.
//

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
