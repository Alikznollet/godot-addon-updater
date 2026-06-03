package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// Get the latest tag from whatever url to version control is provided using git.
func GetLatestTag(repoUrl string) string {
	cmd := exec.Command("git", "ls-remote", "--tags", "--sort=-v:refname", repoUrl)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Split the output from the command.
	parts := strings.Fields(string(output))

	// Grabs the top (so most recent tag) and split by /
	tag := strings.Split(parts[1], "/")

	// "ref/tags/version" and we want the actual tag so index 2
	return tag[2]
}

// Get the latest commit from a branch.
func GetLatestCommitForBranch(repoUrl string, branch string) string {
	cmd := exec.Command("git", "ls-remote", repoUrl, branch)

	// Get the output from the command.
	outputBytes, err := cmd.Output()
	if err != nil {
		return ""
	}

	outputStr := string(outputBytes)

	// If the branch doesn't exist an empty string is returned.
	if strings.TrimSpace(outputStr) == "" {
		return ""
	}

	// [0] is the commit hash, [1] the "refs/heads/branchName"
	parts := strings.Fields(outputStr)

	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

func GitDownload(repoUrl string, version string) error {
	tempDir, err := os.MkdirTemp("", "wisp-clone-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	util.Info("Cloning '%s' into temp folder...", repoUrl)

	// Clone the repo efficiently without downloading everything.
	cloneCmd := exec.Command("git", "clone",
		"--depth", "1",
		"--filter=blob:none",
		"--sparse",
		"--branch", version,
		repoUrl,
		tempDir,
	)

	// Route output to the user
	// TODO: Replace with a loading bar?
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr

	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %v", err)
	}

	// We have to look for the specific addons folder first.
	lsCmd := exec.Command("git", "ls-tree", "-r", "--name-only", "HEAD")
	lsCmd.Dir = tempDir

	output, err := lsCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to scan repository structure: %v", err)
	}

	// Look for any path that ends or is addons.
	var targetAddonsPath string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "addons" || strings.HasSuffix(line, "/addons") {
			targetAddonsPath = line
			break
		}
	}

	if targetAddonsPath == "" {
		return fmt.Errorf("could not find an 'addons' folder anywhere in the repository")
	}

	util.Info("Found 'addons' folder at: %s\n", targetAddonsPath)

	// We can then call sparse-checkout to only get the addons folder.
	sparseCmd := exec.Command("git", "sparse-checkout", "set", targetAddonsPath)
	sparseCmd.Dir = tempDir

	if err := sparseCmd.Run(); err != nil {
		return fmt.Errorf("git sparse-checkout failed: %v", err)
	}

	sourceAddonsPath := filepath.Join(tempDir, targetAddonsPath)
	destAddonsPath := filepath.Join(".", "addons") // We're sure of being inside a Godot project

	// We know the addons path exists from the sparse-checkout check so we can just copy
	util.Info("Moving addon into you Godot project...")

	err = util.CopyDir(sourceAddonsPath, destAddonsPath)
	if err != nil {
		return fmt.Errorf("failed to copy addons folder: %v", err)
	}

	util.Success("Successfully downloaded addon!")
	return nil
}
