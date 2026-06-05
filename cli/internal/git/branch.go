package git

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// The latest commit on a branch.
// Contains the url to the zipball.
type GitHubBranch struct {
	Name       string `json:"name"`
	CommitHash string `json:"commit"`
}

type GitHubCommit struct {
	Sha string `json:"sha"`
}

// Fetches the latest commit from the branch specified from the repo specified.
func GetBranch(repoInfo RepoInfo, branch string) (*GitHubBranch, error) {
	url := repoInfo.BuildRepoUrl()

	util.Info("Fetching latest '%s' branch info for %s...", branch, url)

	commitHash := GetLatestCommitForBranch(url, branch)
	if commitHash == "" {
		return nil, fmt.Errorf("failed to fetch latest commit")
	}

	var branchData GitHubBranch
	branchData.Name = branch
	branchData.CommitHash = commitHash

	util.Success("Found '%s' on branch '%s'", branchData.GetVersion(), branch)

	return &branchData, nil
}
