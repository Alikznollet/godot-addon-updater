package manifest

// Holds clean formatted data for the CLI
type AddonStatus struct {
	Name    string
	Status  string
	Details string
}

// Scans the /addons folder and compares with what's in the manifest.
func (m *AddonManifest) CompareWithDisk() ([]AddonStatus, error) {
	folderNames, err := GetAddonFolderContents()
	if err != nil {
		return nil, err
	}

	var statuses []AddonStatus

	for _, folderName := range folderNames {
		addon, exists := m.Addons[folderName]

		status := AddonStatus{Name: folderName}

		if !exists {
			status.Status = "Unknown"
			status.Details = "Not tracked by Wisp"
		} else if addon.Untracked {
			status.Status = "Untracked"
			status.Details = "Ignored by Wisp"
		} else {
			if addon.Type == Release {
				status.Status = "Release"
				status.Details = addon.GetCurrentVersion()
			} else {
				status.Status = "Branch: " + addon.GetCurrentBranch()
				status.Details = addon.GetCurrentVersion()
			}
		}

		// If it's tracked, prefer showing the Repo name instead of just the folder
		if exists && addon.Repo != "" {
			status.Name = addon.Repo
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
