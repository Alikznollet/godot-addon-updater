package github

// The latest commit on a branch.
// Contains the url to the zipball.
type GitHubBranch struct {
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	ZipballUrl string `json:"zipball_url"`
}
