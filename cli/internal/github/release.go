package github_release

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alikznollet/godot-addon-updater/types"
)

func getLatestRelease(owner string, repo string) (*types.GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("Failed to make request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release types.GitHubRelease
	err = json.NewDecoder((resp.Body)).Decode(&release)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse JSON: %v", err)
	}

	return &release, nil
}
