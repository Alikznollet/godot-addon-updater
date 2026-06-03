# Spatial Remarks


A <a href="https://godotengine.org/" target="_blank">Godot</a> addon for creating remarks in 2D and 3D scenes, useful for documentation, team communication and bug reports.

Currently, remarks can be created ingame with **F3**. This plugin also adds a dock in the editor for selecting, viewing, and editing the saved remark data. Data can be saved locally as JSON-File, HTTP support is experimental.

This plugin can be further configured via `sr_config.cfg`. This file contains some further documentation on the different options.

## Installation
1. **Install the addon** manually:
Download the project, unpack it, and copy the `addons/spatial_remarks` folder to the `addons` folder in your project.
2. **Enable the addon** via `Project Settings` -> `Plugins`. This will prompt you to restart the Godot Editor (to reload the Input Map)

## Keybindings (ingame)

Default Keybindings:

* **Create Remark** (`create_sr`) - F3
* **Toggle Remark Visibility** (`show_sr`) - F4


## License
MIT License (see LICENSE.md)
