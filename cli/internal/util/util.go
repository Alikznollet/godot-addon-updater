package util

import (
	"errors"
	"os"
)

func EnsureGodotProject() error {
	_, err := os.Stat("project.godot")

	// We check whether the Error is "File Does Not Exist"
	if os.IsNotExist(err) {
		return errors.New("No project.godot found. Please run this command in the root of a Godot project.")
	}

	// If a weird different error comes up we return that directly.
	if err != nil {
		return err
	}

	return nil
}
