# Git-360: Real-Time Git Observer & TUI Dashboard

`Git-360` is a keyboard-driven terminal user interface (TUI) written in Go. It provides a **360-degree, real-time snapshot** of a Git repository. 

It is specifically designed for developers working with background automation or running LLM programming agents (e.g., in separate terminal or SSH sessions). It allows you to watch workspace status transitions, stage/unstage files, view colorized diffs, and inspect commit graphs in real-time as changes occur.

---

## 📐 Layout & Visual Design

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
| @@ -14,3 +14,5 @@                                       | - Last check: 14:05:01               |
|  - Old line here                                        | - Status: Watching changes...        |
|  + New layout spec added                                | - Log: [14:04:12] File main.go changed |
+---------------------------------------------------------+--------------------------------------+
| [Tab] Switch Pane | [S] Stage | [U] Unstage | [F] Fetch Origin | [D] Toggle Diff Style | [Q] Quit |
+------------------------------------------------------------------------------------------------+
```

---

## ⚡ Quick Start

### Prerequisites
- Go 1.21 or higher
- Git command-line utility installed and available in your system `$PATH`

### Building from Source
Initialize dependencies and build the binary:
```bash
# Get dependencies
go mod tidy

# Build executable
go build -o git360 ./cmd/git360
```

### Running the App
Run inside your current repository:
```bash
./git360
```

Or pass a path to target another repository:
```bash
./git360 -dir /path/to/other/repo
```

---

## ⌨️ Keybindings

| Key | Context | Action |
|---|---|---|
| `Tab` | Global | Cycle active focus clockwise (Graph $\rightarrow$ Status $\rightarrow$ Diff) |
| `Shift+Tab` | Global | Cycle active focus counter-clockwise |
| `↑` / `↓` or `j` / `k` | List Panes | Scroll up/down through commits, files, or GitLab items |
| `Shift + Arrow keys` | Global | Dynamically resize the vertical/horizontal split layout |
| `Space` | Status Pane | Toggle staged/unstaged state of the selected file |
| `s` | Status Pane | Stage selected file |
| `u` | Status Pane | Unstage selected file |
| `g` | Global | Toggle GitLab Dashboard (MRs, Pipelines, Issues) |
| `h` / `l` | GitLab Pane | Switch between GitLab sub-tabs (MRs $\leftrightarrow$ Pipelines $\leftrightarrow$ Issues) |
| `Enter` | GitLab Pane | Load MR description or Pipeline Jobs list in Diff Viewer |
| `Enter` (on Pipeline Jobs) | GitLab Pane | Fetch and load trace logs of the first failed job |
| `Shift + U` | Global | Perform in-app automatic update to the latest release version |
| `f` | Global | Trigger a background remote `git fetch` (non-blocking) |
| `r` | Global | Force a manual UI and repository/GitLab state refresh |
| `q` or `Ctrl+C` | Global | Exit application |

---

## 🦊 GitLab Integration Config

To authenticate with GitLab, set the `GITLAB_TOKEN` environment variable before running the application:

```bash
export GITLAB_TOKEN="your_personal_access_token"
```

The application automatically parses the remote origin URL of your repository (supporting standard HTTPS/SSH URLs and self-hosted GitLab instances) to connect to the correct project.

---

## 🏗️ Architecture Overview

The codebase is split into modular packages: the Git shell driver, the GitLab API client, and the Bubble Tea TUI render loop.

```
golang-git-graph/
├── cmd/
│   └── git360/
│       └── main.go         # Bootstraps the application, handles path CLI flags
├── internal/
│   ├── git/
│   │   ├── types.go        # Git structures (Commit, FileChange, RepoState)
│   │   ├── client.go       # os/exec wrappers running commands in the repo path
│   │   └── parser.go       # Parses porcelain status outputs and graph structures
│   ├── gitlab/
│   │   ├── client.go       # Lightweight REST API client for GitLab (MRs, CI/CD, Issues)
│   │   └── client_test.go  # Unit tests for GitLab remote URL parsing
│   └── tui/
│       ├── app.go          # Central event-loop coordinator and layout generator
│       ├── styles.go       # Lipgloss Dracula-themed color palette and borders
│       ├── pane_graph.go   # Commit History list view
│       ├── pane_status.go  # Working tree file list view
│       ├── pane_diff.go    # Highlighted file diff rendering using viewport
│       ├── pane_gitlab.go  # GitLab MRs, pipelines, and issues view
│       └── pane_meta.go    # Live activity log and metrics view
```

### Key Technical Decisions
1. **CLI Commands vs. Native Libraries**: `os/exec` wraps the standard `git` CLI instead of utilizing pure Go libraries (like `go-git`). This guarantees full compatibility with system aliases, submodules, configurations, and renders the visual graph structures natively.
2. **Comparison-Based Logging**: The TUI maintains a cache map of the previous file states. During every poll cycle (every 2 seconds), it compares the old map with the new map to identify when files are created, modified, staged, or committed, printing an event feed in the **System Logs** panel.
3. **Null-Delimiter Log Parsing**: Commits are formatted using `%x00` null byte markers during CLI invocation. This isolates the commit hash, author, date, subject, and ref values, preventing parsing errors when branches merge or tree connection lines contain special characters like `|` or `/`.
4. **Lightweight GitLab Integration**: Uses Go's standard library `net/http` to build a zero-dependency GitLab API wrapper, allowing real-time monitoring of pipelines, merge requests, and issue tracking.

---

## 🚀 Future Roadmap & Next Steps

When you resume development of `Git-360`, here are the recommended features to implement next:

1. **Commit Creation Modal**
   - Bind the `c` key in the status pane to open a pop-up text input dialog allowing users to enter a commit message and run `git commit -m "..."` directly.
2. **Unified/Side-by-Side Diff Toggle**
   - Extend `pane_diff.go` to split line segments into left/right arrays and render them side-to-side using horizontal Lip Gloss blocks.
3. **Branch Checking & Checkout**
   - Bind `b` to display a list of local branches. Selecting a branch and hitting `Enter` triggers `git checkout <branch>`.
4. **Interactive Merge Conflict Resolver**
   - Detect files with conflicting markers (`UU` in porcelain status) and parse diff blocks to let the user choose standard incoming/current changes directly.
5. **Configurable Polling Intervals**
   - Provide a CLI flag (e.g. `--tick 1s`) to customize refresh frequencies.
