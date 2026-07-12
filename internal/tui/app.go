package tui

import (
	"fmt"
	"strings"
	"time"

	"golang-git-graph/internal/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type gitStateMsg struct {
	Branch   string
	Upstream string
	Ahead    int
	Behind   int
	Changes  []git.FileChange
	Commits  []git.Commit
}

type fetchFinishedMsg struct {
	Time time.Time
}

type errorMsg struct {
	err error
}

type AppModel struct {
	GitClient         *git.Client
	Styles            Styles
	ActivePane        Pane
	GraphPane         *GraphPane
	StatusPane        *StatusPane
	DiffPane          *DiffPane
	MetaPane          *MetaPane
	BranchBrowser     *BranchBrowser
	ShowBranchBrowser bool
	Width             int
	Height            int
	Err               error
	LastRefresh       time.Time
	PrevChangesMap    map[string]git.FileChange // Used to log changes over time
	IsFetching        bool
}

func NewAppModel(client *git.Client) *AppModel {
	return &AppModel{
		GitClient:         client,
		Styles:            DefaultStyles(),
		ActivePane:        PaneGraph,
		GraphPane:         NewGraphPane(),
		StatusPane:        NewStatusPane(),
		DiffPane:          NewDiffPane(),
		MetaPane:          NewMetaPane(),
		BranchBrowser:     NewBranchBrowser(),
		ShowBranchBrowser: false,
		PrevChangesMap:    make(map[string]git.FileChange),
	}
}

// Start the auto-polling tick command
func (m *AppModel) tickChan() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *AppModel) checkGitStateCmd() tea.Cmd {
	return func() tea.Msg {
		branch, upstream, ahead, behind, changes, err := m.GitClient.GetStatus()
		if err != nil {
			return errorMsg{err}
		}
		
		commits, err := m.GitClient.GetCommits(100)
		if err != nil {
			return errorMsg{err}
		}

		return gitStateMsg{
			Branch:   branch,
			Upstream: upstream,
			Ahead:    ahead,
			Behind:   behind,
			Changes:  changes,
			Commits:  commits,
		}
	}
}

func (m *AppModel) fetchOriginCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.GitClient.FetchOrigin()
		if err != nil {
			return errorMsg{err}
		}
		return fetchFinishedMsg{Time: time.Now()}
	}
}

func (m *AppModel) loadDiffCmd() tea.Cmd {
	return func() tea.Msg {
		switch m.ActivePane {
		case PaneGraph:
			commit := m.GraphPane.SelectedCommit()
			if commit == nil || commit.Hash == "" {
				return nil
			}
			diff, err := m.GitClient.GetCommitDiff(commit.Hash)
			if err != nil {
				return errorMsg{err}
			}
			return diffLoadedMsg{Title: "Commit: " + commit.Hash, Content: diff}
		case PaneStatus:
			file := m.StatusPane.SelectedFile()
			if file == nil {
				return nil
			}
			diff, err := m.GitClient.GetDiff(*file)
			if err != nil {
				return errorMsg{err}
			}
			return diffLoadedMsg{Title: "File Diff: " + file.Path, Content: diff}
		}
		return nil
	}
}

type diffLoadedMsg struct {
	Title   string
	Content string
}

type branchesLoadedMsg struct {
	branches []string
}

type tagsLoadedMsg struct {
	tags []string
}

type checkoutFinishedMsg struct {
	ref string
}

type revertFinishedMsg struct {
	hash string
}

func (m *AppModel) loadBranchesCmd() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.GitClient.GetBranches()
		if err != nil {
			return errorMsg{err}
		}
		return branchesLoadedMsg{branches}
	}
}

func (m *AppModel) loadTagsCmd() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.GitClient.GetTags()
		if err != nil {
			return errorMsg{err}
		}
		return tagsLoadedMsg{tags}
	}
}

func (m *AppModel) checkoutCmd(ref string) tea.Cmd {
	return func() tea.Msg {
		err := m.GitClient.Checkout(ref)
		if err != nil {
			return errorMsg{err}
		}
		return checkoutFinishedMsg{ref}
	}
}

func (m *AppModel) compareRefCmd(ref string) tea.Cmd {
	return func() tea.Msg {
		diff, err := m.GitClient.GetCompareDiff(ref)
		if err != nil {
			return errorMsg{err}
		}
		return diffLoadedMsg{Title: "Compare: HEAD.." + ref, Content: diff}
	}
}

func (m *AppModel) compareWithOriginCmd() tea.Cmd {
	return func() tea.Msg {
		diff, err := m.GitClient.GetDiffWithOrigin(m.MetaPane.Branch)
		if err != nil {
			return errorMsg{err}
		}
		return diffLoadedMsg{Title: "Compare: HEAD..origin/" + m.MetaPane.Branch, Content: diff}
	}
}

func (m *AppModel) revertCommitCmd(hash string) tea.Cmd {
	return func() tea.Msg {
		err := m.GitClient.RevertCommit(hash)
		if err != nil {
			return errorMsg{err}
		}
		return revertFinishedMsg{hash}
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.tickChan(),
		m.checkGitStateCmd(),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			// Cycle pane clockwise
			m.ActivePane = (m.ActivePane + 1) % 3
			cmds = append(cmds, m.loadDiffCmd())

		case "shift+tab":
			// Cycle pane counter-clockwise
			if m.ActivePane == 0 {
				m.ActivePane = 2
			} else {
				m.ActivePane--
			}
			cmds = append(cmds, m.loadDiffCmd())

		case "r":
			// Manual refresh
			m.MetaPane.StatusText = "Refreshing..."
			m.MetaPane.AddLog("Manual refresh triggered")
			cmds = append(cmds, m.checkGitStateCmd())

		case "f":
			// Manual fetch
			if !m.IsFetching {
				m.IsFetching = true
				m.MetaPane.StatusText = "Fetching origin..."
				m.MetaPane.AddLog("Fetching origin in background...")
				cmds = append(cmds, m.fetchOriginCmd())
			}

		case "b":
			// Toggle branch browser
			if m.ShowBranchBrowser && !m.BranchBrowser.IsTags {
				m.ShowBranchBrowser = false
				cmds = append(cmds, m.loadDiffCmd())
			} else {
				m.ShowBranchBrowser = true
				m.ActivePane = PaneStatus
				m.MetaPane.AddLog("Browsing branches...")
				cmds = append(cmds, m.loadBranchesCmd())
			}

		case "t":
			// Toggle tag browser
			if m.ShowBranchBrowser && m.BranchBrowser.IsTags {
				m.ShowBranchBrowser = false
				cmds = append(cmds, m.loadDiffCmd())
			} else {
				m.ShowBranchBrowser = true
				m.ActivePane = PaneStatus
				m.MetaPane.AddLog("Browsing tags...")
				cmds = append(cmds, m.loadTagsCmd())
			}

		case "escape":
			if m.ShowBranchBrowser {
				m.ShowBranchBrowser = false
				cmds = append(cmds, m.loadDiffCmd())
			}

		case "enter":
			if m.ShowBranchBrowser && m.ActivePane == PaneStatus {
				selected := m.BranchBrowser.SelectedItem()
				if selected != "" {
					m.MetaPane.StatusText = "Checking out..."
					cmds = append(cmds, m.checkoutCmd(selected))
				}
			}

		case "c":
			if m.ShowBranchBrowser && m.ActivePane == PaneStatus {
				selected := m.BranchBrowser.SelectedItem()
				if selected != "" {
					cmds = append(cmds, m.compareRefCmd(selected))
				}
			}

		case "o":
			if m.ActivePane == PaneGraph {
				m.MetaPane.AddLog("Comparing current HEAD with origin...")
				cmds = append(cmds, m.compareWithOriginCmd())
			}

		case "v":
			if m.ActivePane == PaneGraph {
				commit := m.GraphPane.SelectedCommit()
				if commit != nil && commit.Hash != "" {
					m.MetaPane.AddLog(fmt.Sprintf("Reverting commit %s...", commit.Hash))
					cmds = append(cmds, m.revertCommitCmd(commit.Hash))
				}
			}

		case "up", "k":
			if m.ActivePane == PaneGraph {
				m.GraphPane.ScrollUp()
				cmds = append(cmds, m.loadDiffCmd())
			} else if m.ActivePane == PaneStatus {
				if m.ShowBranchBrowser {
					m.BranchBrowser.ScrollUp()
					selected := m.BranchBrowser.SelectedItem()
					if selected != "" {
						cmds = append(cmds, m.compareRefCmd(selected))
					}
				} else {
					m.StatusPane.ScrollUp()
					cmds = append(cmds, m.loadDiffCmd())
				}
			}

		case "down", "j":
			if m.ActivePane == PaneGraph {
				m.GraphPane.ScrollDown()
				cmds = append(cmds, m.loadDiffCmd())
			} else if m.ActivePane == PaneStatus {
				if m.ShowBranchBrowser {
					m.BranchBrowser.ScrollDown()
					selected := m.BranchBrowser.SelectedItem()
					if selected != "" {
						cmds = append(cmds, m.compareRefCmd(selected))
					}
				} else {
					m.StatusPane.ScrollDown()
					cmds = append(cmds, m.loadDiffCmd())
				}
			}

		case "s":
			// Stage current file if in status pane
			if m.ActivePane == PaneStatus && !m.ShowBranchBrowser {
				file := m.StatusPane.SelectedFile()
				if file != nil && !file.Staged {
					err := m.GitClient.StageFile(file.Path)
					if err != nil {
						m.Err = err
						m.MetaPane.AddLog(fmt.Sprintf("Stage error: %v", err))
					} else {
						m.MetaPane.AddLog(fmt.Sprintf("Staged file: %s", file.Path))
						cmds = append(cmds, m.checkGitStateCmd())
					}
				}
			}

		case "u":
			// Unstage current file if in status pane
			if m.ActivePane == PaneStatus && !m.ShowBranchBrowser {
				file := m.StatusPane.SelectedFile()
				if file != nil && file.Staged {
					err := m.GitClient.UnstageFile(file.Path)
					if err != nil {
						m.Err = err
						m.MetaPane.AddLog(fmt.Sprintf("Unstage error: %v", err))
					} else {
						m.MetaPane.AddLog(fmt.Sprintf("Unstaged file: %s", file.Path))
						cmds = append(cmds, m.checkGitStateCmd())
					}
				}
			}

		case " ": // Space bar toggles staging
			if m.ActivePane == PaneStatus && !m.ShowBranchBrowser {
				file := m.StatusPane.SelectedFile()
				if file != nil {
					var err error
					if file.Staged {
						err = m.GitClient.UnstageFile(file.Path)
						if err == nil {
							m.MetaPane.AddLog(fmt.Sprintf("Unstaged file: %s", file.Path))
						}
					} else {
						err = m.GitClient.StageFile(file.Path)
						if err == nil {
							m.MetaPane.AddLog(fmt.Sprintf("Staged file: %s", file.Path))
						}
					}
					if err != nil {
						m.Err = err
						m.MetaPane.AddLog(fmt.Sprintf("Toggle error: %v", err))
					} else {
						cmds = append(cmds, m.checkGitStateCmd())
					}
				}
			}
		}

	case tickMsg:
		cmds = append(cmds, m.checkGitStateCmd())
		cmds = append(cmds, m.tickChan())

	case gitStateMsg:
		m.LastRefresh = time.Now()
		m.MetaPane.Branch = msg.Branch
		m.MetaPane.Upstream = msg.Upstream
		m.MetaPane.Ahead = msg.Ahead
		m.MetaPane.Behind = msg.Behind
		
		// If we're not currently doing a manual fetch background check, state is idle polling
		if !m.IsFetching {
			m.MetaPane.StatusText = "Watching changes..."
		}
		
		// Parse diff logs to see what has changed since last tick
		m.logStateTransitions(msg.Changes)
		
		m.GraphPane.UpdateCommits(msg.Commits)
		m.StatusPane.UpdateChanges(msg.Changes)
		
		// Reload diff in case of file modifications or changes
		if !m.ShowBranchBrowser {
			cmds = append(cmds, m.loadDiffCmd())
		}

	case fetchFinishedMsg:
		m.IsFetching = false
		m.MetaPane.LastFetch = msg.Time
		m.MetaPane.StatusText = "Watching changes..."
		m.MetaPane.AddLog("Fetch origin complete")
		// Refresh state to read new ahead/behind info
		cmds = append(cmds, m.checkGitStateCmd())

	case diffLoadedMsg:
		m.DiffPane.SetContent(msg.Title, msg.Content, m.Styles)

	case branchesLoadedMsg:
		m.BranchBrowser.SetItems(msg.branches, false)
		selected := m.BranchBrowser.SelectedItem()
		if selected != "" {
			cmds = append(cmds, m.compareRefCmd(selected))
		}

	case tagsLoadedMsg:
		m.BranchBrowser.SetItems(msg.tags, true)
		selected := m.BranchBrowser.SelectedItem()
		if selected != "" {
			cmds = append(cmds, m.compareRefCmd(selected))
		}

	case checkoutFinishedMsg:
		m.MetaPane.AddLog(fmt.Sprintf("Switched to %s", msg.ref))
		m.ShowBranchBrowser = false
		cmds = append(cmds, m.checkGitStateCmd())

	case revertFinishedMsg:
		m.MetaPane.AddLog(fmt.Sprintf("Reverted commit %s", msg.hash))
		cmds = append(cmds, m.checkGitStateCmd())

	case errorMsg:
		m.Err = msg.err
		m.MetaPane.AddLog(fmt.Sprintf("Error: %v", msg.err))

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	// Delegate diff pane update if active (for scrolling diff viewport)
	if m.ActivePane == PaneDiff {
		diffCmd := m.DiffPane.Update(msg)
		cmds = append(cmds, diffCmd)
	}

	return m, tea.Batch(cmds...)
}

// logStateTransitions compares file status changes to log agent activity.
func (m *AppModel) logStateTransitions(newChanges []git.FileChange) {
	newMap := make(map[string]git.FileChange)
	for _, change := range newChanges {
		// Unique key including path and staged state
		key := fmt.Sprintf("%s:%t", change.Path, change.Staged)
		newMap[key] = change
	}

	// 1. Detect additions / modifications
	for key, file := range newMap {
		oldFile, exists := m.PrevChangesMap[key]
		if !exists {
			// Check if same path existed with different stage state to determine staging action
			oppositeKey := fmt.Sprintf("%s:%t", file.Path, !file.Staged)
			if _, stagedStateChanged := m.PrevChangesMap[oppositeKey]; stagedStateChanged {
				// Handled by staging toggle detection below
				continue
			}
			
			// This is a new change detected
			action := "modified"
			if file.Status == "?" {
				action = "untracked"
			} else if file.Status == "A" {
				action = "created"
			} else if file.Status == "D" {
				action = "deleted"
			}
			m.MetaPane.AddLog(fmt.Sprintf("File %s: %s", action, file.Path))
		} else if oldFile.Status != file.Status {
			m.MetaPane.AddLog(fmt.Sprintf("File changed status (%s -> %s): %s", oldFile.Status, file.Status, file.Path))
		}
	}

	// 2. Detect staging toggles (e.g. was staged:false and is now staged:true)
	for _, file := range newChanges {
		if file.Staged {
			oldUnstagedKey := fmt.Sprintf("%s:false", file.Path)
			if _, wasUnstaged := m.PrevChangesMap[oldUnstagedKey]; wasUnstaged {
				m.MetaPane.AddLog(fmt.Sprintf("Stage change: %s staged", file.Path))
			}
		} else {
			oldStagedKey := fmt.Sprintf("%s:true", file.Path)
			if _, wasStaged := m.PrevChangesMap[oldStagedKey]; wasStaged {
				m.MetaPane.AddLog(fmt.Sprintf("Stage change: %s unstaged", file.Path))
			}
		}
	}

	// 3. Detect files that disappeared from status (meaning they were committed or reverted)
	for key, file := range m.PrevChangesMap {
		_, stillExists := newMap[key]
		oppositeKey := fmt.Sprintf("%s:%t", file.Path, !file.Staged)
		_, stagedStateChanged := newMap[oppositeKey]
		
		if !stillExists && !stagedStateChanged {
			m.MetaPane.AddLog(fmt.Sprintf("File cleared: %s", file.Path))
		}
	}

	m.PrevChangesMap = newMap
}

func (m *AppModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing layout..."
	}

	// Calculate layout splits (headers & footers take 1 line each)
	availableHeight := m.Height - 2
	if availableHeight < 4 {
		return "Terminal window is too small!"
	}

	topHeight := availableHeight / 2
	bottomHeight := availableHeight - topHeight

	leftWidth := (m.Width * 6) / 10
	rightWidth := m.Width - leftWidth

	// Style title headers for each pane
	paneTitle := func(title string, pane Pane) string {
		if m.ActivePane == pane {
			return m.Styles.PaneTitleActive.Render("█ " + title)
		}
		return m.Styles.PaneTitleInactive.Render("░ " + title)
	}

	// 1. Render Top Row
	graphContent := m.GraphPane.View(m.Styles, leftWidth-4, topHeight-2, m.ActivePane == PaneGraph)
	graphBox := m.Styles.InactivePaneBorder.Copy()
	if m.ActivePane == PaneGraph {
		graphBox = m.Styles.ActivePaneBorder.Copy()
	}
	graphView := graphBox.
		Width(leftWidth - 2).
		Height(topHeight - 2).
		Render(paneTitle("GIT HISTORY GRAPH", PaneGraph) + "\n" + graphContent)

	// Determine status pane contents and title dynamically
	var statusContent string
	var statusTitle string
	if m.ShowBranchBrowser {
		statusContent = m.BranchBrowser.View(m.Styles, rightWidth-4, topHeight-2)
		if m.BranchBrowser.IsTags {
			statusTitle = "TAG BROWSER (Enter=Checkout, c/Scroll=Compare, Esc=Close)"
		} else {
			statusTitle = "BRANCH BROWSER (Enter=Checkout, c/Scroll=Compare, Esc=Close)"
		}
	} else {
		statusContent = m.StatusPane.View(m.Styles, rightWidth-4, topHeight-2, m.ActivePane == PaneStatus)
		statusTitle = "STAGE CONTROL (Space=Toggle, s=Stage, u=Unstage)"
	}

	statusBox := m.Styles.InactivePaneBorder.Copy()
	if m.ActivePane == PaneStatus {
		statusBox = m.Styles.ActivePaneBorder.Copy()
	}
	statusView := statusBox.
		Width(rightWidth - 2).
		Height(topHeight - 2).
		Render(paneTitle(statusTitle, PaneStatus) + "\n" + statusContent)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, graphView, statusView)

	// 2. Render Bottom Row
	// Resize diff viewport to fit its container
	m.DiffPane.Resize(leftWidth-4, bottomHeight-2)
	diffContent := m.DiffPane.View()
	diffBox := m.Styles.InactivePaneBorder.Copy()
	if m.ActivePane == PaneDiff {
		diffBox = m.Styles.ActivePaneBorder.Copy()
	}
	diffView := diffBox.
		Width(leftWidth - 2).
		Height(bottomHeight - 2).
		Render(paneTitle(m.DiffPane.Title, PaneDiff) + "\n" + diffContent)

	metaContent := m.MetaPane.View(m.Styles, rightWidth-4, bottomHeight-2)
	metaBox := m.Styles.InactivePaneBorder.Copy() // Meta pane is informational, never directly focused
	metaView := metaBox.
		Width(rightWidth - 2).
		Height(bottomHeight - 2).
		Render(m.Styles.PaneTitleInactive.Render("░ AGENT METRICS & SYSTEM LOGS") + "\n" + metaContent)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, diffView, metaView)

	// 3. Assemble Header and Footer
	headerText := fmt.Sprintf(" GIT-360  |  Active Branch: %s  |  Last update: %s", 
		m.MetaPane.Branch, 
		m.LastRefresh.Format("15:04:05"),
	)
	header := m.Styles.HeaderStyle.Width(m.Width).Render(headerText)

	// Help guidelines footer
	var helpParts []string
	helpParts = append(helpParts, fmt.Sprintf("%s Switch Pane", m.Styles.HelpKeyStyle.Render("Tab")))
	helpParts = append(helpParts, fmt.Sprintf("%s/%s Scroll", m.Styles.HelpKeyStyle.Render("↑↓"), m.Styles.HelpKeyStyle.Render("j/k")))
	
	if m.ShowBranchBrowser {
		helpParts = append(helpParts, fmt.Sprintf("%s Checkout", m.Styles.HelpKeyStyle.Render("Enter")))
		helpParts = append(helpParts, fmt.Sprintf("%s Compare", m.Styles.HelpKeyStyle.Render("c")))
		helpParts = append(helpParts, fmt.Sprintf("%s Close", m.Styles.HelpKeyStyle.Render("Esc")))
	} else {
		helpParts = append(helpParts, fmt.Sprintf("%s Toggle Stage", m.Styles.HelpKeyStyle.Render("Space")))
		helpParts = append(helpParts, fmt.Sprintf("%s Stage", m.Styles.HelpKeyStyle.Render("s")))
		helpParts = append(helpParts, fmt.Sprintf("%s Unstage", m.Styles.HelpKeyStyle.Render("u")))
	}
	
	helpParts = append(helpParts, fmt.Sprintf("%s Branches", m.Styles.HelpKeyStyle.Render("b")))
	helpParts = append(helpParts, fmt.Sprintf("%s Tags", m.Styles.HelpKeyStyle.Render("t")))
	
	if m.ActivePane == PaneGraph {
		helpParts = append(helpParts, fmt.Sprintf("%s Compare Origin", m.Styles.HelpKeyStyle.Render("o")))
		helpParts = append(helpParts, fmt.Sprintf("%s Revert", m.Styles.HelpKeyStyle.Render("v")))
	}
	
	helpParts = append(helpParts, fmt.Sprintf("%s Fetch", m.Styles.HelpKeyStyle.Render("f")))
	helpParts = append(helpParts, fmt.Sprintf("%s Refresh", m.Styles.HelpKeyStyle.Render("r")))
	helpParts = append(helpParts, fmt.Sprintf("%s Quit", m.Styles.HelpKeyStyle.Render("q")))
	
	helpText := " " + strings.Join(helpParts, "  •  ")
	footer := m.Styles.FooterStyle.Width(m.Width).Render(helpText)

	if m.Err != nil {
		errorBar := lipgloss.NewStyle().Background(lipgloss.Color(ColorRed)).Foreground(lipgloss.Color(ColorForeground)).Bold(true).Width(m.Width).Render(fmt.Sprintf(" ERROR: %v (press 'r' to retry)", m.Err))
		return lipgloss.JoinVertical(lipgloss.Left, header, topRow, bottomRow, errorBar, footer)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, topRow, bottomRow, footer)
}

func paneName(p Pane) string {
	switch p {
	case PaneGraph:
		return "Git History Graph"
	case PaneStatus:
		return "Stage Control"
	case PaneDiff:
		return "Diff Viewer"
	default:
		return "Unknown"
	}
}
