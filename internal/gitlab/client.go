package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GitLab models
type MergeRequest struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	Author       Author `json:"author"`
	WebURL       string `json:"web_url"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
}

type Author struct {
	Name string `json:"name"`
}

type Pipeline struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Ref       string `json:"ref"`
	WebURL    string `json:"web_url"`
	CreatedAt string `json:"created_at"`
}

type Job struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Stage  string `json:"stage"`
	Status string `json:"status"`
}

type Issue struct {
	IID         int       `json:"iid"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	Description string    `json:"description"`
	WebURL      string    `json:"web_url"`
	Assignee    *Assignee `json:"assignee"`
	CreatedAt   string    `json:"created_at"`
}

type Assignee struct {
	Name string `json:"name"`
}

type Client struct {
	Host        string
	ProjectPath string
	Token       string
	HTTPClient  *http.Client
}

// ParseRemoteURL extracts host and URL-encoded project path from Git remote URL
func ParseRemoteURL(remoteURL string) (host string, projectPath string, err error) {
	remoteURL = strings.TrimSpace(remoteURL)
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	var rawPath string
	if strings.HasPrefix(remoteURL, "git@") {
		// Format: git@gitlab.com:group/subgroup/project
		parts := strings.SplitN(remoteURL[4:], ":", 2)
		if len(parts) == 2 {
			host = parts[0]
			rawPath = parts[1]
		}
	} else if strings.HasPrefix(remoteURL, "https://") {
		// Format: https://gitlab.com/group/subgroup/project
		urlWithoutProto := remoteURL[8:]
		parts := strings.SplitN(urlWithoutProto, "/", 2)
		if len(parts) == 2 {
			host = parts[0]
			rawPath = parts[1]
		}
	} else if strings.HasPrefix(remoteURL, "http://") {
		// Format: http://gitlab.com/group/subgroup/project
		urlWithoutProto := remoteURL[7:]
		parts := strings.SplitN(urlWithoutProto, "/", 2)
		if len(parts) == 2 {
			host = parts[0]
			rawPath = parts[1]
		}
	}

	if host == "" || rawPath == "" {
		return "", "", fmt.Errorf("unable to parse git remote URL: %s", remoteURL)
	}

	// URL encode the project path (e.g. group/subgroup/project -> group%2Fsubgroup%2Fproject)
	projectPath = url.PathEscape(rawPath)
	return host, projectPath, nil
}

// NewClient returns a new GitLab API client
func NewClient(remoteURL string) (*Client, error) {
	host, projectPath, err := ParseRemoteURL(remoteURL)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("GITLAB_TOKEN")

	return &Client{
		Host:        host,
		ProjectPath: projectPath,
		Token:       token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// makeRequest helper method to run GitLab API requests
func (c *Client) makeRequest(method string, path string, queryParams url.Values) ([]byte, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN environment variable is not set")
	}

	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/%s", c.Host, c.ProjectPath, path)
	if queryParams != nil && len(queryParams) > 0 {
		apiURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequest(method, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed: invalid GITLAB_TOKEN")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetMergeRequests fetches open Merge Requests for the project
func (c *Client) GetMergeRequests() ([]MergeRequest, error) {
	q := url.Values{}
	q.Set("state", "opened")
	q.Set("per_page", "20")

	data, err := c.makeRequest("GET", "merge_requests", q)
	if err != nil {
		return nil, err
	}

	var mrs []MergeRequest
	err = json.Unmarshal(data, &mrs)
	return mrs, err
}

// GetPipelines fetches the latest pipelines for the project
func (c *Client) GetPipelines() ([]Pipeline, error) {
	q := url.Values{}
	q.Set("per_page", "20")

	data, err := c.makeRequest("GET", "pipelines", q)
	if err != nil {
		return nil, err
	}

	var pipelines []Pipeline
	err = json.Unmarshal(data, &pipelines)
	return pipelines, err
}

// GetPipelineJobs fetches jobs for a given pipeline
func (c *Client) GetPipelineJobs(pipelineID int) ([]Job, error) {
	path := fmt.Sprintf("pipelines/%d/jobs", pipelineID)
	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var jobs []Job
	err = json.Unmarshal(data, &jobs)
	return jobs, err
}

// GetJobLogs fetches trace/logs for a specific job
func (c *Client) GetJobLogs(jobID int) (string, error) {
	path := fmt.Sprintf("jobs/%d/trace", jobID)
	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetIssues fetches open issues for the project
func (c *Client) GetIssues() ([]Issue, error) {
	q := url.Values{}
	q.Set("state", "opened")
	q.Set("per_page", "20")

	data, err := c.makeRequest("GET", "issues", q)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	err = json.Unmarshal(data, &issues)
	return issues, err
}
