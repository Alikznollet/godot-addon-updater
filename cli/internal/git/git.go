package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// Get the latest tag from whatever url to version control is provided using git.
func GetLatestTag(ctx context.Context, repoUrl string) string {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--sort=-v:refname", repoUrl)
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
func GetLatestCommitForBranch(ctx context.Context, repoUrl string, branch string) string {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", repoUrl, branch)

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

func GitDownload(ctx context.Context, repoUrl string, version string) (string, error) {
	tempDir, err := os.MkdirTemp("", "wisp-clone-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	util.Info("Cloning '%s' into temp folder...", repoUrl)

	// Clone the repo efficiently without downloading everything.
	cloneCmd := exec.CommandContext(ctx, "git", "clone",
		"--depth", "1",
		"--filter=blob:none",
		"--sparse",
		"--branch", version,
		repoUrl,
		tempDir,
	)

	spinner := util.NewSpinner("Cloning repository")
	done := make(chan bool)

	// Start the animation
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				spinner.Add(1)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	err = cloneCmd.Run()

	// Stop the animation.
	done <- true
	spinner.Finish()

	if err != nil {
		return "", fmt.Errorf("git clone failed: %v", err)
	}

	// We have to look for the specific addons folder first.
	lsCmd := exec.CommandContext(ctx, "git", "ls-tree", "-r", "--name-only", "HEAD")
	lsCmd.Dir = tempDir

	output, err := lsCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to scan repository structure: %v", err)
	}

	// Look for any path that ends or is addons.
	var targetAddonsPath string
	lines := strings.SplitSeq(string(output), "\n")

	for rawLine := range lines {
		line := strings.TrimSpace(rawLine)

		if line == "" {
			continue
		}

		// Work with case-insensitive string
		lowerLine := strings.ToLower(line)

		// The addons folder is at the absolute root (e.g., "addons/plugin/script.gd")
		if strings.HasPrefix(lowerLine, "addons/") {
			// The original casing for root is usually just "addons", but let's take exactly what's there
			targetAddonsPath = line[:6] // "addons" is 6 characters
			break
		}

		// The addons folder is nested (e.g., "example_project/addons/plugin/script.gd")
		idx := strings.Index(lowerLine, "/addons/")
		if idx != -1 {
			// We slice the string to grab everything up to the end of the word "addons"
			// len("/addons") is 7. This preserves the exact casing of the original path!
			targetAddonsPath = line[:idx+7]
			break
		}
	}

	if targetAddonsPath == "" {
		return "", fmt.Errorf("could not find an 'addons' folder anywhere in the repository")
	}

	util.Info("Found 'addons' folder.")

	// We can then call sparse-checkout to only get the addons folder.
	sparseCmd := exec.CommandContext(ctx, "git", "sparse-checkout", "set", targetAddonsPath)
	sparseCmd.Dir = tempDir

	if err := sparseCmd.Run(); err != nil {
		return "", fmt.Errorf("git sparse-checkout failed: %v", err)
	}

	sourceAddonsPath := filepath.Join(tempDir, targetAddonsPath)
	destAddonsPath := filepath.Join(".", "addons") // We're sure of being inside a Godot project

	// We know the addons path exists from the sparse-checkout check so we can just copy
	util.Info("Moving addon into you Godot project...")

	loc, err := util.CopyDir(sourceAddonsPath, destAddonsPath)
	if err != nil {
		// If something went wrong copying files to the addons folder we want to remove them all again.
		os.RemoveAll(destAddonsPath)
		return "", fmt.Errorf("failed to copy addons folder: %v", err)
	}

	util.Success("Successfully downloaded addon!")
	return loc, nil
}

// Ensures git is installed and is at least 2.27
func VerifyGitRequirements(ctx context.Context) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not in system PATH")
	}

	// Check the git version
	cmd := exec.CommandContext(ctx, "git", "--version")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not execute git --version: %v", err)
	}

	// Trim the output
	versionStr := strings.TrimSpace(string(out))
	parts := strings.Fields(versionStr)

	if len(parts) < 3 {
		return fmt.Errorf("unrecognized git version output: %s", versionStr)
	}

	// Extract the version numbers
	rawVersion := parts[2]
	versionParts := strings.Split(rawVersion, ".")

	if len(versionParts) < 2 {
		return fmt.Errorf("could not parse major/minor version from: %s", rawVersion)
	}

	// Parse Major and Minor
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return fmt.Errorf("failed to parse major version: %v", err)
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return fmt.Errorf("failed to parse minor version: %v", err)
	}

	// Check the version against req
	// ! Change version req here if anything ever changes.
	if major < 2 || (major == 2 && minor < 27) {
		return fmt.Errorf("Wisp requires Git v2.27 or newer for sparse-checkout features. You are using v%d.%d", major, minor)
	}

	return nil
}
