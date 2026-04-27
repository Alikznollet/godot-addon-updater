package github

import "fmt"

type AddonRef interface {
	GetVersion() string
	GetZipballUrl() string
}

func (r *GitHubRelease) GetVersion() string {
	return r.TagName
}

func (r *GitHubRelease) GetZipballUrl() string {
	return r.ZipballUrl
}

func (b *GitHubBranch) GetVersion() string {
	return b.Commit.Sha
}

func (b *GitHubBranch) GetZipballUrl() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", b.owner, b.repo, b.Commit.Sha)
}

func GetAddonRef(owner string, repo string, target string, isBranch bool) (AddonRef, error) {
	if isBranch {
		return GetBranch(owner, repo, target)
	}
	return GetRelease(owner, repo, target)
}
