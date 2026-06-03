package git

import (
	"fmt"
	"strings"
)

// A reference to a repo.
type RepoInfo struct {
	Domain string `json:"domain"`
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
}

// Will parse a repository string of either format "owner/repo" and default it to "github.com" as the domain
// Or if there is a domain before it'll take whatever domain was there.
func ParseRepoString(repoString string) (RepoInfo, error) {
	// Clean up the string to prevent parsing errors
	cleanString := strings.TrimPrefix(repoString, "https://")
	cleanString = strings.TrimPrefix(cleanString, "http://")
	cleanString = strings.TrimSuffix(cleanString, "/")
	cleanString = strings.TrimSuffix(cleanString, ".git")

	// Split into parts
	parts := strings.Split(cleanString, "/")

	// Validate we have at least owner/repo
	if len(parts) < 2 {
		return RepoInfo{}, fmt.Errorf("invalid repository string '%s': must contain at least 'owner/repo'", repoString)
	}

	// 4. Extract the exact last two elements
	repo := parts[len(parts)-1]
	owner := parts[len(parts)-2]

	domain := "github.com" // Implicit Default

	if len(parts) > 2 {
		// Take everything BEFORE the last two parts and join it back together
		domain = strings.Join(parts[:len(parts)-2], "/")
	}

	return RepoInfo{
		Domain: domain,
		Owner:  owner,
		Repo:   repo,
	}, nil
}

// Build an URL for the repo based on the info object.
func (r *RepoInfo) BuildRepoUrl() string {
	url := fmt.Sprintf("https://%s/%s/%s", r.Domain, r.Owner, r.Repo)
	return url
}
