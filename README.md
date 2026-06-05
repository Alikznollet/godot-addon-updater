# 🌀 Wisp
Wisp is a lightweight, **completely optional** addon manager for Godot. It provides a streamlined way to install, update, and track your project's dependencies via Git, while maintaining the standard Godot workflow where addons live directly in your `res://addons/` folder.

### The Wisp Philosophy
Wisp is a **convenience tool**, not a requirement. In Godot, addons are typically committed directly to your version control. Wisp respects this:
- **No Lock-in:** If you stop using Wisp, your project continues to work perfectly. Your addons remain in `res://addons/` just like any other folder.
- **Better than Submodules:** Most game developers find Git submodules cumbersome. Wisp acts as a "lite" replacement—using `git sparse-checkout` to pull only the files you need without the overhead of nested repository management.
- **Convenience First:** Wisp automates the tedious parts—finding the right folder in a repo, handling checkouts, and manually enabling plugins in `project.godot`.
- **Transparency:** All tracking happens in a human-readable `wisp.json` file.

---

## 🎯 Core Features
- **Smart Extraction:** Uses `git sparse-checkout` to pull only the `addons/` folder from a repository, even if it's nested deep within a project.
- **Auto-Injection:** Automatically enables plugins in your `project.godot` file upon installation, saving you a trip to the Project Settings.
- **Bi-directional Sync:** The `sync` command ensures your `wisp.json` and physical disk state match—discovering manually added addons and pruning stale entries.
- **Studio Ready:** Built for professional environments with native support for SSH, private repositories, and internal Git instances (GitLab, Gitea, etc.).
- **Editor Integration:** A companion plugin that allows you to check for and apply updates without leaving the engine.

---

## ⬇️ Installation

### 1. Install the CLI

**Windows (Scoop):**
```powershell
# Add the bucket and install
scoop bucket add alikznollet https://github.com/alikznollet/scoop-bucket
scoop install wisp
```

**Linux & macOS (Homebrew):**
```bash
# Tap the repository and install
brew tap alikznollet/homebrew-tap
brew install wisp
```

**From Source:**
Requires [Go 1.21+](https://go.dev/).
```bash
go install github.com/alikznollet/godot-wisp/cli@latest
```

### 2. Project Setup
Navigate to your Godot project root and initialize Wisp:
```bash
wisp init
```
This will create a `wisp.json` file. You will also be asked if you'd like to install the **Wisp Editor Plugin**—it is highly recommended for the best experience.

---

## 🔎 CLI Documentation

### `wisp install <repo>`
Installs an addon and tracks it in `wisp.json`.
- `<repo>`: Can be `owner/repo`, a full URL, or an SSH path.
- `--tag`, `-t`: Install a specific version/tag (e.g., `v1.2.0`).
- `--branch`, `-b`: Track a specific branch (e.g., `main`) to receive the latest commits.

### `wisp update [addons...]`
Checks for and applies updates. If no addons are specified, it checks all tracked dependencies.
- `--yes`, `-y`: Automatically apply all found updates without prompting.

### `wisp sync`
Maintains the integrity between your manifest and your disk:
- **Prune:** Removes entries from `wisp.json` for folders that no longer exist in `res://addons/`.
- **Discover:** Identifies untracked folders in `res://addons/` and lets you link them to a repo, mark them as local (ignored), or skip them.

### `wisp uninstall <repo>`
Stops tracking an addon. It will prompt you if you'd like to also delete the files and disable the plugin in Godot.

### `wisp list`
Displays a formatted list of all tracked addons, their sources, and their current version/branch status.

### `wisp check`
A read-only check for updates.
- `--json`, `-j`: Returns a machine-readable JSON object (used by the Editor Plugin).

---

## 🔒 Studio & Internal Use

Wisp excels in professional studio environments where internal tools are shared across projects. Because Wisp uses your system's `git` binary, it seamlessly uses your existing SSH keys and credentials.

### Private Repositories
```bash
# Install an internal tool via SSH
wisp install git@git.yourstudio.com:tools/lighting-system.git

# Track a specific development branch
wisp install studio/core-lib --branch develop
```

### Team Workflow
We recommend committing both `wisp.json` and your `res://addons/` folder to your repository. This ensures that:
1. The project is always "ready to run" for anyone who clones it.
2. Wisp serves as the source of truth for *where* those addons came from and when they should be updated.

---

## 🧩 Editor Plugin

The Wisp Editor Plugin adds a "Sync" icon to the top-right toolbar in the Godot Editor for easy access to updates.

- **Check for Updates:** Clicking the icon initiates a check across all tracked addons using the Wisp CLI.
- **Review & Apply:** A dialog will appear showing all available updates. You can select exactly which addons you want to update.
- **Automatic Refresh:** Once the update is complete, Wisp automatically triggers a filesystem rescan so Godot reflects the new files immediately.

---

## 🗺️ Roadmap
- [ ] **Version/Commit Pinning:** Lock an addon to a specific Commit or Version for absolute stability.

---

## 🤝 Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Requirements
- **Go:** 1.21 or higher (for the CLI)
- **Godot:** 4.2 or higher (for the Editor Plugin)
- **Git:** 2.27 or higher (required for sparse-checkout)

---

## ⚖️ License
Distributed under the GNU GPL License. See `LICENSE` for more information.
