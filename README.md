# bmgr

A terminal keybinding manager. Store, browse, and search keybindings across all your tools using [fzf](https://github.com/junegunn/fzf).

## Install

### Homebrew

```
brew install davidpanter/tap/bmgr
```

### From source

```
go install github.com/davidpanter/bmgr@latest
```

### Build from source

```
git clone https://github.com/davidpanter/bmgr.git
cd bmgr
go build -o bmgr .
```

## Dependencies

- [fzf](https://github.com/junegunn/fzf) (required for browsing)
- [fuzzel](https://codeberg.org/dnkl/fuzzel) (required for `--gui`)

## Usage

```
bmgr                        # Browse all bindings with fzf
bmgr --app tmux             # Filter to a specific app
bmgr add                    # Add a binding interactively
bmgr edit                   # Select and edit a binding in $EDITOR
bmgr remove                 # Select and remove a binding
bmgr import-keyb [file]     # Import from a keyb YAML file
```

### GUI / floating popup

```
bmgr --gui                  # Open a floating popup window
bmgr --gui --app tmux       # Popup filtered to a specific app
bmgr --gui --theme dracula  # Popup with a specific color theme
```

The `--gui` / `-g` flag opens bmgr outside the terminal using [fuzzel](https://codeberg.org/dnkl/fuzzel) in dmenu mode for a native, decoration-free Wayland popup. fuzzel must be installed and in PATH.

#### Themes

The `--theme` flag sets the fuzzel color scheme (default: `cosmic`):

| Theme     | Description                          |
|-----------|--------------------------------------|
| `cosmic`  | System76 COSMIC dark palette         |
| `dracula` | Dracula color scheme                 |
| `nord`    | Nord color scheme                    |

Themes are passed directly as fuzzel color arguments.

### Non-interactive (scripting)

```bash
# Add a binding
bmgr add --json '{"app":"tmux","description":"Split horizontal","keys":["prefix + -"]}'

# Update by ID
bmgr update abc123ef --json '{"app":"tmux","description":"Split horizontal","keys":["prefix + -"]}'
```

### Data format

Bindings are stored as JSON files in `~/.config/bmgr/` (or `$XDG_CONFIG_HOME/bmgr/`). All `*.json` files in the directory are loaded, so you can organize bindings into separate files (e.g. `tmux.json`, `vim.json`). New bindings are saved to `bindings.json` by default.

Each binding has:

| Field         | Type     | Required |
|---------------|----------|----------|
| `id`          | string   | auto     |
| `app`         | string   | yes      |
| `description` | string   | yes      |
| `keys`        | string[] | no       |
| `alternates`  | string[] | no       |
| `tags`        | string[] | no       |
| `notes`       | string   | no       |

## Editor and shell integration

See the [`contrib/`](contrib/) directory for ready-to-use snippets.

### tmux

Add to your `~/.tmux.conf`:

```tmux
# Browse all bindings (prefix + k)
bind k display-popup -E -w 90% -h 80% -T " Keybindings " "bmgr"

# Browse tmux bindings only (prefix + K)
bind K display-popup -E -w 90% -h 80% -T " Keybindings " "bmgr --app tmux"
```

### Neovim

Copy `contrib/bmgr.lua` to `~/.config/nvim/plugin/bmgr.lua`, or add the snippet to your `init.lua`. Provides:

- `:Bmgr [app]` — browse bindings (optionally filtered by app)
- `<leader>k` — browse all bindings in a floating terminal

## License

[BSD 2-Clause](LICENSE)
