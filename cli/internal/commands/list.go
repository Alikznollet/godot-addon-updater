package commands

import (
	"os"
	"text/tabwriter"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp list' command.
//

type ListCmd struct {
	RequiresManifestCmd
}

func (cmd *ListCmd) Run() error {
	// Also grab all of the folder names from addons.
	statuses, err := cmd.Manifest.CompareWithDisk()
	if err != nil {
		return err
	}

	// If no addons print a message and exit.
	if len(statuses) == 0 {
		util.Warn("There are currently no addons installed in this project.")
		return nil
	}

	util.Info("Installed Addons:")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	for _, s := range statuses {
		util.PrintListItem(w, s.Name, s.Status, s.Details)
	}
	w.Flush()

	return nil
}
