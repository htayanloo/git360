package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Client handles executing git CLI commands and parsing the results.
type Client struct {
	RepoDir string
}

// NewClient creates a new Git client for the specified repository directory.
func NewClient(repoDir string) *Client {
	return &Client{RepoDir: repoDir}
}

// runCmd helper executes a git command in the repo directory and returns output.
func (c *Client) runCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.RepoDir
	
	// Keep environment variables but enforce PAGER=cat and standard encoding
	cmd.Env = append(os.Environ(), "PAGER=cat", "LANG=C", "LC_ALL=C")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetStatus retrieves the git status details.
func (c *Client) GetStatus() (branch string, upstream string, ahead int, behind int, changes []FileChange, err error) {
	out, err := c.runCmd("status", "--porcelain", "-b")
	if err != nil {
		return "", "", 0, 0, nil, err
	}
	
	branch, upstream, ahead, behind, changes = ParseStatus(out)
	return branch, upstream, ahead, behind, changes, nil
}

// GetCommits retrieves the commit history graph.
func (c *Client) GetCommits(limit int) ([]Commit, error) {
	// We use the null-byte separator trick to parse graph structure correctly
	// Format fields: hash (%h), author (%an), relative date (%ar), absolute date (%as), subject (%s), refs (%D)
	args := []string{
		"log",
		"--graph",
		"--all", // Show all branches and tags in a colorful graph
		"--pretty=format:%h%x00%an%x00%ar%x00%as%x00%s%x00%D",
		"--abbrev-commit",
		"-n", fmt.Sprintf("%d", limit),
	}
	
	out, err := c.runCmd(args...)
	if err != nil {
		return nil, err
	}
	
	return ParseCommits(out), nil
}

// GetDiff retrieves the diff for a modified file.
func (c *Client) GetDiff(change FileChange) (string, error) {
	// If the file is untracked (status is '?'), git diff won't show it directly.
	// We read its contents and construct a virtual diff of added lines.
	if change.Status == "?" {
		filePath := change.Path
		if c.RepoDir != "" {
			filePath = c.RepoDir + "/" + change.Path
		}
		
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()
		
		// Read a reasonable amount of data for diff viewing
		buf := make([]byte, 1024*64)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		
		content := string(buf[:n])
		lines := strings.Split(content, "\n")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n", change.Path, len(lines)))
		for _, line := range lines {
			sb.WriteString("+" + line + "\n")
		}
		return sb.String(), nil
	}

	var args []string
	if change.Staged {
		args = []string{"diff", "--cached", "--color=never", "--", change.Path}
	} else {
		args = []string{"diff", "--color=never", "--", change.Path}
	}
	
	return c.runCmd(args...)
}

// GetCommitDiff retrieves the diff of changes introduced by a specific commit.
func (c *Client) GetCommitDiff(hash string) (string, error) {
	parents, err := c.runCmd("rev-list", "--parents", "-n", "1", hash)
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(parents))
		if len(parts) > 2 {
			// It's a merge commit (more than 1 parent).
			// We diff against the first parent (parts[1]) to show changes introduced by the merge request.
			return c.runCmd("diff", parts[1]+".."+hash, "--color=never")
		}
	}
	return c.runCmd("show", "--color=never", hash)
}

// StageFile stages a file (adds it to the index).
func (c *Client) StageFile(path string) error {
	_, err := c.runCmd("add", "--", path)
	return err
}

// UnstageFile unstages a file (removes it from the index).
func (c *Client) UnstageFile(path string) error {
	// Use reset HEAD which is universally supported
	_, err := c.runCmd("reset", "HEAD", "--", path)
	return err
}

// FetchOrigin runs "git fetch" in the background to update remote tracking details.
func (c *Client) FetchOrigin() error {
	// Since this is network-bound, it's run in a separate go-routine by the caller,
	// but the client execution is simple:
	_, err := c.runCmd("fetch", "--quiet")
	return err
}

// IsInsideRepo checks if the folder is a valid git repository.
func (c *Client) IsInsideRepo() bool {
	_, err := c.runCmd("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// GetBranches returns a list of all local and remote branches.
func (c *Client) GetBranches() ([]string, error) {
	out, err := c.runCmd("branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetTags returns a list of all tags.
func (c *Client) GetTags() ([]string, error) {
	out, err := c.runCmd("tag", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var tags []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			tags = append(tags, line)
		}
	}
	return tags, nil
}

// Checkout checks out a branch, tag, or specific commit.
func (c *Client) Checkout(ref string) error {
	_, err := c.runCmd("checkout", ref)
	return err
}

// GetCompareDiff returns the diff comparing current HEAD with another ref.
func (c *Client) GetCompareDiff(ref string) (string, error) {
	return c.runCmd("diff", "HEAD.."+ref, "--color=never")
}

// GetDiffWithOrigin returns the diff between the current HEAD and its origin counterpart.
func (c *Client) GetDiffWithOrigin(branch string) (string, error) {
	// Try to diff with the upstream branch (e.g. origin/master)
	out, err := c.runCmd("diff", "HEAD..@{u}", "--color=never")
	if err == nil {
		return out, nil
	}
	// If no upstream is configured, try default origin/<branch>
	if branch != "" && branch != "HEAD (detached)" {
		return c.runCmd("diff", "HEAD..origin/"+branch, "--color=never")
	}
	return "", fmt.Errorf("no upstream tracking branch found for current branch")
}

// RevertCommit attempts to revert the specified commit and automatically creates a revert commit.
func (c *Client) RevertCommit(hash string) error {
	_, err := c.runCmd("revert", "--no-edit", hash)
	return err
}

// GetRemoteURL returns the remote URL for origin.
func (c *Client) GetRemoteURL() (string, error) {
	out, err := c.runCmd("remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
