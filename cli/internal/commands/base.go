package commands

import (
	"github.com/alikznollet/godot-wisp/cli/internal/godot"
	"github.com/alikznollet/godot-wisp/cli/internal/manifest"
)

// A command that needs to be ran inside of a godot project.
type RequiresGodotProjectCmd struct{}

func (cmd *RequiresGodotProjectCmd) AfterApply() error {
	return godot.EnsureGodotProject()
}

// A command that needs the manifest to function.
// Chains the RequiresGodotProjectCmd.
type RequiresManifestCmd struct {
	RequiresGodotProjectCmd
	Manifest *manifest.AddonManifest `kong:"-"`
}

func (cmd *RequiresManifestCmd) AfterApply() error {
	// Chain the GodotProject call.
	err := cmd.RequiresGodotProjectCmd.AfterApply()
	if err != nil {
		return err
	}

	// Load the manifest.
	m, err := manifest.LoadManifest()
	if err != nil {
		return err
	}

	cmd.Manifest = m
	return nil
}
