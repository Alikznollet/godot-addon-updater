package manifest

// Enum used as type of Addon.
type AddonType int

const (
	Release AddonType = iota
	Branch
)

// A single addon struct with JSON support.
// This struct is written to addons.json.
type Addon struct {
	Repo      string    `json:"repo"`
	Type      AddonType `json:"type"`
	Version   string    `json:"version"`
	Untracked bool      `json:"untracked,omitempty"`
	Commit    string    `json:"commit,omitempty"` // Only used when tracking a branch.
}

// The complete list of addons mapping their repo names
// to their respective Addon structs.
type AddonManifest struct {
	Addons map[string]Addon `json:"addons"`
}

// Adds an addon to the struct.
// Should only be called after installing the addon
// succeeds.
func (m *AddonManifest) AddAddon(repo string, version string) {
	// We don't have to check if the map exists because that was
	// done when the object was created.
	m.Addons[repo] = Addon{
		Version: version,
	}
}

// Removes an addon from the struct.
// Will silently fail if the addon wasn't installed in the first place.
func (m *AddonManifest) RemoveAddon(repo string) {
	delete(m.Addons, repo)
}
