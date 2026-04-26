package manifest

import "fmt"

// Enum used as type of Addon.
type AddonType string

const (
	Release AddonType = "release"
	Branch  AddonType = "branch"
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

// Adds an addon based on the branch of a repository.
// Will track the commit and the branch name.
// Should only be called after installation succeeds.
func (m *AddonManifest) AddBranch(folder string, repo string, branch string, commit string) {
	m.Addons[folder] = Addon{
		Repo:    repo,
		Type:    Branch,
		Version: branch, // This is the branch name.
		Commit:  commit,
	}
}

// Adds an addon to the struct.
// Should only be called after installing the addon
// succeeds. The name of the folder is used to index.
func (m *AddonManifest) AddRelease(folder string, repo string, version string) {
	// We don't have to check if the map exists because that was
	// done when the object was created.
	m.Addons[folder] = Addon{
		Repo:    repo,
		Type:    Release,
		Version: version,
	}
}

// Removes an addon from the struct.
// Will silently fail if the addon wasn't installed in the first place.
// Will also remove the folder the addon was installed in if prompted.
func (m *AddonManifest) RemoveAddon(repo string, keep bool) error {
	folderName, _, isTracked := m.FindByRepo(repo)
	if !isTracked {
		fmt.Printf("%s is not actively being tracked in this project.", repo)
		return nil
	}

	if !keep {
		// Removes all files related to this addon.
		fmt.Printf("Removing all files associated to %s\n", repo)
		err := deleteAddonFolder(folderName)
		if err != nil {
			return err
		}
	}

	// Removes the addon from the dictionary.
	delete(m.Addons, folderName)
	return nil
}

// Looks for an addon by their repo name.
func (m *AddonManifest) FindByRepo(repo string) (string, Addon, bool) {
	for folderName, addon := range m.Addons {
		if addon.Repo == repo {
			return folderName, addon, true
		}
	}
	return "", Addon{}, false // No Addon found.
}
