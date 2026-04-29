package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp check' command.
//

type CheckCmd struct {
	RequiresManifestCmd
	Json bool `short:"j" help:"Return a structured JSON object instead of CLI output."`
}

func (cmd *CheckCmd) Run() error {
	if cmd.Json {
		util.MuteUI() // Mute the UI if we ask for JSON.
	}

	// Fetch the outdated addons.
	outdated, err := manifest.FetchOutdatedAddons(cmd.Manifest)
	if err != nil {
		return err
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

	// Standard CLI
	if len(outdated) == 0 {
		util.Success("All addons are up to date!")
		return nil
	}

	util.Warn("Found %d available updates:", len(outdated))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	for _, o := range outdated {
		util.PrintListItem(w, o.Repo, "Update", fmt.Sprintf("%s -> %s", o.Current, o.Latest))
	}
	w.Flush()

	return nil
}
