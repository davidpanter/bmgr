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

// spawnGui runs the GUI popup via fuzzel (native Wayland, no decorations).
func spawnGui() error {
	if _, err := exec.LookPath("fuzzel"); err != nil {
		return fmt.Errorf("fuzzel not found in PATH — install it with: brew install fuzzel  or  apt install fuzzel")
	}
	return runFuzzel()
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

