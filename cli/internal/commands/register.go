package commands

import (
	"github.com/alikznollet/godot-wisp/cli/internal/auth"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

//
// This file holds the 'wisp register' command.
//

// Register Command structure used by Kong.
type RegisterCmd struct {
	Host  string `arg:"" name:"host" help:"The hostname to associate with the Personal Access Token (e.g. github.com)."`
	Token string `arg:"" name:"token" help:"The Personal Access Token to to save to the os keyring."`
}

// Code Ran by the register command.
func (cmd *RegisterCmd) Run() error {
	util.Info("Binding token to '%s'...", cmd.Host)

	// Check if the user provided a token
	if cmd.Token == "" {
		util.Error("Please provide a token.")
		return nil
	}

	err := auth.StoreToken(cmd.Host, cmd.Token)
	if err != nil {
		return err
	}

	util.Success("Token for '%s' securely saved to your OS keychain.", cmd.Host)

	return nil
}
