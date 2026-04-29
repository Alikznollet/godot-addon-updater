package manifest

import (
	"fmt"
	"strings"

	"github.com/alikznollet/godot-wisp/cli/internal/github"
	"github.com/alikznollet/godot-wisp/cli/internal/util"
)

// Enum used as type of Addon.
type AddonType string

const (
	Release AddonType = "release"
	Branch  AddonType = "branch"
)

// This is used as a way to pass data to the editor plugin.
type OutdatedAddon struct {
	Folder  string    `json:"folder"`
	Repo    string    `json:"repo"`
	Current string    `json:"current_version"`
	Latest  string    `json:"latest_version"`
	Branch  string    `json:"branch,omitempty"`
	Type    AddonType `json:"addon_type"`
}

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
		// Silently exit if it wasn't installed in the first place.
		return nil
	}

	if !keep {
		// Removes all files related to this addon.
		util.Warn("removing all files associated to %s", repo)
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

// Returns whether a repository is up to date or not.
func (m *AddonManifest) CheckAddon(repo string) (bool, github.AddonRef, error) {
	_, addon, isTracked := m.FindByRepo(repo)

	if !isTracked {
		return false, nil, fmt.Errorf("%s is not tracked in the current project", repo)
	}

	// Split the repo name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false, nil, fmt.Errorf("invalid repository format. Must be 'owner/repo'")
	}

	o := parts[0]
	r := parts[1]

	var ref github.AddonRef
	var err error

	switch addon.Type {
	case Branch:
		ref, err = github.GetAddonRef(o, r, addon.Version, true)
		if err != nil {
			return false, ref, err
		}

		// Check whether the commit hashes are the same.
		// If not then the ref can be used to pull the updated addon.
		if ref.GetVersion() == addon.Commit {
			return true, ref, nil
		} else {
			return false, ref, nil
		}
	case Release:
		ref, err = github.GetAddonRef(o, r, "latest", false)
		if err != nil {
			return false, ref, err
		}

		// Check whether the release is still up to date.
		if ref.GetVersion() == addon.Version {
			return true, ref, nil
		} else {
			return false, ref, nil
		}
	}

	return true, ref, nil
}
