package github

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// The latest commit on a branch.
// Contains the url to the zipball.
type GitHubBranch struct {
	Name   string       `json:"name"`
	Commit GitHubCommit `json:"commit"`
	owner  string
	repo   string
}

type GitHubCommit struct {
	Sha string `json:"sha"`
}

// Fetches the latest commit from the branch specified from the repo specified.
func GetBranch(owner string, repo string, branch string) (*GitHubBranch, error) {
	util.Info("fetching latest '%s' branch info for %s/%s", branch, owner, repo)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", owner, repo, branch)

	// Create the request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach GitHub: %v", err)
	}
	defer resp.Body.Close()

	// Handle common HTTP codes
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("branch '%s' not found on repository %s/%s", branch, owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var branchData GitHubBranch
	if err := json.NewDecoder(resp.Body).Decode(&branchData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	branchData.owner = owner
	branchData.repo = repo

	util.Success("found '%s' on branch '%s'", branchData.GetVersion(), branch)

	return &branchData, nil
}
