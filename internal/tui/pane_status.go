package tui

import (
	"fmt"
	"strings"

	"golang-git-graph/internal/git"

	"github.com/charmbracelet/lipgloss"
)

type StatusPane struct {
	StagedFiles   []git.FileChange
	UnstagedFiles []git.FileChange
	Cursor        int
	Offset        int
}

func NewStatusPane() *StatusPane {
	return &StatusPane{
		StagedFiles:   []git.FileChange{},
		UnstagedFiles: []git.FileChange{},
		Cursor:        0,
		Offset:        0,
	}
}

func (sp *StatusPane) UpdateChanges(changes []git.FileChange) {
	staged := []git.FileChange{}
	unstaged := []git.FileChange{}

	for _, change := range changes {
		if change.Staged {
			staged = append(staged, change)
		} else {
			unstaged = append(unstaged, change)
		}
	}

	sp.StagedFiles = staged
	sp.UnstagedFiles = unstaged

	total := len(staged) + len(unstaged)
	if sp.Cursor >= total {
		sp.Cursor = total - 1
	}
	if sp.Cursor < 0 {
		sp.Cursor = 0
	}
}

func (sp *StatusPane) Total() int {
	return len(sp.StagedFiles) + len(sp.UnstagedFiles)
}

func (sp *StatusPane) ScrollUp() {
	if sp.Cursor > 0 {
		sp.Cursor--
	}
}

func (sp *StatusPane) ScrollDown() {
	if sp.Cursor < sp.Total()-1 {
		sp.Cursor++
	}
}

func (sp *StatusPane) SelectedFile() *git.FileChange {
	totalStaged := len(sp.StagedFiles)
	if sp.Cursor < 0 || sp.Cursor >= sp.Total() {
		return nil
	}

	if sp.Cursor < totalStaged {
		return &sp.StagedFiles[sp.Cursor]
	}
	return &sp.UnstagedFiles[sp.Cursor-totalStaged]
}

func (sp *StatusPane) View(styles Styles, width int, height int, focused bool) string {
	total := sp.Total()
	if total == 0 {
		return "Working tree clean. No changes."
	}

	// Adjust offset for scrolling.
	// Since headers are printed, we estimate that headers take 2 lines each.
	// But let's keep the scrolling offset computation simple based on cursor index.
	if sp.Cursor < sp.Offset {
		sp.Offset = sp.Cursor
	} else if sp.Cursor >= sp.Offset+height-4 { // Leave room for headers
		sp.Offset = sp.Cursor - height + 5
	}
	if sp.Offset < 0 {
		sp.Offset = 0
	}

	var sb strings.Builder
	var lines []string

	// 1. Gather all staged lines
	stagedHeader := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Bold(true).Render("● Staged Changes:")
	lines = append(lines, stagedHeader)
	if len(sp.StagedFiles) == 0 {
		lines = append(lines, "  (none)")
	} else {
		for i, file := range sp.StagedFiles {
			isSelected := i == sp.Cursor
			line := sp.formatFileLine(file, isSelected, focused, styles)
			lines = append(lines, line)
		}
	}

	// Add separation
	lines = append(lines, "")

	// 2. Gather all unstaged/untracked lines
	unstagedHeader := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange)).Bold(true).Render("○ Unstaged Changes:")
	lines = append(lines, unstagedHeader)
	if len(sp.UnstagedFiles) == 0 {
		lines = append(lines, "  (none)")
	} else {
		totalStaged := len(sp.StagedFiles)
		for i, file := range sp.UnstagedFiles {
			globalIdx := totalStaged + i
			isSelected := globalIdx == sp.Cursor
			line := sp.formatFileLine(file, isSelected, focused, styles)
			lines = append(lines, line)
		}
	}

	// Slice matching the viewport offset and height
	end := sp.Offset + height
	if end > len(lines) {
		end = len(lines)
	}
	
	for i := sp.Offset; i < end; i++ {
		sb.WriteString(lines[i] + "\n")
	}

	return sb.String()
}

func (sp *StatusPane) formatFileLine(file git.FileChange, isSelected bool, focused bool, styles Styles) string {
	// Status indicator formatting
	var statusStyle lipgloss.Style
	switch file.Status {
	case "A":
		statusStyle = styles.StagedFileStyle
	case "M":
		statusStyle = styles.UnstagedFileStyle
	case "D":
		statusStyle = styles.ConflictedFileStyle // Red for deleted
	case "R":
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple))
	default:
		statusStyle = styles.UntrackedFileStyle
	}

	statusIcon := statusStyle.Render(file.Status)
	
	// Checkbox indicator
	checkbox := "[ ]"
	if file.Staged {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Render("[x]")
	}

	// File path
	pathStr := file.Path
	if file.OldPath != "" {
		pathStr = fmt.Sprintf("%s -> %s", file.OldPath, file.Path)
	}

	lineContent := fmt.Sprintf("  %s %s  %s", checkbox, statusIcon, pathStr)

	if isSelected && focused {
		return styles.SelectedLineStyle.Render(lineContent)
	}
	return styles.NormalLineStyle.Render(lineContent)
}
