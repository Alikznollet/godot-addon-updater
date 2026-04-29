package godot

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const ProjectFile = "project.godot"

// EnablePlugin safely injects an addon into the project.godot file.
func EnablePlugin(addonFolder string) error {
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

		// 2. If we are in the right section and find the array
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
