package github

import (
	"github.com/alikznollet/godot-wisp/cli/internal/git"
)

type AddonRef interface {
	GetVersion() string
}

func (r *GitHubRelease) GetVersion() string {
	return r.TagName
}

func (b *GitHubBranch) GetVersion() string {
	return b.CommitHash
}

func GetAddonRef(repoInfo git.RepoInfo, target string, isBranch bool) (AddonRef, error) {
	if isBranch {
		return GetBranch(repoInfo, target)
	}
	return GetRelease(repoInfo, target)
}
