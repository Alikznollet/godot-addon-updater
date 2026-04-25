package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// -- Command Structs -- //

// Init

type InitCmd struct {
	Force bool `short:"f" help:"Overwrites existing addons.json."`
}

func (cmd *InitCmd) Run() error {
	fmt.Println("Initializing addons.json...")
	fmt.Printf("Force Overwrite: %v\n", cmd.Force)

	return nil
}

// Install

type InstallCmd struct {
	Repo    string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
	Version string `short:"v" default:"latest" help:"Specific version tag to install."`
}

func (cmd *InstallCmd) Run() error {
	fmt.Printf("Installing %s\n", cmd.Repo)
	fmt.Printf("Requested Version: %s\n", cmd.Version)

	return nil
}

// Uninstall

type UninstallCmd struct {
	Repo string `arg:"" name:"repo" help:"The GitHub repository (e.g. ramokz/phantom-camera)."`
	Keep bool   `short:"k" help:"Keep the addon files in res://addons/ but remove from tracking."`
}

func (cmd *UninstallCmd) Run() error {
	fmt.Printf("Attempting to uninstall %s\n", cmd.Repo)

	if cmd.Keep {
		fmt.Println("Removing from addons.json ONLY. Files will be kept.")
	} else {
		fmt.Println("Removing from addons.json AND deleting files from res://addons")
	}

	return nil
}

// Update

type UpdateCmd struct {
	Name string `short:"n" default:"all" help:"Specific addon to update."`
}

func (cmd *UpdateCmd) Run() error {
	if cmd.Name == "all" {
		fmt.Println("Attempting to update all installed addons...")
	} else {
		fmt.Printf("Attempting to update %s\n", cmd.Name)
	}

	return nil
}

// Sync

type SyncCmd struct{}

func (cmd *SyncCmd) Run() error {
	fmt.Println("Looking for untracked addons...")
	return nil
}

// Check

type CheckCmd struct {
	Json bool `short:"j" help:"Return a structured JSON object instead of CLI output."`
}

func (cmd *CheckCmd) Run() error {
	fmt.Println("Checking for updates...")

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
		kong.Name("godot-addon-updater"),
		kong.Description("A CLI tool to manage Godot addons from GitHub."),
		kong.UsageOnError(),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
