package commands

import (
	"encoding/json"
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
)

//
// This file holds the 'wisp check' command.
//

type CheckCmd struct {
	RequiresManifestCmd
	Json bool `short:"j" help:"Return a structured JSON object instead of CLI output."`
}

func (cmd *CheckCmd) Run() error {
	fmt.Println("Checking for updates...")

	var outdated []manifest.OutdatedAddon

	for folderName, addon := range cmd.Manifest.Addons {
		isUpToDate, ref, err := cmd.Manifest.CheckAddon(addon.Repo)
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
