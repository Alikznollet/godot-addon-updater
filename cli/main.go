package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// -- Command Structs -- //

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

// Init

type InitCmd struct {
	Force bool `short:"f" help:"Overwrites existing addons.json."`
}

func (cmd *InitCmd) Run() error {
	fmt.Println("Initializing addons.json...")
	fmt.Printf("Force Overwrite: %v\n", cmd.Force)

	return nil
}

// CLI

var cli struct {
	Install InstallCmd `cmd:"" help:"Install a new addon from GitHub."`
	Update  UpdateCmd  `cmd:"" help:"Check for updates for all installed addons."`
	Init    InitCmd    `cmd:"" help:"Initialize a new addons.json file."`
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
