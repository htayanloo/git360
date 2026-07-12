# Git-360: Golang Git Graph & Real-Time TUI Dashboard

`Git-360` is a modern, beautiful terminal user interface (TUI) written in Go. It is designed to provide a **360-degree real-time view** of your Git repository. It is especially optimized for developers who run background tasks or automated LLM agents in separate terminal/SSH sessions and want to see what is happening in the workspace instantly.

---

## 🌟 Key Features

1. **Beautiful Git Graph Tree**
   - Colorized branch lanes, commit tags, and commit authors.
   - Interactive commit scrolling. Selecting a commit details the changes in that commit.
2. **Split-Screen Dashboard Layout**
   - **Pane A (Git Graph):** Displays the visual commit history graph.
   - **Pane B (Git Status):** Lists staged and unstaged file modifications.
   - **Pane C (Diff Viewer):** Displays a scrollable, color-coded diff (unified or side-by-side) of the selected file or commit.
   - **Pane D (Origin Status):** Dedicated panel displaying local vs. remote tracking differences, checking if `origin` has new commits in the background.
3. **Agent-Activity Watcher (Auto-Refresh)**
   - Background poller that monitors workspace modifications and staging events.
   - Auto-fetches remote status (safely using `git fetch` / `git remote` checks in a separate background routine) to warn if origin has diverged.
4. **Interactive File/Stage Navigation**
   - Keyboard shortcuts to stage (`s`), unstage (`u`), or toggle stage (`Space`) on files.
   - Instant side-by-side or inline diff viewing of staged/unstaged changes.

---

## 📐 UI Layout Architecture

The TUI is divided into 4 main regions, styled with premium borders, colors, and responsive layouts that adapt to terminal resizing.

```
+------------------------------------------------------------------------------------------------+
|  GIT-360 | Repo: golang-git-graph | Branch: main | Origin: 1 commit ahead, 0 behind            |
+---------------------------------------------------------+--------------------------------------+
| [1] COMMIT GRAPH & HISTORY                              | [2] WORKING DIRECTORY STATUS         |
| * (HEAD -> main) feat: added project.md structure       | Staged Changes:                      |
| * origin/main docs: update README                       |  [x] project.md                      |
| * merge branch 'fix-bugs'                               | Unstaged Changes:                    |
| |\                                                      |  [ ] main.go                         |
| | * fix: handle remote origin nil check                 |  [ ] git/client.go                   |
| | * test: add test cases for diff parser                |                                      |
+---------------------------------------------------------+--------------------------------------+
| [3] DIFF VIEWER (unified/side-by-side)                  | [4] SYSTEM METRICS & LOGS            |
| --- a/project.md                                        | - Active branch: main                |
| +++ b/project.md                                        | - Tracking: origin/main              |
| @@ -14,3 +14,5 @@                                       | - Last fetch: 10 seconds ago         |
|  - Old line here                                        | - LLM Agent Activity: Idle           |
|  + New layout spec added                                | - Status: Polling active             |
+---------------------------------------------------------+--------------------------------------+
| [Tab] Switch Pane | [S] Stage | [U] Unstage | [F] Fetch Origin | [D] Toggle Diff Style | [Q] Quit |
+------------------------------------------------------------------------------------------------+
```

---

## 🛠️ Technical Stack & Dependencies

- **Language:** Go 1.21+
- **TUI Framework:**
  - `github.com/charmbracelet/bubbletea` — Elm Architecture implementation for event loops and rendering.
  - `github.com/charmbracelet/lipgloss` — Advanced style definitions, margins, borders, and colors (using premium HSL/ANSI color scales).
  - `github.com/charmbracelet/bubbles` — Pre-built components like lists, viewports, and text inputs.
- **Git Engine:**
  - Custom Go wrappers executing CLI `git` processes via `os/exec`. This provides the best performance, speed, and accuracy compared to pure-Go packages, which can be slow and missing features like `git graph` rendering.

---

## 🗂️ Project Directory Structure

```
golang-git-graph/
├── cmd/
│   └── git360/
│       └── main.go         # Entry point for the CLI
├── internal/
│   ├── git/
│   │   ├── client.go       # Shell wrappers for status, graph, diff, and remote commands
│   │   ├── parser.go       # Parsers for git status, git log graph, and diff outputs
│   │   └── types.go        # Shared Git data structs (Commit, FileChange, DiffLine)
│   └── tui/
│       ├── app.go          # Core Bubble Tea Model & Update/View loops
│       ├── styles.go       # Lipgloss color definitions and borders
│       ├── pane_graph.go   # Graph pane sub-model and layout
│       ├── pane_status.go  # Status & files list pane sub-model
│       ├── pane_diff.go    # Colorized diff pane (handling scrolling and toggle view)
│       └── pane_meta.go    # Metadata & poller summary pane
├── go.mod                  # Modules configuration
├── go.sum
└── project.md              # Project plan & design documentation (This File)
```

---

## 🚀 Implementation Phases

### Phase 1: Go Git Engine & Client Setup
- Initialize the Go module.
- Build the `git` client wrapping CLI calls:
  - `git status --porcelain` to fetch staged, unstaged, untracked files.
  - `git log --graph --oneline --decorate --color=never` to get raw commit trees.
  - `git diff` & `git diff --cached` to retrieve unified file diffs.
  - `git fetch` and `git status -sb` to count ahead/behind commits compared to tracking branch.

### Phase 2: Bubble Tea Boilerplate & Split Layout
- Setup basic `bubbletea` structure (`Init`, `Update`, `View`).
- Define the main TUI dimensions and layout grids using `lipgloss`.
- Implement basic pane focus system using the `Tab` key. Focused panes get a vibrant border, unfocused panes have a subtle dimmed border.

### Phase 3: Interactive Git Graph
- Build the Commit Graph rendering engine using the parsed `git log --graph` results.
- Implement scrollable navigation on the commit list using arrow keys or `j`/`k`.
- Trigger diff extraction for the selected commit when focused.

### Phase 4: Status Pane & File Staging
- Display lists of staged, unstaged, and untracked files.
- Enable file navigation and toggle staging status using keyboard commands.
- Highlight conflicts and special git states (merging, rebasing).

### Phase 5: Side-by-Side and Unified Diff Viewer
- Build a dedicated text view pane for file diffs.
- Highlight deleted lines in red and added lines in green.
- Add support for unified and side-by-side presentation toggling.

### Phase 6: Agent Activity Watcher & Auto-Poll
- Run a background goroutine ticking every 2 seconds.
- Trigger non-blocking state updates (e.g., checks if index has changed, checks for modifications).
- Implement background fetching for `origin` updates with visual status indicator.

---

## ⌨️ Keybindings Reference

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus clockwise through panels (Graph -> Status -> Diff -> Logs) |
| `Shift+Tab`| Cycle focus counter-clockwise |
| `↑` / `↓` or `j` / `k` | Navigate within lists / commits / diffs |
| `Space` | Toggle stage/unstage status on selected file |
| `s` | Stage current file |
| `u` | Unstage current file |
| `d` | Toggle Diff View mode (Unified vs. Side-by-Side) |
| `f` | Fetch origin immediately |
| `r` | Hard refresh TUI data |
| `q` or `Ctrl+C` | Quit the application |
