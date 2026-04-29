package godot

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const ProjectFile = "project.godot"

// EnableAddon safely injects an addon into the project.godot file.
func EnableAddon(addonFolder string) error {
	file, err := os.Open(ProjectFile)
	if err != nil {
		return fmt.Errorf("could not open project.godot: %v", err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close() // Close early so we can overwrite it later

	targetPath := fmt.Sprintf(`"res://addons/%s/plugin.cfg"`, addonFolder)
	inEditorPlugins := false
	foundEnabled := false
	modified := false

	// Scan line by line
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check which section we are in
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if trimmed == "[editor_plugins]" {
				inEditorPlugins = true
			} else if inEditorPlugins && !foundEnabled {
				// We left the plugins section without finding 'enabled'. Inject it here!
				newLine := fmt.Sprintf("enabled=PackedStringArray(%s)", targetPath)
				lines = append(lines[:i], append([]string{newLine, ""}, lines[i:]...)...)
				modified = true
				break
			} else {
				inEditorPlugins = false
			}
		}

		// If we are in the right section and find the array
		if inEditorPlugins && strings.HasPrefix(trimmed, "enabled=PackedStringArray(") {
			foundEnabled = true

			// Prevent duplicates
			if strings.Contains(trimmed, targetPath) {
				return nil
			}

			// Extract existing contents between the parentheses
			start := strings.Index(trimmed, "(") + 1
			end := strings.LastIndex(trimmed, ")")
			contents := strings.TrimSpace(trimmed[start:end])

			// Splice in the new plugin
			if contents == "" {
				lines[i] = fmt.Sprintf("enabled=PackedStringArray(%s)", targetPath)
			} else {
				lines[i] = fmt.Sprintf("enabled=PackedStringArray(%s, %s)", contents, targetPath)
			}

			modified = true
			break
		}
	}

	// 3. Edge Cases (File ended before we could inject)
	if !modified {
		if inEditorPlugins && !foundEnabled {
			// Reached end of file while inside [editor_plugins]
			lines = append(lines, fmt.Sprintf("enabled=PackedStringArray(%s)", targetPath))
		} else if !foundEnabled {
			// [editor_plugins] didn't exist in the file at all!
			lines = append(lines, "", "[editor_plugins]", "", fmt.Sprintf("enabled=PackedStringArray(%s)", targetPath))
		}
	}

	// Write the safely modified lines back to disk
	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(ProjectFile, []byte(output), 0644)
}

// DisableAddon safely removes an addon from the project.godot file.
func DisableAddon(addonFolder string) error {
	file, err := os.Open(ProjectFile)
	if err != nil {
		return nil // If project.godot doesn't exist, we don't care!
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	targetPath := fmt.Sprintf(`"res://addons/%s/plugin.cfg"`, addonFolder)
	inEditorPlugins := false
	modified := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			inEditorPlugins = (trimmed == "[editor_plugins]")
		}

		if inEditorPlugins && strings.HasPrefix(trimmed, "enabled=PackedStringArray(") {
			// Extract the contents between the parentheses
			start := strings.Index(trimmed, "(") + 1
			end := strings.LastIndex(trimmed, ")")
			contents := strings.TrimSpace(trimmed[start:end])

			// Split by comma and rebuild the array without our target
			parts := strings.Split(contents, ",")
			var newParts []string
			for _, p := range parts {
				cleanP := strings.TrimSpace(p)
				if cleanP != targetPath && cleanP != "" {
					newParts = append(newParts, cleanP)
				}
			}

			// Replace the line
			lines[i] = fmt.Sprintf("enabled=PackedStringArray(%s)", strings.Join(newParts, ", "))
			modified = true
			break
		}
	}

	if !modified {
		return nil // It wasn't enabled in the first place, do nothing!
	}

	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(ProjectFile, []byte(output), 0644)
}
