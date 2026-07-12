package git

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	// Test case 1: normal branch with ahead/behind
	input1 := "## master...origin/master [ahead 1, behind 2]\n M internal/git/client.go\n?? untracked.go"
	branch, upstream, ahead, behind, changes := ParseStatus(input1)
	if branch != "master" {
		t.Errorf("Expected branch master, got %s", branch)
	}
	if upstream != "origin/master" {
		t.Errorf("Expected upstream origin/master, got %s", upstream)
	}
	if ahead != 1 {
		t.Errorf("Expected ahead 1, got %d", ahead)
	}
	if behind != 2 {
		t.Errorf("Expected behind 2, got %d", behind)
	}
	if len(changes) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(changes))
	}

	// Test case 2: detached HEAD
	input2 := "## HEAD (no branch)\n"
	branch2, _, _, _, _ := ParseStatus(input2)
	if branch2 != "HEAD (detached)" {
		t.Errorf("Expected HEAD (detached), got %s", branch2)
	}
}

func TestParseCommits(t *testing.T) {
	// Format: graph prefix + hash, then fields separated by null bytes
	input := "* 125168e\x00Hadi Tayanloo\x002 hours ago\x002026-07-09\x00merge branch 'test-branch'\x00HEAD -> master, origin/master\n" +
		"| * ea95c20\x00Hadi Tayanloo\x003 hours ago\x002026-07-09\x00feat: some changes on branch\x00test-branch\n" +
		"|/  \n"
	
	commits := ParseCommits(input)
	if len(commits) != 3 {
		t.Fatalf("Expected 3 commits, got %d", len(commits))
	}

	// Commit 1
	c1 := commits[0]
	if c1.Hash != "125168e" {
		t.Errorf("Expected Hash 125168e, got %s", c1.Hash)
	}
	if c1.Author != "Hadi Tayanloo" {
		t.Errorf("Expected Author Hadi Tayanloo, got %s", c1.Author)
	}
	if c1.Date != "2 hours ago" {
		t.Errorf("Expected Date 2 hours ago, got %s", c1.Date)
	}
	if c1.Subject != "merge branch 'test-branch'" {
		t.Errorf("Expected Subject merge branch 'test-branch', got %s", c1.Subject)
	}
	if c1.Refs != "HEAD -> master, origin/master" {
		t.Errorf("Expected Refs HEAD -> master, origin/master, got %s", c1.Refs)
	}
	if c1.GraphChars != "* " {
		t.Errorf("Expected GraphChars '* ', got '%s'", c1.GraphChars)
	}

	// Commit 3 (pure graph line)
	c3 := commits[2]
	if c3.Hash != "" {
		t.Errorf("Expected empty hash for pure graph line, got %s", c3.Hash)
	}
	if c3.GraphChars != "|/  " {
		t.Errorf("Expected GraphChars '|/  ', got '%s'", c3.GraphChars)
	}
}
