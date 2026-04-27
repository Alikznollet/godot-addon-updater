package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// The Tag name of a release and the url
// to the zipball that could be downloaded.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	ZipballUrl string `json:"zipball_url"`
}

// Returns a GitHub Release.
func GetRelease(owner string, repo string, version string) (*GitHubRelease, error) {
	var url string
	if version == "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/%s", owner, repo, version)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, version)
	}

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("Failed to make request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	err = json.NewDecoder((resp.Body)).Decode(&release)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse JSON: %v", err)
	}

	return &release, nil
}
