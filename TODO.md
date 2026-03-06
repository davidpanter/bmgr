# TODO

## Configuration file for GUI themes

Allow users to define custom themes in a config file instead of hardcoding them.

- Add support for `~/.config/bmgr/config.toml` (or similar)
- Config file can define custom named themes with the same fields as built-in themes (`bg`, `text`, `match`, `selection`, `selectionText`, `selectionMatch`, `border`)
- Config file can also set a default theme so `--theme` is not needed every time
- Merge user-defined themes with built-in themes at startup; user themes take precedence
- Update `--theme` flag description and README to mention the config file

Example config:

```toml
[gui]
default_theme = "myTheme"

[themes.myTheme]
bg             = "1e1e2eff"
text           = "cdd6f4ff"
match          = "89b4faff"
selection      = "313244ff"
selection_text = "cdd6f4ff"
selection_match = "89b4faff"
border         = "89b4faff"
```

## Command-line option for lookup

Add a `lookup` (or `search`) subcommand for non-interactive, scriptable key lookup.

- `bmgr lookup <query>` — print bindings matching query to stdout (app, description, keys)
- Support `--app` filter (already a global flag)
- Support `--format` flag: `text` (default), `json`, `tsv`
- Exit code 0 if matches found, 1 if none
- Useful for shell scripts, status bars, and editor integrations

Example usage:

```bash
bmgr lookup "split"
bmgr lookup --app tmux "split" --format json
bmgr lookup --app vim "save" --format tsv
```
