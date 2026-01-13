package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Config holds configuration for the Harness Code client
type Config struct {
	Endpoint  string
	Token     string
	AccountID string
	OrgID     string
	ProjectID string
}

// Client is a specialized client for Harness Code
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
	log        *logrus.Entry
}

// NewClient creates a new Harness Code client
func NewClient(cfg Config) (*Client, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	baseURL := cfg.Endpoint
	if baseURL == "" {
		baseURL = "https://app.harness.io"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		log:        logrus.WithField("component", "harness"),
	}, nil
}

// CreateComment creates a comment on a pull request
func (c *Client) CreateComment(ctx context.Context, repo string, prNumber int, body string) error {
	path := c.apiPath(repo, fmt.Sprintf("pullreq/%d/comments", prNumber))

	payload := map[string]interface{}{
		"text": body,
	}

	resp, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	c.log.Info("created PR comment")
	return nil
}

// CreateInlineComment creates an inline comment on a specific file/line
func (c *Client) CreateInlineComment(ctx context.Context, repo string, prNumber int, filePath string, line int, body string) error {
	// First, get PR details to obtain commit SHAs (required for code comments)
	pr, err := c.getPR(ctx, repo, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR details: %w", err)
	}

	path := c.apiPath(repo, fmt.Sprintf("pullreq/%d/comments", prNumber))

	payload := map[string]interface{}{
		"text":              body,
		"path":              filePath,
		"line_start":        line,
		"line_end":          line,
		"source_commit_sha": pr.SourceSHA,
		"target_commit_sha": pr.MergeBaseSHA,
	}

	resp, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return fmt.Errorf("failed to create inline comment: %w", err)
	}

	c.log.WithFields(logrus.Fields{
		"file": filePath,
		"line": line,
	}).Info("created inline comment")
	return nil
}

// prInfo holds minimal PR info needed for code comments
type prInfo struct {
	SourceSHA    string `json:"source_sha"`
	MergeBaseSHA string `json:"merge_base_sha"`
}

func (c *Client) getPR(ctx context.Context, repo string, prNumber int) (*prInfo, error) {
	path := c.apiPath(repo, fmt.Sprintf("pullreq/%d", prNumber))

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var pr prInfo
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// CodeComment represents a code comment with all fields
type CodeComment struct {
	Text            string `json:"text"`
	LineStart       int    `json:"line_start"`
	LineEnd         int    `json:"line_end"`
	LineStartNew    bool   `json:"line_start_new"`
	LineEndNew      bool   `json:"line_end_new"`
	Path            string `json:"path"`
	SourceCommitSHA string `json:"source_commit_sha"`
	TargetCommitSHA string `json:"target_commit_sha"`
	ParentID        int    `json:"parent_id,omitempty"`
}

// CreateCodeComment creates a code comment with full control over all fields
func (c *Client) CreateCodeComment(ctx context.Context, repo string, prNumber int, comment CodeComment) error {
	path := c.apiPath(repo, fmt.Sprintf("pullreq/%d/comments", prNumber))

	payload := map[string]interface{}{
		"text":              comment.Text,
		"path":              comment.Path,
		"line_start":        comment.LineStart,
		"line_end":          comment.LineEnd,
		"source_commit_sha": comment.SourceCommitSHA,
		"target_commit_sha": comment.TargetCommitSHA,
	}

	if comment.ParentID > 0 {
		payload["parent_id"] = comment.ParentID
	}

	resp, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return fmt.Errorf("failed to create code comment: %w", err)
	}

	c.log.WithFields(logrus.Fields{
		"file":       comment.Path,
		"line_start": comment.LineStart,
		"line_end":   comment.LineEnd,
	}).Info("created code comment")
	return nil
}

// CreateCodeComments creates multiple code comments from a slice
func (c *Client) CreateCodeComments(ctx context.Context, repo string, prNumber int, comments []CodeComment) error {
	for i, comment := range comments {
		if err := c.CreateCodeComment(ctx, repo, prNumber, comment); err != nil {
			c.log.WithError(err).WithField("index", i).Warn("failed to create comment")
			// Continue with remaining comments
		}
	}
	c.log.WithField("count", len(comments)).Info("finished creating code comments")
	return nil
}

// CreateStatus creates a commit status check
func (c *Client) CreateStatus(ctx context.Context, repo, commitSHA, state, statusContext, description, targetURL string) error {
	path := c.apiPath(repo, fmt.Sprintf("commits/%s/checks", commitSHA))

	payload := map[string]interface{}{
		"identifier": statusContext,
		"status":     mapState(state),
		"summary":    description,
	}
	if targetURL != "" {
		payload["link"] = targetURL
	}

	resp, err := c.do(ctx, http.MethodPut, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return fmt.Errorf("failed to create status: %w", err)
	}

	c.log.WithFields(logrus.Fields{
		"status":  state,
		"context": statusContext,
	}).Info("created commit status")
	return nil
}

func (c *Client) apiPath(repo, suffix string) string {
	repoPath := c.buildRepoPath(repo)
	path := fmt.Sprintf("%s/gateway/code/api/v1/repos/%s/%s", c.baseURL, url.PathEscape(repoPath), suffix)

	// Add routingId query parameter
	if c.config.AccountID != "" {
		path = fmt.Sprintf("%s?routingId=%s", path, url.QueryEscape(c.config.AccountID))
	}
	return path
}

func (c *Client) buildRepoPath(repo string) string {
	// If org/project are provided, build the full path
	if c.config.OrgID != "" && c.config.ProjectID != "" {
		parts := strings.Split(repo, "/")
		switch len(parts) {
		case 1:
			return fmt.Sprintf("%s/%s/+/repos/%s", c.config.OrgID, c.config.ProjectID, repo)
		case 2:
			return fmt.Sprintf("%s/%s/+/repos/%s", c.config.OrgID, parts[0], parts[1])
		default:
			return fmt.Sprintf("%s/%s/+/repos/%s", parts[0], parts[1], strings.Join(parts[2:], "/"))
		}
	}
	// For account-level repos, just use the repo name directly
	return repo
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.Token)
	if c.config.AccountID != "" {
		req.Header.Set("Harness-Account", c.config.AccountID)
	}

	return c.httpClient.Do(req)
}

func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
}

func mapState(state string) string {
	switch strings.ToLower(state) {
	case "success":
		return "success"
	case "failure", "failed":
		return "failure"
	case "error":
		return "error"
	case "pending":
		return "pending"
	case "running":
		return "running"
	default:
		return "pending"
	}
}
