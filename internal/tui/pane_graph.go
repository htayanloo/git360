package tui

import (
	"fmt"
	"strings"

	"golang-git-graph/internal/git"

	"github.com/charmbracelet/lipgloss"
)

type GraphPane struct {
	Commits []git.Commit
	Cursor  int
	Offset  int
}

func NewGraphPane() *GraphPane {
	return &GraphPane{
		Commits: []git.Commit{},
		Cursor:  0,
		Offset:  0,
	}
}

func (gp *GraphPane) UpdateCommits(commits []git.Commit) {
	gp.Commits = commits
	// Bounds check cursor
	if gp.Cursor >= len(gp.Commits) {
		gp.Cursor = len(gp.Commits) - 1
	}
	if gp.Cursor < 0 {
		gp.Cursor = 0
	}
}

func (gp *GraphPane) ScrollUp() {
	if gp.Cursor > 0 {
		gp.Cursor--
	}
}

func (gp *GraphPane) ScrollDown() {
	if gp.Cursor < len(gp.Commits)-1 {
		gp.Cursor++
	}
}

func (gp *GraphPane) SelectedCommit() *git.Commit {
	if len(gp.Commits) == 0 || gp.Cursor < 0 || gp.Cursor >= len(gp.Commits) {
		return nil
	}
	return &gp.Commits[gp.Cursor]
}

func (gp *GraphPane) View(styles Styles, width int, height int, focused bool) string {
	if len(gp.Commits) == 0 {
		return "No commits found."
	}

	// Adjust offset for scrolling
	if gp.Cursor < gp.Offset {
		gp.Offset = gp.Cursor
	} else if gp.Cursor >= gp.Offset+height {
		gp.Offset = gp.Cursor - height + 1
	}

	var sb strings.Builder
	
	// Print lines within viewport
	end := gp.Offset + height
	if end > len(gp.Commits) {
		end = len(gp.Commits)
	}

	for i := gp.Offset; i < end; i++ {
		c := gp.Commits[i]
		isSelected := i == gp.Cursor

		// 1. Colorize the graph lines (characters like |, *, \, /, etc.)
		coloredGraph := colorizeGraph(c.GraphChars, styles.GraphColors)

		// 2. Build the rest of the commit line
		var commitText string
		if c.Hash != "" {
			// Style components
			hashStr := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow)).Render(c.Hash)
			refsStr := colorizeRefs(c.Refs, styles)
			
			// Format author with single unicode icon (truncating on narrow screens)
			authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCyan))
			authorText := c.Author
			if width < 80 && len(authorText) > 10 {
				authorText = authorText[:7] + "..."
			}
			authorStr := " 👤 " + authorStyle.Render(authorText)
			
			// Format date with clock icon and both formats (hiding/truncating on narrow screens)
			dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorComment))
			dateText := c.Date
			if width >= 100 && c.AbsDate != "" {
				dateText = fmt.Sprintf("%s (%s)", c.Date, c.AbsDate)
			}
			dateStr := ""
			if width > 60 {
				dateStr = " ⏱️ " + dateStyle.Render(dateText)
			}
			
			// Safely truncate subject to fit the remaining width
			graphWidth := lipgloss.Width(coloredGraph)
			hashWidth := lipgloss.Width(hashStr)
			refsWidth := lipgloss.Width(refsStr)
			authorWidth := lipgloss.Width(authorStr)
			dateWidth := lipgloss.Width(dateStr)
			
			maxSubjectWidth := width - graphWidth - hashWidth - refsWidth - authorWidth - dateWidth - 4
			
			subjectStr := c.Subject
			if maxSubjectWidth > 5 {
				if len(subjectStr) > maxSubjectWidth {
					subjectStr = subjectStr[:maxSubjectWidth-3] + "..."
				}
			} else if maxSubjectWidth > 0 {
				if len(subjectStr) > maxSubjectWidth {
					subjectStr = subjectStr[:maxSubjectWidth]
				}
			} else {
				subjectStr = ""
			}
			
			commitText = fmt.Sprintf("%s%s %s%s%s", hashStr, refsStr, subjectStr, authorStr, dateStr)
		} else {
			// Pure graph line
			commitText = ""
		}

		lineContent := fmt.Sprintf("%s%s", coloredGraph, commitText)
		
		if isSelected && focused {
			sb.WriteString(styles.SelectedLineStyle.Width(width).Render(lineContent) + "\n")
		} else {
			sb.WriteString(styles.NormalLineStyle.Render(lineContent) + "\n")
		}
	}

	return sb.String()
}

// colorizeGraph loops through characters in the graph line, maps them to Unicode, and colorizes by column index.
func colorizeGraph(graphStr string, palette []lipgloss.Style) string {
	if len(palette) == 0 {
		return graphStr
	}

	var sb strings.Builder
	
	for i, char := range graphStr {
		if char == ' ' {
			sb.WriteRune(char)
		} else {
			// Map ASCII log graph characters to beautiful unicode line segments
			mappedStr := mapGraphRune(char)
			
			// Select color from palette based on horizontal index i to align vertical lanes
			style := palette[i % len(palette)]
			
			// Bold the node marker
			if char == '*' {
				style = style.Copy().Bold(true)
			}
			
			sb.WriteString(style.Render(mappedStr))
		}
	}
	
	return sb.String()
}

func mapGraphRune(r rune) string {
	switch r {
	case '|':
		return "│"
	case '/':
		return "╱"
	case '\\':
		return "╲"
	case '*':
		return "●"
	case '_':
		return "─"
	default:
		return string(r)
	}
}

func colorizeRefs(refsStr string, styles Styles) string {
	if refsStr == "" {
		return ""
	}
	
	items := strings.Split(refsStr, ", ")
	var colorizedItems []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		
		if strings.HasPrefix(item, "HEAD -> ") {
			headText := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCyan)).Bold(true).Render("HEAD ➔ ")
			branchName := strings.TrimPrefix(item, "HEAD -> ")
			branchText := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Bold(true).Render(branchName)
			colorizedItems = append(colorizedItems, headText+branchText)
		} else if strings.HasPrefix(item, "tag: ") {
			tagName := strings.TrimPrefix(item, "tag: ")
			tagText := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow)).Bold(true).Render("🏷️ " + tagName)
			colorizedItems = append(colorizedItems, tagText)
		} else if strings.HasPrefix(item, "origin/") || strings.Contains(item, "/") {
			remoteText := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)).Render(item)
			colorizedItems = append(colorizedItems, remoteText)
		} else {
			branchText := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen)).Bold(true).Render(item)
			colorizedItems = append(colorizedItems, branchText)
		}
	}
	
	if len(colorizedItems) == 0 {
		return ""
	}
	return " (" + strings.Join(colorizedItems, ", ") + ")"
}
