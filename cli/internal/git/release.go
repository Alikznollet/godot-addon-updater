package git

import (
	"fmt"

	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// The Tag name of a release and the url
// to the zipball that could be downloaded.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// Returns a GitHub Release.
func GetRelease(repoInfo RepoInfo, version string) (*GitHubRelease, error) {
	url := repoInfo.BuildRepoUrl()

	util.Info("Fetching '%s' release info for %s...", version, url)

	var v string
	if version == "latest" {
		v = GetLatestTag(url)
	} else {
		v = version
	}

	if v == "" {
		return nil, fmt.Errorf("invalid version tag")
	}

	var release GitHubRelease
	release.TagName = v

	util.Success("Found '%s'", release.GetVersion())

	return &release, nil
}
