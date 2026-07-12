package git

import (
	"strings"
)

// ParseStatus parses the output of "git status --porcelain -b".
// It returns the local branch name, upstream name, ahead/behind counts, and lists of staged and unstaged file changes.
func ParseStatus(output string) (branch string, upstream string, ahead int, behind int, changes []FileChange) {
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", "", 0, 0, nil
	}

	// Parse first line (branch info, starts with "## ")
	header := lines[0]
	if strings.HasPrefix(header, "## ") {
		header = strings.TrimPrefix(header, "## ")
		// Header can be like:
		// main...origin/main [ahead 1, behind 2]
		// main...origin/main
		// main
		// HEAD (no branch)
		
		// Check for upstream
		if idx := strings.Index(header, "..."); idx != -1 {
			branch = header[:idx]
			rest := header[idx+3:]
			
			// Check for ahead/behind metadata
			if metaIdx := strings.Index(rest, " ["); metaIdx != -1 {
				upstream = rest[:metaIdx]
				meta := rest[metaIdx+2 : len(rest)-1] // strip " [" and "]"
				
				// Parse ahead/behind
				parts := strings.Split(meta, ", ")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.HasPrefix(part, "ahead ") {
						var val int
						_, _ = parserScan(strings.TrimPrefix(part, "ahead "), &val)
						ahead = val
					} else if strings.HasPrefix(part, "behind ") {
						var val int
						_, _ = parserScan(strings.TrimPrefix(part, "behind "), &val)
						behind = val
					}
				}
			} else {
				upstream = rest
			}
		} else {
			// No upstream
			if strings.Contains(header, "Initial commit on") {
				parts := strings.Fields(header)
				if len(parts) >= 4 {
					branch = parts[3]
				}
			} else if strings.Contains(header, "no branch") {
				branch = "HEAD (detached)"
			} else {
				branch = strings.Fields(header)[0]
			}
		}
	}

	// Parse file changes from the rest of the lines
	for _, line := range lines[1:] {
		if len(line) < 4 {
			continue
		}
		
		// Status is in the first 2 characters
		x := line[0] // Index status
		y := line[1] // Worktree status
		
		// File path starts at index 3
		pathPart := line[3:]
		
		// Handle renames: e.g. "R  oldpath -> newpath"
		var path, oldPath string
		if x == 'R' || y == 'R' {
			if arrowIdx := strings.Index(pathPart, " -> "); arrowIdx != -1 {
				oldPath = strings.Trim(pathPart[:arrowIdx], "\"")
				path = strings.Trim(pathPart[arrowIdx+4:], "\"")
			} else {
				path = strings.Trim(pathPart, "\"")
			}
		} else {
			path = strings.Trim(pathPart, "\"")
		}

		// A file can have changes in both staged index and unstaged worktree.
		// We'll create separate entries for staged and unstaged status, as planned, to make staging navigation clear.
		
		// If staged (index status is not ' ' and not '?')
		if x != ' ' && x != '?' {
			changes = append(changes, FileChange{
				Path:    path,
				OldPath: oldPath,
				Status:  string(x),
				Staged:  true,
			})
		}
		
		// If unstaged (worktree status is not ' ' or it is untracked '??')
		if y != ' ' || (x == '?' && y == '?') {
			status := string(y)
			if x == '?' && y == '?' {
				status = "?"
			}
			changes = append(changes, FileChange{
				Path:    path,
				OldPath: oldPath,
				Status:  status,
				Staged:  false,
			})
		}
	}

	return branch, upstream, ahead, behind, changes
}

// Simple integer scanner to avoid fmt.Sscanf dependency overhead
func parserScan(s string, val *int) (int, error) {
	var n int
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int(s[i]-'0')
		} else {
			break
		}
	}
	*val = n
	return 1, nil
}

// ParseCommits parses the lines of output from git log command.
// Expected format: git log --graph --pretty=format:"%h%x00%an%x00%ar%x00%as%x00%s%x00%D"
func ParseCommits(output string) []Commit {
	var commits []Commit
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Split(line, "\x00")
		if len(parts) >= 6 {
			// This is a commit line
			graphAndHash := parts[0]
			author := parts[1]
			date := parts[2]
			absDate := parts[3]
			subject := parts[4]
			refs := parts[5]
			
			// Separate graph symbols and hash from Part 0
			// The hash is the last word in graphAndHash
			fields := strings.Fields(graphAndHash)
			var hash, graphChars string
			if len(fields) > 0 {
				hash = fields[len(fields)-1]
				// Find where the hash starts in the string to get the graph prefix characters accurately
				hashIdx := strings.LastIndex(graphAndHash, hash)
				if hashIdx > 0 {
					graphChars = graphAndHash[:hashIdx]
				}
			} else {
				hash = ""
				graphChars = graphAndHash
			}
			
			commits = append(commits, Commit{
				Hash:       hash,
				Author:     author,
				Date:       date,
				AbsDate:    absDate,
				Subject:    subject,
				Refs:       refs,
				GraphChars: graphChars,
				RawLine:    line,
			})
		} else {
			// This is a pure graph line (e.g. connections, merges, lines)
			commits = append(commits, Commit{
				GraphChars: line,
				RawLine:    line,
			})
		}
	}
	
	return commits
}
