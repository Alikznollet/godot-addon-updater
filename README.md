> [!WARNING]  
> This repository is a WIP and should not be used until an official release

# godot-addon-updater
A completely optional addon updater for Godot. Includes a CLI to manage and install addons and an Editor plugin that notifies you of any updates.

# Installation

## Windows (via Scoop)
If you use [Scoop](https://scoop.sh/), you can add the custom bucket and install `gau` instantly:
```powershell
# Add the custom bucket
scoop bucket add alikznollet https://github.com/alikznollet/scoop-bucket

# Install the CLI
scoop install gau
```

## Linux & macOS (via Homebrew)
If you use [Homebrew](https://brew.sh/), you can tap the repository and install `gau` via the terminal:
```bash
# Tap the repository
brew tap alikznollet/homebrew-tap

# Install the CLI
brew install gau
```

---

## Updating 
Because `gau` is installed via your package manager, keeping it up to date is completely frictionless.

**Windows:**
```powershell
scoop update gau
```

**Linux/macOS:**
```bash
brew upgrade gau
```
