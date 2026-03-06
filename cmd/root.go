package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davidpanter/bmgr/store"

	"github.com/spf13/cobra"
)

var appFilter string

var rootCmd = &cobra.Command{
	Use:   "bmgr",
	Short: "Keybinding manager — browse and manage application keybindings",
	RunE:  runList,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Browse keybindings with fzf",
	RunE:  runList,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	fzfCommands := map[string]bool{"": true, "list": true, "edit": true, "remove": true}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if guiFlag {
			if err := spawnGui(); err != nil {
				return err
			}
			os.Exit(0)
		}
		if !fzfCommands[cmd.Name()] {
			return nil
		}
		if _, err := exec.LookPath("fzf"); err != nil {
			return fmt.Errorf("fzf not found in PATH — install it with: brew install fzf  or  apt install fzf")
		}
		return nil
	}
	rootCmd.PersistentFlags().StringVar(&appFilter, "app", "", "limit display to a specific application (case-insensitive substring match)")
	rootCmd.PersistentFlags().BoolVarP(&guiFlag, "gui", "g", false, "open in a floating terminal window")
	rootCmd.PersistentFlags().StringVar(&guiTheme, "theme", "cosmic", "fuzzel color theme: cosmic, dracula, nord")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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
		if appFilter != "" {
			fmt.Printf("No bindings found for app %q.\n", appFilter)
		} else {
			fmt.Println("No bindings yet. Use 'bmgr add' to add one.")
		}
		return nil
	}

	lines := make([]string, len(bindings))
	for i, b := range bindings {
		lines[i] = formatLine(b)
	}

	_, err = runFzf(lines, []string{
		"--header=APP                  DESCRIPTION                                        KEYS",
	})
	return err
}

func runFzf(lines []string, extraArgs []string) (string, error) {
	baseArgs := []string{
		"--delimiter=\t",
		"--with-nth=2..4",
		"--layout=reverse",
		"--preview=echo {5}",
		"--preview-window=down:3:wrap",
	}
	args := append(baseArgs, extraArgs...)

	fzfCmd := exec.Command("fzf", args...)
	fzfCmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	var out bytes.Buffer
	fzfCmd.Stdout = &out
	fzfCmd.Stderr = os.Stderr

	err := fzfCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 1 || code == 130 {
				// no match or user cancelled
				return "", nil
			}
		}
		return "", fmt.Errorf("fzf error: %w", err)
	}

	return strings.TrimRight(out.String(), "\n"), nil
}

func formatLine(b store.Binding) string {
	app := sanitizeTab(b.App)
	desc := sanitizeTab(b.Description)
	keys := sanitizeTab(strings.Join(b.Keys, ", "))
	notes := sanitizeTab(b.Notes)
	return fmt.Sprintf("%s\t%-20s\t%-50s\t%s\t%s", b.ID, app, desc, keys, notes)
}

func sanitizeTab(s string) string {
	return strings.ReplaceAll(s, "\t", " ")
}
