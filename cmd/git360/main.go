package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"golang-git-graph/internal/git"
	"golang-git-graph/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Optional flag to specify a directory
	dirFlag := flag.String("dir", "", "Path to the Git repository")
	flag.Parse()

	targetDir := "."
	if flag.NArg() > 0 {
		targetDir = flag.Arg(0)
	} else if *dirFlag != "" {
		targetDir = *dirFlag
	}

	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	client := git.NewClient(absPath)
	if !client.IsInsideRepo() {
		fmt.Fprintf(os.Stderr, "Error: %s is not inside a valid Git repository.\n", absPath)
		os.Exit(1)
	}

	model := tui.NewAppModel(client)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error running Git-360: %v\n", err)
		os.Exit(1)
	}
}
