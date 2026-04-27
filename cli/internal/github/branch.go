package github

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", owner, repo, branch)

	// Create the request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to reach GitHub: %v", err)
	}
	defer resp.Body.Close()

	// Handle common HTTP codes
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Branch '%s' not found on repository %s/%s.", branch, owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned an error: %s", resp.Status)
	}

	var branchData GitHubBranch
	if err := json.NewDecoder(resp.Body).Decode(&branchData); err != nil {
		return nil, fmt.Errorf("Failed to parse GitHub response: %v", err)
	}
	branchData.owner = owner
	branchData.repo = repo

	return &branchData, nil
}
