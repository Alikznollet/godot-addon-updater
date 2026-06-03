package git

import (
	"os/exec"
	"strings"
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
