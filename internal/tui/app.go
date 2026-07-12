package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang-git-graph/internal/git"
	"golang-git-graph/internal/gitlab"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const Version = "v0.2.0"

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
	GitClient             *git.Client
	GitLabClient          *gitlab.Client
	Styles                Styles
	ActivePane            Pane
	GraphPane             *GraphPane
	StatusPane            *StatusPane
	DiffPane              *DiffPane
	MetaPane              *MetaPane
	BranchBrowser         *BranchBrowser
	GitLabPane            *GitLabPane
	ShowBranchBrowser     bool
	ShowGitLab            bool
	Width                 int
	Height                int
	Err                   error
	LastRefresh           time.Time
	PrevChangesMap        map[string]git.FileChange // Used to log changes over time
	IsFetching            bool
	HorizontalSplitOffset int
	VerticalSplitOffset   int
	LatestVersion         string
	IsUpdating            bool
}

func NewAppModel(client *git.Client) *AppModel {
	m := &AppModel{
		GitClient:             client,
		Styles:                DefaultStyles(),
		ActivePane:            PaneGraph,
		GraphPane:             NewGraphPane(),
		StatusPane:            NewStatusPane(),
		DiffPane:              NewDiffPane(),
		MetaPane:              NewMetaPane(),
		BranchBrowser:         NewBranchBrowser(),
		GitLabPane:            NewGitLabPane(),
		ShowBranchBrowser:     false,
		ShowGitLab:            false,
		PrevChangesMap:        make(map[string]git.FileChange),
		HorizontalSplitOffset: 0,
		VerticalSplitOffset:   0,
		LatestVersion:         "",
		IsUpdating:            false,
	}

	// Try to initialize GitLab client from remote URL
	remoteURL, err := client.GetRemoteURL()
	if err == nil {
		glClient, glErr := gitlab.NewClient(remoteURL)
		if glErr == nil {
			m.GitLabClient = glClient
		}
	}

	return m
}

func (m *AppModel) layoutDimensions() (int, int, int, int) {
	availableHeight := m.Height - 2
	if availableHeight < 4 {
		return 0, 0, 0, 0
	}

	topHeight := (availableHeight / 2) + m.VerticalSplitOffset
	if topHeight < 5 {
		topHeight = 5
	}
	if topHeight > availableHeight-5 {
		topHeight = availableHeight - 5
	}
	bottomHeight := availableHeight - topHeight

	leftWidth := ((m.Width * 6) / 10) + m.HorizontalSplitOffset
	if leftWidth < 15 {
		leftWidth = 15
	}
	if leftWidth > m.Width-15 {
		leftWidth = m.Width - 15
	}
	rightWidth := m.Width - leftWidth

	return leftWidth, rightWidth, topHeight, bottomHeight
}

type gitlabLoadedMsg struct {
	mrs       []gitlab.MergeRequest
	pipelines []gitlab.Pipeline
	issues    []gitlab.Issue
	err       error
}

func (m *AppModel) loadGitLabDataCmd() tea.Cmd {
	return func() tea.Msg {
		if m.GitLabClient == nil {
			return gitlabLoadedMsg{err: fmt.Errorf("GitLab client not initialized (check remote origin URL)")}
		}
		if m.GitLabClient.Token == "" {
			return gitlabLoadedMsg{err: fmt.Errorf("token_missing")}
		}

		mrs, err := m.GitLabClient.GetMergeRequests()
		if err != nil {
			return gitlabLoadedMsg{err: err}
		}

		pipelines, err := m.GitLabClient.GetPipelines()
		if err != nil {
			return gitlabLoadedMsg{err: err}
		}

		issues, err := m.GitLabClient.GetIssues()
		if err != nil {
			return gitlabLoadedMsg{err: err}
		}

		return gitlabLoadedMsg{
			mrs:       mrs,
			pipelines: pipelines,
			issues:    issues,
		}
	}
}

type gitlabJobsLoadedMsg struct {
	pipelineID int
	jobs       []gitlab.Job
	err        error
}

func (m *AppModel) loadGitLabJobsCmd(pipelineID int) tea.Cmd {
	return func() tea.Msg {
		jobs, err := m.GitLabClient.GetPipelineJobs(pipelineID)
		return gitlabJobsLoadedMsg{
			pipelineID: pipelineID,
			jobs:       jobs,
			err:        err,
		}
	}
}

type gitlabJobLogsLoadedMsg struct {
	jobID int
	logs  string
	err   error
}

func (m *AppModel) loadGitLabJobLogsCmd(jobID int) tea.Cmd {
	return func() tea.Msg {
		logs, err := m.GitLabClient.GetJobLogs(jobID)
		return gitlabJobLogsLoadedMsg{
			jobID: jobID,
			logs:  logs,
			err:   err,
		}
	}
}

func (m *AppModel) loadGitLabDiffCmd() tea.Cmd {
	return func() tea.Msg {
		pane := m.GitLabPane
		if pane.TotalItems() == 0 || pane.Cursor >= pane.TotalItems() {
			return nil
		}

		switch pane.ActiveTab {
		case GitLabTabMR:
			mr := pane.MRs[pane.Cursor]
			title := fmt.Sprintf("Merge Request !%d: %s", mr.IID, mr.Title)
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Title:      %s\n", mr.Title))
			sb.WriteString(fmt.Sprintf("State:      %s\n", mr.State))
			sb.WriteString(fmt.Sprintf("Author:     %s\n", mr.Author.Name))
			sb.WriteString(fmt.Sprintf("Created At: %s\n", mr.CreatedAt))
			sb.WriteString(fmt.Sprintf("Source:     %s\n", mr.SourceBranch))
			sb.WriteString(fmt.Sprintf("Target:     %s\n", mr.TargetBranch))
			sb.WriteString(fmt.Sprintf("URL:        %s\n\n", mr.WebURL))
			sb.WriteString("Description:\n")
			sb.WriteString("----------------------------------------\n")
			sb.WriteString(mr.Description)
			return diffLoadedMsg{Title: title, Content: sb.String()}

		case GitLabTabPipelines:
			p := pane.Pipelines[pane.Cursor]
			title := fmt.Sprintf("Pipeline #%d Status", p.ID)
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Pipeline ID: %d\n", p.ID))
			sb.WriteString(fmt.Sprintf("Status:      %s\n", p.Status))
			sb.WriteString(fmt.Sprintf("Ref/Branch:  %s\n", p.Ref))
			sb.WriteString(fmt.Sprintf("URL:         %s\n\n", p.WebURL))
			
			// Show jobs if loaded
			if jobs, ok := pane.JobsMap[p.ID]; ok {
				sb.WriteString("Jobs in this pipeline:\n")
				sb.WriteString("----------------------------------------\n")
				for _, job := range jobs {
					sb.WriteString(fmt.Sprintf("  - [%s] %s (Stage: %s) -> ID: %d (Press Enter to view logs)\n", job.Status, job.Name, job.Stage, job.ID))
				}
			} else {
				sb.WriteString("Press Enter to fetch jobs and log outputs.\n")
			}
			return diffLoadedMsg{Title: title, Content: sb.String()}

		case GitLabTabIssues:
			issue := pane.Issues[pane.Cursor]
			title := fmt.Sprintf("Issue #%d: %s", issue.IID, issue.Title)
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Title:      %s\n", issue.Title))
			sb.WriteString(fmt.Sprintf("State:      %s\n", issue.State))
			assigneeName := "unassigned"
			if issue.Assignee != nil {
				assigneeName = issue.Assignee.Name
			}
			sb.WriteString(fmt.Sprintf("Assignee:   %s\n", assigneeName))
			sb.WriteString(fmt.Sprintf("Created At: %s\n", issue.CreatedAt))
			sb.WriteString(fmt.Sprintf("URL:        %s\n\n", issue.WebURL))
			sb.WriteString("Description:\n")
			sb.WriteString("----------------------------------------\n")
			sb.WriteString(issue.Description)
			return diffLoadedMsg{Title: title, Content: sb.String()}
		}
		return nil
	}
}

func (m *AppModel) handleGitLabEnter() tea.Cmd {
	pane := m.GitLabPane
	if pane.TotalItems() == 0 || pane.Cursor >= pane.TotalItems() {
		return nil
	}

	if pane.ActiveTab == GitLabTabPipelines {
		p := pane.Pipelines[pane.Cursor]
		if jobs, ok := pane.JobsMap[p.ID]; ok {
			var targetJobID int
			var targetJobName string
			for _, job := range jobs {
				if job.Status == "failed" {
					targetJobID = job.ID
					targetJobName = job.Name
					break
				}
			}
			if targetJobID == 0 && len(jobs) > 0 {
				targetJobID = jobs[0].ID
				targetJobName = jobs[0].Name
			}
			if targetJobID != 0 {
				m.MetaPane.AddLog(fmt.Sprintf("Fetching logs for job %s (%d)...", targetJobName, targetJobID))
				return m.loadGitLabJobLogsCmd(targetJobID)
			}
		} else {
			m.MetaPane.AddLog(fmt.Sprintf("Loading jobs for pipeline #%d...", p.ID))
			return m.loadGitLabJobsCmd(p.ID)
		}
	}
	return nil
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
		m.loadGitLabDataCmd(),
		m.checkUpdateCmd(),
	)
}

type updateCheckFinishedMsg struct {
	version string
}

func (m *AppModel) checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest("GET", "https://api.github.com/repos/htayanloo/git360/releases/latest", nil)
		if err != nil {
			return nil
		}
		
		req.Header.Set("User-Agent", "git360-update-checker")
		resp, err := client.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil
		}

		var result struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil
		}

		return updateCheckFinishedMsg{version: result.TagName}
	}
}

type selfUpdateFinishedMsg struct {
	err error
}

func getTargetAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	if goos == "linux" && goarch == "amd64" {
		return "git360-linux-amd64"
	} else if goos == "darwin" && goarch == "amd64" {
		return "git360-darwin-amd64"
	} else if goos == "darwin" && goarch == "arm64" {
		return "git360-darwin-arm64"
	} else if goos == "windows" && goarch == "amd64" {
		return "git360-windows-amd64.exe"
	}
	return ""
}

func (m *AppModel) runSelfUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		assetName := getTargetAssetName()
		if assetName == "" {
			return selfUpdateFinishedMsg{err: fmt.Errorf("unsupported OS/architecture for auto-updates")}
		}

		downloadURL := fmt.Sprintf("https://github.com/htayanloo/git360/releases/download/%s/%s", m.LatestVersion, assetName)

		client := &http.Client{Timeout: 60 * time.Second}
		req, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			return selfUpdateFinishedMsg{err: err}
		}
		
		req.Header.Set("User-Agent", "git360-updater")
		resp, err := client.Do(req)
		if err != nil {
			return selfUpdateFinishedMsg{err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return selfUpdateFinishedMsg{err: fmt.Errorf("failed to download asset: HTTP %d", resp.StatusCode)}
		}

		exePath, err := os.Executable()
		if err != nil {
			return selfUpdateFinishedMsg{err: err}
		}

		// Write to a temp file in the same directory (ensures same filesystem partition for atomic rename)
		tmpFile, err := os.CreateTemp(filepath.Dir(exePath), "git360-update-")
		if err != nil {
			return selfUpdateFinishedMsg{err: err}
		}
		defer os.Remove(tmpFile.Name())

		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			tmpFile.Close()
			return selfUpdateFinishedMsg{err: err}
		}
		tmpFile.Close()

		// Set executable permissions
		err = os.Chmod(tmpFile.Name(), 0755)
		if err != nil {
			return selfUpdateFinishedMsg{err: err}
		}

		// Overwrite the current binary
		err = os.Rename(tmpFile.Name(), exePath)
		if err != nil {
			return selfUpdateFinishedMsg{err: err}
		}

		return selfUpdateFinishedMsg{err: nil}
	}
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
			if m.ShowGitLab {
				m.GitLabPane.IsLoading = true
				cmds = append(cmds, m.loadGitLabDataCmd())
			}

		case "f":
			// Manual fetch
			if !m.IsFetching {
				m.IsFetching = true
				m.MetaPane.StatusText = "Fetching origin..."
				m.MetaPane.AddLog("Fetching origin in background...")
				cmds = append(cmds, m.fetchOriginCmd())
			}

		case "g":
			if m.ShowGitLab {
				m.ShowGitLab = false
				cmds = append(cmds, m.loadDiffCmd())
			} else {
				m.ShowGitLab = true
				m.ShowBranchBrowser = false
				m.ActivePane = PaneStatus
				m.GitLabPane.IsLoading = true
				m.MetaPane.AddLog("Loading GitLab dashboard...")
				cmds = append(cmds, m.loadGitLabDataCmd())
			}

		case "h":
			if m.ShowGitLab && m.ActivePane == PaneStatus {
				m.GitLabPane.PrevTab()
				cmds = append(cmds, m.loadGitLabDiffCmd())
			}

		case "l":
			if m.ShowGitLab && m.ActivePane == PaneStatus {
				m.GitLabPane.NextTab()
				cmds = append(cmds, m.loadGitLabDiffCmd())
			}

		case "U":
			if m.LatestVersion != "" && m.LatestVersion != Version && !m.IsUpdating {
				m.IsUpdating = true
				m.MetaPane.StatusText = "Updating app..."
				m.MetaPane.AddLog(fmt.Sprintf("Downloading and installing %s...", m.LatestVersion))
				cmds = append(cmds, m.runSelfUpdateCmd())
			}

		case "b":
			// Toggle branch browser
			if m.ShowBranchBrowser && !m.BranchBrowser.IsTags {
				m.ShowBranchBrowser = false
				cmds = append(cmds, m.loadDiffCmd())
			} else {
				m.ShowBranchBrowser = true
				m.ShowGitLab = false
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
				m.ShowGitLab = false
				m.ActivePane = PaneStatus
				m.MetaPane.AddLog("Browsing tags...")
				cmds = append(cmds, m.loadTagsCmd())
			}

		case "escape":
			if m.ShowGitLab {
				m.ShowGitLab = false
				cmds = append(cmds, m.loadDiffCmd())
			} else if m.ShowBranchBrowser {
				m.ShowBranchBrowser = false
				cmds = append(cmds, m.loadDiffCmd())
			}

		case "enter":
			if m.ShowGitLab && m.ActivePane == PaneStatus {
				cmds = append(cmds, m.handleGitLabEnter())
			} else if m.ShowBranchBrowser && m.ActivePane == PaneStatus {
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
				if m.ShowGitLab {
					m.GitLabPane.ScrollUp()
					cmds = append(cmds, m.loadGitLabDiffCmd())
				} else if m.ShowBranchBrowser {
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
				if m.ShowGitLab {
					m.GitLabPane.ScrollDown()
					cmds = append(cmds, m.loadGitLabDiffCmd())
				} else if m.ShowBranchBrowser {
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

		case "ctrl+left", "alt+left", "shift+left":
			m.HorizontalSplitOffset--
			cmds = append(cmds, m.loadDiffCmd())

		case "ctrl+right", "alt+right", "shift+right":
			m.HorizontalSplitOffset++
			cmds = append(cmds, m.loadDiffCmd())

		case "ctrl+up", "alt+up", "shift+up":
			m.VerticalSplitOffset--
			cmds = append(cmds, m.loadDiffCmd())

		case "ctrl+down", "alt+down", "shift+down":
			m.VerticalSplitOffset++
			cmds = append(cmds, m.loadDiffCmd())

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

	case gitlabLoadedMsg:
		m.GitLabPane.IsLoading = false
		if msg.err != nil {
			if msg.err.Error() == "token_missing" {
				m.GitLabPane.TokenMissing = true
			} else {
				m.GitLabPane.Error = msg.err
				m.MetaPane.AddLog(fmt.Sprintf("GitLab error: %v", msg.err))
			}
		} else {
			m.GitLabPane.MRs = msg.mrs
			m.GitLabPane.Pipelines = msg.pipelines
			m.GitLabPane.Issues = msg.issues
			m.GitLabPane.Error = nil
			m.GitLabPane.TokenMissing = false
			if m.ShowGitLab {
				cmds = append(cmds, m.loadGitLabDiffCmd())
			}
		}

	case gitlabJobsLoadedMsg:
		if msg.err != nil {
			m.MetaPane.AddLog(fmt.Sprintf("GitLab jobs error: %v", msg.err))
		} else {
			m.GitLabPane.JobsMap[msg.pipelineID] = msg.jobs
			m.MetaPane.AddLog(fmt.Sprintf("Loaded %d jobs for pipeline #%d", len(msg.jobs), msg.pipelineID))
			cmds = append(cmds, m.loadGitLabDiffCmd())
		}

	case gitlabJobLogsLoadedMsg:
		if msg.err != nil {
			m.MetaPane.AddLog(fmt.Sprintf("GitLab logs error: %v", msg.err))
		} else {
			m.DiffPane.SetContent(fmt.Sprintf("CI/CD Job %d Logs", msg.jobID), msg.logs, m.Styles)
		}

	case updateCheckFinishedMsg:
		m.LatestVersion = msg.version

	case selfUpdateFinishedMsg:
		m.IsUpdating = false
		if msg.err != nil {
			m.MetaPane.AddLog(fmt.Sprintf("Update failed: %v", msg.err))
			m.MetaPane.StatusText = "Update failed"
		} else {
			m.MetaPane.AddLog("Update installed successfully! Please restart git360.")
			m.MetaPane.StatusText = "Restart required"
		}

	case errorMsg:
		m.Err = msg.err
		m.MetaPane.AddLog(fmt.Sprintf("Error: %v", msg.err))

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case tea.MouseMsg:
		leftWidth, _, topHeight, _ := m.layoutDimensions()
		if leftWidth > 0 && topHeight > 0 {
			isTopRow := msg.Y >= 1 && msg.Y <= topHeight
			isBottomRow := msg.Y >= topHeight+1 && msg.Y <= m.Height-2
			isLeftCol := msg.X >= 0 && msg.X < leftWidth
			isRightCol := msg.X >= leftWidth && msg.X < m.Width

			// Handle clicks to change active pane
			if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
				if isTopRow && isLeftCol {
					m.ActivePane = PaneGraph
					cmds = append(cmds, m.loadDiffCmd())
				} else if isTopRow && isRightCol {
					m.ActivePane = PaneStatus
					cmds = append(cmds, m.loadDiffCmd())
				} else if isBottomRow && isLeftCol {
					m.ActivePane = PaneDiff
				}
			}

			// Handle mouse wheel scrolling
			if msg.Button == tea.MouseButtonWheelUp {
				if isTopRow && isLeftCol {
					m.GraphPane.ScrollUp()
					cmds = append(cmds, m.loadDiffCmd())
				} else if isTopRow && isRightCol {
					if m.ShowGitLab {
						m.GitLabPane.ScrollUp()
						cmds = append(cmds, m.loadGitLabDiffCmd())
					} else if m.ShowBranchBrowser {
						m.BranchBrowser.ScrollUp()
						selected := m.BranchBrowser.SelectedItem()
						if selected != "" {
							cmds = append(cmds, m.compareRefCmd(selected))
						}
					} else {
						m.StatusPane.ScrollUp()
						cmds = append(cmds, m.loadDiffCmd())
					}
				} else if isBottomRow && isLeftCol {
					m.DiffPane.Viewport.LineUp(3)
				}
			} else if msg.Button == tea.MouseButtonWheelDown {
				if isTopRow && isLeftCol {
					m.GraphPane.ScrollDown()
					cmds = append(cmds, m.loadDiffCmd())
				} else if isTopRow && isRightCol {
					if m.ShowGitLab {
						m.GitLabPane.ScrollDown()
						cmds = append(cmds, m.loadGitLabDiffCmd())
					} else if m.ShowBranchBrowser {
						m.BranchBrowser.ScrollDown()
						selected := m.BranchBrowser.SelectedItem()
						if selected != "" {
							cmds = append(cmds, m.compareRefCmd(selected))
						}
					} else {
						m.StatusPane.ScrollDown()
						cmds = append(cmds, m.loadDiffCmd())
					}
				} else if isBottomRow && isLeftCol {
					m.DiffPane.Viewport.LineDown(3)
				}
			}
		}
	}

	// Delegate diff pane update if active (for scrolling diff viewport)
	if m.ActivePane == PaneDiff {
		// Do not pass mouse messages to avoid duplicate wheel scroll handling,
		// or viewport scrolling when mouse is over other panes.
		if _, isMouse := msg.(tea.MouseMsg); !isMouse {
			diffCmd := m.DiffPane.Update(msg)
			cmds = append(cmds, diffCmd)
		}
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

	leftWidth, rightWidth, topHeight, bottomHeight := m.layoutDimensions()
	if leftWidth == 0 || topHeight == 0 {
		return "Terminal window is too small!"
	}

	// Style title headers for each pane
	paneTitle := func(title string, pane Pane) string {
		if m.ActivePane == pane {
			return m.Styles.PaneTitleActive.Render("█ " + title)
		}
		return m.Styles.PaneTitleInactive.Render("░ " + title)
	}

	// 1. Render Top Row
	topInnerHeight := topHeight - 4
	if topInnerHeight < 1 {
		topInnerHeight = 1
	}

	graphContent := m.GraphPane.View(m.Styles, leftWidth-4, topInnerHeight, m.ActivePane == PaneGraph)
	graphBox := m.Styles.InactivePaneBorder.Copy()
	if m.ActivePane == PaneGraph {
		graphBox = m.Styles.ActivePaneBorder.Copy()
	}
	graphView := graphBox.
		Width(leftWidth).
		Height(topHeight).
		Render(paneTitle("GIT HISTORY GRAPH", PaneGraph) + "\n" + graphContent)

	// Determine status pane contents and title dynamically
	var statusContent string
	var statusTitle string
	if m.ShowGitLab {
		statusContent = m.GitLabPane.View(m.Styles, rightWidth-4, topInnerHeight)
		statusTitle = "GITLAB DASHBOARD (h/l=Tabs, Enter=View, Esc=Close)"
	} else if m.ShowBranchBrowser {
		statusContent = m.BranchBrowser.View(m.Styles, rightWidth-4, topInnerHeight)
		if m.BranchBrowser.IsTags {
			statusTitle = "TAG BROWSER (Enter=Checkout, c/Scroll=Compare, Esc=Close)"
		} else {
			statusTitle = "BRANCH BROWSER (Enter=Checkout, c/Scroll=Compare, Esc=Close)"
		}
	} else {
		statusContent = m.StatusPane.View(m.Styles, rightWidth-4, topInnerHeight, m.ActivePane == PaneStatus)
		statusTitle = "STAGE CONTROL (Space=Toggle, s=Stage, u=Unstage)"
	}

	statusBox := m.Styles.InactivePaneBorder.Copy()
	if m.ActivePane == PaneStatus {
		statusBox = m.Styles.ActivePaneBorder.Copy()
	}
	statusView := statusBox.
		Width(rightWidth).
		Height(topHeight).
		Render(paneTitle(statusTitle, PaneStatus) + "\n" + statusContent)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, graphView, statusView)

	// 2. Render Bottom Row
	bottomInnerHeight := bottomHeight - 4
	if bottomInnerHeight < 1 {
		bottomInnerHeight = 1
	}

	// Resize diff viewport to fit its container
	m.DiffPane.Resize(leftWidth-4, bottomInnerHeight)
	diffContent := m.DiffPane.View()
	diffBox := m.Styles.InactivePaneBorder.Copy()
	if m.ActivePane == PaneDiff {
		diffBox = m.Styles.ActivePaneBorder.Copy()
	}
	diffView := diffBox.
		Width(leftWidth).
		Height(bottomHeight).
		Render(paneTitle(m.DiffPane.Title, PaneDiff) + "\n" + diffContent)

	metaContent := m.MetaPane.View(m.Styles, rightWidth-4, bottomInnerHeight)
	metaBox := m.Styles.InactivePaneBorder.Copy() // Meta pane is informational, never directly focused
	metaView := metaBox.
		Width(rightWidth).
		Height(bottomHeight).
		Render(m.Styles.PaneTitleInactive.Render("░ AGENT METRICS & SYSTEM LOGS") + "\n" + metaContent)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, diffView, metaView)

	// 3. Assemble Header and Footer
	updateNotify := ""
	if m.LatestVersion != "" && m.LatestVersion != Version {
		updateNotify = fmt.Sprintf("  |  %s", m.Styles.HelpKeyStyle.Render("✨ Update available: "+m.LatestVersion))
	}
	headerText := fmt.Sprintf(" GIT-360  |  Active Branch: %s  |  Last update: %s%s", 
		m.MetaPane.Branch, 
		m.LastRefresh.Format("15:04:05"),
		updateNotify,
	)
	header := m.Styles.HeaderStyle.Width(m.Width).Render(headerText)

	// Help guidelines footer
	var helpParts []string
	helpParts = append(helpParts, fmt.Sprintf("%s Switch Pane", m.Styles.HelpKeyStyle.Render("Tab")))
	helpParts = append(helpParts, fmt.Sprintf("%s/%s Scroll", m.Styles.HelpKeyStyle.Render("↑↓"), m.Styles.HelpKeyStyle.Render("j/k")))
	helpParts = append(helpParts, fmt.Sprintf("%s Resize", m.Styles.HelpKeyStyle.Render("Shift+Arrows")))
	
	if m.ShowBranchBrowser {
		helpParts = append(helpParts, fmt.Sprintf("%s Checkout", m.Styles.HelpKeyStyle.Render("Enter")))
		helpParts = append(helpParts, fmt.Sprintf("%s Compare", m.Styles.HelpKeyStyle.Render("c")))
		helpParts = append(helpParts, fmt.Sprintf("%s Close", m.Styles.HelpKeyStyle.Render("Esc")))
	} else if m.ShowGitLab {
		helpParts = append(helpParts, fmt.Sprintf("%s/%s Tab", m.Styles.HelpKeyStyle.Render("h"), m.Styles.HelpKeyStyle.Render("l")))
		helpParts = append(helpParts, fmt.Sprintf("%s View Logs/Details", m.Styles.HelpKeyStyle.Render("Enter")))
		helpParts = append(helpParts, fmt.Sprintf("%s Close", m.Styles.HelpKeyStyle.Render("Esc")))
	} else {
		helpParts = append(helpParts, fmt.Sprintf("%s Toggle Stage", m.Styles.HelpKeyStyle.Render("Space")))
		helpParts = append(helpParts, fmt.Sprintf("%s Stage", m.Styles.HelpKeyStyle.Render("s")))
		helpParts = append(helpParts, fmt.Sprintf("%s Unstage", m.Styles.HelpKeyStyle.Render("u")))
	}
	helpParts = append(helpParts, fmt.Sprintf("%s GitLab", m.Styles.HelpKeyStyle.Render("g")))
	
	if m.LatestVersion != "" && m.LatestVersion != Version {
		helpParts = append(helpParts, fmt.Sprintf("%s Update App", m.Styles.HelpKeyStyle.Render("Shift+U")))
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
