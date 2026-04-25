package types

// A single addon struct with JSON support.
// This struct is written to addons.json.
type Addon struct {
	Version   string `json:"version"`
	Untracked bool   `json:"untracked,omitempty"`
}

// The complete list of addons mapping their repo names
// to their respective Addon structs.
type AddonManifest struct {
	Addons map[string]Addon `json:"addons"`
}

// The Tag name of a release and the url
// to the zipball that could be downloaded.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	ZipballUrl string `json:"zipball_url"`
}
