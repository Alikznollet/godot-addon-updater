// Copyright 2026 Alikznollet
// GNU GPL

package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/alikznollet/godot-wisp/cli/internal/commands"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// CLI entrypoint for wisp
//

// GoReleaser will inject the current version here on release.
var version = "dev"

var cli struct {
	Init      commands.InitCmd      `cmd:"" help:"Initialize a new addons.json file."`
	Install   commands.InstallCmd   `cmd:"" help:"Install a new addon from GitHub."`
	Uninstall commands.UninstallCmd `cmd:"" help:"Uninstall an addon from the project."`
	Update    commands.UpdateCmd    `cmd:"" help:"Check for updates for all installed addons."`
	Sync      commands.SyncCmd      `cmd:"" help:"Synchronize any untracked addons in the project."`
	Check     commands.CheckCmd     `cmd:"" help:"Check for updates without directly installing them."`
	List      commands.ListCmd      `cmd:"" help:"List all addons in the current project."`
	Version   kong.VersionFlag      `short:"v" help:"Print the current version and exit."`
}

func main() {
	// Build parser manually so we can format errors.
	parser, err := kong.New(
		&cli,
		kong.Name("wisp"),
		kong.Description("The lightweight way to manage your Godot addons."),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)
	if err != nil {
		// If it hits this something is fucked...
		util.Fatal("Failed to build CLI: %v", err)
	}

	// Parse flags and run hooks
	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		util.Fatal("Something went wrong: %v", err)
	}

	err = ctx.Run()

	if err != nil {
		util.Fatal("A problem occurred: %v", err)
	}
}
