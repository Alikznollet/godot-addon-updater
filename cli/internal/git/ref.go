package git

import "context"

type AddonRef interface {
	GetVersion() string
}

func (r *GitHubRelease) GetVersion() string {
	return r.TagName
}

func (b *GitHubBranch) GetVersion() string {
	return b.CommitHash
}

func GetAddonRef(ctx context.Context, repoInfo RepoInfo, target string, isBranch bool) (AddonRef, error) {
	if isBranch {
		return GetBranch(ctx, repoInfo, target)
	}
	return GetRelease(ctx, repoInfo, target)
}
