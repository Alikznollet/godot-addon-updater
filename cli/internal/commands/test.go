package commands

import (
	"github.com/alikznollet/godot-wisp/cli/internal/git"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp init' command.
//

// Initialization Command structure used by Kong.
type TestCmd struct{}

// Code Ran by the initialization.
func (cmd *TestCmd) Run() error {
	tag := git.GetLatestTag("https://github.com/Alikznollet/Godot-Steam-Lobby.git")
	commit := git.GetLatestCommitForBranch("https://github.com/Alikznollet/Godot-Steam-Lobby.git", "main")

	util.Success(tag)
	util.Success(commit)

	return nil
}
