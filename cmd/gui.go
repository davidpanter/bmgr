package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davidpanter/bmgr/store"
)

var guiFlag bool
var guiTheme string

type fuzzelTheme struct {
	bg            string // background
	text          string // text
	match         string // matched substring
	selection     string // selection background
	selectionText string // selection text
	selectionMatch string // matched substring in selection
	border        string // border color
}

var themes = map[string]fuzzelTheme{
	// Derived from ~/.config/cosmic/com.system76.CosmicTheme.Dark
	// bg=primary.base, text=primary.on, selection=primary.component.selected,
	// selText=primary.on, accent=accent.base used for matches/border
	"cosmic": {
		bg:             "272727ff", // primary.base
		text:           "f8f8f8ff", // primary.on
		match:          "63d0dfff", // accent.base
		selection:      "4a4a4aff", // primary.component.selected
		selectionText:  "f8f8f8ff", // primary.on
		selectionMatch: "63d0dfff", // accent.base
		border:         "63d0dfff", // accent.base
	},
	"dracula": {
		bg:             "282a36ff",
		text:           "f8f8f2ff",
		match:          "50fa7bff",
		selection:      "44475aff",
		selectionText:  "f8f8f2ff",
		selectionMatch: "50fa7bff",
		border:         "bd93f9ff",
	},
	"nord": {
		bg:             "2e3440ff",
		text:           "eceff4ff",
		match:          "88c0d0ff",
		selection:      "4c566aff",
		selectionText:  "eceff4ff",
		selectionMatch: "88c0d0ff",
		border:         "81a1c1ff",
	},
}

// spawnGui runs the GUI popup. Tries fuzzel (native Wayland, no decorations)
// first, then falls back to spawning a terminal with fzf.
func spawnGui() error {
	if _, err := exec.LookPath("fuzzel"); err == nil {
		return runFuzzel()
	}
	return spawnTerminal()
}

// runFuzzel pipes bindings into fuzzel --dmenu and displays a native popup.
func runFuzzel() error {
	s, err := store.Load()
	if err != nil {
		return err
	}

	bindings := s.Bindings
	if appFilter != "" {
		filter := strings.ToLower(appFilter)
		filtered := bindings[:0]
		for _, b := range bindings {
			if strings.Contains(strings.ToLower(b.App), filter) {
				filtered = append(filtered, b)
			}
		}
		bindings = filtered
	}

	if len(bindings) == 0 {
		return nil
	}

	lines := make([]string, len(bindings))
	for i, b := range bindings {
		keys := strings.Join(b.Keys, ", ")
		lines[i] = fmt.Sprintf("%-15s  %-45s  %s", b.App, b.Description, keys)
	}

	args := []string{
		"--dmenu",
		"--prompt=bmgr  ",
		"--lines=30",
		"--width=100",
	}

	if t, ok := themes[guiTheme]; ok {
		args = append(args,
			"--background-color="+t.bg,
			"--text-color="+t.text,
			"--match-color="+t.match,
			"--selection-color="+t.selection,
			"--selection-text-color="+t.selectionText,
			"--selection-match-color="+t.selectionMatch,
			"--border-color="+t.border,
		)
	}

	cmd := exec.Command("fuzzel", args...)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil // user cancelled
		}
		return err
	}
	return nil
}

// spawnTerminal re-launches bmgr in a floating terminal window, passing through
// all args except --gui itself.
func spawnTerminal() error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	var args []string
	for _, a := range os.Args[1:] {
		if a == "--gui" || a == "-g" {
			continue
		}
		args = append(args, a)
	}

	terminals := []struct {
		bin  string
		argv func(self string, args []string) []string
	}{
		{"cosmic-term", func(self string, args []string) []string {
			return append([]string{"--"}, append([]string{self}, args...)...)
		}},
		{"foot", func(self string, args []string) []string {
			return append([]string{"--app-id=bmgr-popup", "--override=csd.preferred=none", "--"}, append([]string{self}, args...)...)
		}},
		{"kitty", func(self string, args []string) []string {
			return append([]string{"--class=bmgr-popup", "--"}, append([]string{self}, args...)...)
		}},
		{"alacritty", func(self string, args []string) []string {
			return append([]string{"--class", "bmgr-popup", "-e"}, append([]string{self}, args...)...)
		}},
		{"xterm", func(self string, args []string) []string {
			return append([]string{"-e"}, append([]string{self}, args...)...)
		}},
	}

	for _, t := range terminals {
		path, err := exec.LookPath(t.bin)
		if err != nil {
			continue
		}
		cmd := exec.Command(path, t.argv(self, args)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return fmt.Errorf("no supported terminal emulator found (tried: cosmic-term, foot, kitty, alacritty, xterm)")
}
