package git

import "time"

// Commit represents a single commit in the history with its associated graph visualization text.
type Commit struct {
	Hash       string // Abbreviated or full hash
	Author     string
	Date       string // Relative date (e.g. "2 hours ago")
	AbsDate    string // Short absolute date (e.g. "2026-07-09")
	Subject    string
	Refs       string // e.g. "(HEAD -> main, origin/main)"
	GraphChars string // The graph prefix characters (e.g. "*", "|", "| \")
	RawLine    string // The complete raw line for fallback
}

// FileChange represents a modified, added, deleted, or untracked file in the workspace.
type FileChange struct {
	Path     string
	OldPath  string // Used for renames
	Status   string // "A", "M", "D", "R", "C", "?", etc.
	Staged   bool   // True if staged (index), false if unstaged (worktree)
}

// RepoState contains the full snapshot of the repository at a point in time.
type RepoState struct {
	Branch        string
	Upstream      string
	Ahead         int
	Behind        int
	FileChanges   []FileChange
	Commits       []Commit
	LastFetchTime time.Time
}
