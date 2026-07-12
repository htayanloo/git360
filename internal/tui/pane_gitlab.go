package tui

import (
	"fmt"
	"strings"

	"golang-git-graph/internal/gitlab"

	"github.com/charmbracelet/lipgloss"
)

type GitLabTab int

const (
	GitLabTabMR GitLabTab = iota
	GitLabTabPipelines
	GitLabTabIssues
)

type GitLabPane struct {
	ActiveTab    GitLabTab
	MRs          []gitlab.MergeRequest
	Pipelines    []gitlab.Pipeline
	Issues       []gitlab.Issue
	JobsMap      map[int][]gitlab.Job // pipelineID -> jobs
	Cursor       int
	Offset       int
	IsLoading    bool
	TokenMissing bool
	Error        error
}

func NewGitLabPane() *GitLabPane {
	return &GitLabPane{
		ActiveTab: GitLabTabMR,
		MRs:       []gitlab.MergeRequest{},
		Pipelines: []gitlab.Pipeline{},
		Issues:    []gitlab.Issue{},
		JobsMap:   make(map[int][]gitlab.Job),
		Cursor:    0,
		Offset:    0,
	}
}

func (gp *GitLabPane) TotalItems() int {
	switch gp.ActiveTab {
	case GitLabTabMR:
		return len(gp.MRs)
	case GitLabTabPipelines:
		return len(gp.Pipelines)
	case GitLabTabIssues:
		return len(gp.Issues)
	default:
		return 0
	}
}

func (gp *GitLabPane) ScrollUp() {
	if gp.Cursor > 0 {
		gp.Cursor--
	}
}

func (gp *GitLabPane) ScrollDown() {
	if gp.Cursor < gp.TotalItems()-1 {
		gp.Cursor++
	}
}

func (gp *GitLabPane) SetTab(tab GitLabTab) {
	gp.ActiveTab = tab
	gp.Cursor = 0
	gp.Offset = 0
}

func (gp *GitLabPane) NextTab() {
	gp.SetTab((gp.ActiveTab + 1) % 3)
}

func (gp *GitLabPane) PrevTab() {
	if gp.ActiveTab == 0 {
		gp.SetTab(2)
	} else {
		gp.SetTab(gp.ActiveTab - 1)
	}
}

func (gp *GitLabPane) View(styles Styles, width int, height int) string {
	if gp.TokenMissing {
		var sb strings.Builder
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)).Bold(true).Render("⚠️  Authentication Token Missing") + "\n\n")
		sb.WriteString(styles.HelpDescStyle.Render("To access GitLab integrations, you must set the ") + "\n")
		sb.WriteString(styles.HelpKeyStyle.Render("GITLAB_TOKEN") + styles.HelpDescStyle.Render(" environment variable:") + "\n\n")
		sb.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(ColorCurrent)).Foreground(lipgloss.Color(ColorForeground)).Padding(0, 1).Render("export GITLAB_TOKEN=\"your_access_token\"") + "\n\n")
		sb.WriteString(styles.HelpDescStyle.Render("Create a Personal Access Token in GitLab settings") + "\n")
		sb.WriteString(styles.HelpDescStyle.Render("with API read permissions."))
		return sb.String()
	}

	if gp.Error != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed)).Render(fmt.Sprintf("GitLab Error: %v\nPress 'r' to try again.", gp.Error))
	}

	if gp.IsLoading {
		return "⏳ Loading GitLab data..."
	}

	// 1. Render Sub-Tabs Header
	tabStyle := func(title string, active bool) string {
		if active {
			return lipgloss.NewStyle().
				Background(lipgloss.Color(ColorCurrent)).
				Foreground(lipgloss.Color(ColorCyan)).
				Bold(true).
				Padding(0, 1).
				Render(title)
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorComment)).
			Padding(0, 1).
			Render(title)
	}

	tabMR := tabStyle("MRs", gp.ActiveTab == GitLabTabMR)
	tabPipe := tabStyle("Pipelines", gp.ActiveTab == GitLabTabPipelines)
	tabIssue := tabStyle("Issues", gp.ActiveTab == GitLabTabIssues)
	tabsRow := tabMR + " " + tabPipe + " " + tabIssue + "\n\n"

	// 2. Adjust scrolling offset
	contentHeight := height - 3 // leave room for tabs header
	if contentHeight < 1 {
		contentHeight = 1
	}

	total := gp.TotalItems()
	if total == 0 {
		var msg string
		switch gp.ActiveTab {
		case GitLabTabMR:
			msg = "No open Merge Requests found."
		case GitLabTabPipelines:
			msg = "No pipelines found."
		case GitLabTabIssues:
			msg = "No open Issues found."
		}
		return tabsRow + styles.HelpDescStyle.Render(msg)
	}

	if gp.Cursor < gp.Offset {
		gp.Offset = gp.Cursor
	} else if gp.Cursor >= gp.Offset+contentHeight {
		gp.Offset = gp.Cursor - contentHeight + 1
	}

	var sb strings.Builder
	sb.WriteString(tabsRow)

	end := gp.Offset + contentHeight
	if end > total {
		end = total
	}

	for i := gp.Offset; i < end; i++ {
		isSelected := i == gp.Cursor
		var line string

		switch gp.ActiveTab {
		case GitLabTabMR:
			mr := gp.MRs[i]
			title := mr.Title
			maxTitleLen := width - 15
			if maxTitleLen > 5 && len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}
			line = fmt.Sprintf("  !%d  %s", mr.IID, title)

		case GitLabTabPipelines:
			p := gp.Pipelines[i]
			statusIcon := "[?]"
			statusColor := ColorComment
			switch p.Status {
			case "success":
				statusIcon = "✔"
				statusColor = ColorGreen
			case "failed":
				statusIcon = "✘"
				statusColor = ColorRed
			case "running":
				statusIcon = "●"
				statusColor = ColorYellow
			case "pending":
				statusIcon = "○"
				statusColor = ColorOrange
			}
			statusStr := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(statusIcon)
			refName := p.Ref
			maxRefLen := width - 20
			if maxRefLen > 5 && len(refName) > maxRefLen {
				refName = refName[:maxRefLen-3] + "..."
			}
			line = fmt.Sprintf("  %s  #%d (%s)", statusStr, p.ID, refName)

		case GitLabTabIssues:
			issue := gp.Issues[i]
			title := issue.Title
			maxTitleLen := width - 15
			if maxTitleLen > 5 && len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}
			line = fmt.Sprintf("  #%d  %s", issue.IID, title)
		}

		if isSelected {
			sb.WriteString(styles.SelectedLineStyle.Width(width).Render(line) + "\n")
		} else {
			sb.WriteString(styles.NormalLineStyle.Render(line) + "\n")
		}
	}

	return sb.String()
}
