package util

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Ensures the user is running the tool in a godot project.
func EnsureGodotProject() error {
	_, err := os.Stat("project.godot")

	// We check whether the Error is "File Does Not Exist"
	if os.IsNotExist(err) {
		return errors.New("no project.godot found. Please run wisp in the root of a Godot project.")
	}

	// If a weird different error comes up we return that directly.
	if err != nil {
		return err
	}

	return nil
}

// Prompts the user with a yes/no question and waits for confirmation.
func AskForConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", prompt)

		// Read until enter
		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		// Clean up the input
		input = strings.TrimSpace(strings.ToLower(input))

		// Default to no.
		if input == "" || input == "n" || input == "no" {
			return false
		}
		if input == "y" || input == "yes" {
			return true
		}

		// If anything else was typed loop back.
		fmt.Println("Please type 'y' for yes or 'n' for no.")
	}
}

// Asks for an input
func AskInput(prompt string, scanner *bufio.Scanner) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
