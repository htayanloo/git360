package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type DiffPane struct {
	Viewport viewport.Model
	RawDiff  string
	Title    string
	Width    int
	Height   int
}

func NewDiffPane() *DiffPane {
	// Initialize with placeholder dimensions, will resize in App
	vp := viewport.New(0, 0)
	return &DiffPane{
		Viewport: vp,
		RawDiff:  "",
		Title:    "Diff Viewer",
	}
}

func (dp *DiffPane) Resize(width, height int) {
	dp.Width = width
	dp.Height = height
	dp.Viewport.Width = width
	dp.Viewport.Height = height
}

func (dp *DiffPane) SetContent(title string, rawDiff string, styles Styles) {
	dp.Title = title
	dp.RawDiff = rawDiff
	
	if rawDiff == "" {
		dp.Viewport.SetContent("No changes to display.")
		dp.Viewport.GotoTop()
		return
	}

	// Apply syntax highlighting line by line
	lines := strings.Split(rawDiff, "\n")
	highlighted := make([]string, len(lines))

	for i, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			highlighted[i] = styles.DiffAddedStyle.Render(line)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			highlighted[i] = styles.DiffRemovedStyle.Render(line)
		} else if strings.HasPrefix(line, "@@") {
			highlighted[i] = styles.DiffHeaderStyle.Render(line)
		} else if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			highlighted[i] = styles.DiffHeaderStyle.Render(line)
		} else {
			highlighted[i] = styles.NormalLineStyle.Render(line)
		}
	}

	dp.Viewport.SetContent(strings.Join(highlighted, "\n"))
	dp.Viewport.GotoTop()
}

func (dp *DiffPane) Update(msg tea.Msg) (tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			dp.Viewport.LineDown(1)
			return nil
		case "k":
			dp.Viewport.LineUp(1)
			return nil
		case "ctrl+d":
			dp.Viewport.HalfPageDown()
			return nil
		case "ctrl+u":
			dp.Viewport.HalfPageUp()
			return nil
		}
	}
	var cmd tea.Cmd
	dp.Viewport, cmd = dp.Viewport.Update(msg)
	return cmd
}

func (dp *DiffPane) View() string {
	return dp.Viewport.View()
}
