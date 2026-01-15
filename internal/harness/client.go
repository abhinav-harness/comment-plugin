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

	log := logrus.WithField("component", "harness")

	// Log configuration (mask token for security)
	tokenPreview := ""
	if len(cfg.Token) > 8 {
		tokenPreview = cfg.Token[:8] + "..."
	}

	log.WithFields(logrus.Fields{
		"base_url":   baseURL,
		"account_id": cfg.AccountID,
		"org_id":     cfg.OrgID,
		"project_id": cfg.ProjectID,
		"token":      tokenPreview,
	}).Info("initialized Harness Code client")

	return &Client{
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		log:        log,
	}, nil
}

// CreateComment creates a comment on a pull request
func (c *Client) CreateComment(ctx context.Context, repo string, prNumber int, body string) error {
	c.log.WithFields(logrus.Fields{
		"repo":      repo,
		"pr_number": prNumber,
	}).Info("creating PR comment")

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

	c.log.Info("created PR comment successfully")
	return nil
}

// CreateInlineComment creates an inline comment on a specific file/line
func (c *Client) CreateInlineComment(ctx context.Context, repo string, prNumber int, filePath string, line int, body string) error {
	c.log.WithFields(logrus.Fields{
		"repo":      repo,
		"pr_number": prNumber,
		"file_path": filePath,
		"line":      line,
	}).Info("creating inline comment")

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
	}).Info("created inline comment successfully")
	return nil
}

// prInfo holds minimal PR info needed for code comments
type prInfo struct {
	SourceSHA    string `json:"source_sha"`
	MergeBaseSHA string `json:"merge_base_sha"`
}

// PRDetails holds PR details for external use
type PRDetails struct {
	SourceSHA string
	TargetSHA string
}

func (c *Client) getPR(ctx context.Context, repo string, prNumber int) (*prInfo, error) {
	c.log.WithFields(logrus.Fields{
		"repo":      repo,
		"pr_number": prNumber,
	}).Debug("fetching PR details")

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

	c.log.WithFields(logrus.Fields{
		"source_sha":     pr.SourceSHA,
		"merge_base_sha": pr.MergeBaseSHA,
	}).Debug("fetched PR details")

	return &pr, nil
}

// GetPRDetails retrieves PR details including commit SHAs
func (c *Client) GetPRDetails(ctx context.Context, repo string, prNumber int) (*PRDetails, error) {
	c.log.WithFields(logrus.Fields{
		"repo":      repo,
		"pr_number": prNumber,
	}).Info("getting PR details for batch comments")

	pr, err := c.getPR(ctx, repo, prNumber)
	if err != nil {
		return nil, err
	}

	c.log.WithFields(logrus.Fields{
		"source_sha": pr.SourceSHA,
		"target_sha": pr.MergeBaseSHA,
	}).Info("retrieved PR commit SHAs")

	return &PRDetails{
		SourceSHA: pr.SourceSHA,
		TargetSHA: pr.MergeBaseSHA,
	}, nil
}

// CreateReviewComment creates a review comment on a specific file/line
func (c *Client) CreateReviewComment(ctx context.Context, repo string, prNumber int, filePath string, lineStart, lineEnd int, reviewType, reviewText, sourceSHA, targetSHA string) error {
	c.log.WithFields(logrus.Fields{
		"repo":       repo,
		"pr_number":  prNumber,
		"file_path":  filePath,
		"line_start": lineStart,
		"line_end":   lineEnd,
		"type":       reviewType,
	}).Debug("creating review comment")

	path := c.apiPath(repo, fmt.Sprintf("pullreq/%d/comments", prNumber))

	// Format the comment text with type prefix
	commentText := reviewText
	if reviewType != "" {
		commentText = fmt.Sprintf("**%s:** %s", reviewType, reviewText)
	}

	payload := map[string]interface{}{
		"text":              commentText,
		"path":              filePath,
		"line_start":        lineStart,
		"line_end":          lineEnd,
		"source_commit_sha": sourceSHA,
		"target_commit_sha": targetSHA,
	}

	resp, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return fmt.Errorf("failed to create review comment: %w", err)
	}

	c.log.WithFields(logrus.Fields{
		"file":       filePath,
		"line_start": lineStart,
		"line_end":   lineEnd,
		"type":       reviewType,
	}).Info("created review comment successfully")
	return nil
}

// CreateStatus creates a commit status check
func (c *Client) CreateStatus(ctx context.Context, repo, commitSHA, state, statusContext, description, targetURL string) error {
	c.log.WithFields(logrus.Fields{
		"repo":       repo,
		"commit_sha": commitSHA,
		"state":      state,
		"context":    statusContext,
	}).Info("creating commit status")

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
	}).Info("created commit status successfully")
	return nil
}

func (c *Client) apiPath(repo, suffix string) string {
	repoPath := c.buildRepoPath(repo)
	// Don't URL-encode the repo path - it contains slashes that need to remain as-is
	path := fmt.Sprintf("%s/gateway/code/api/v1/repos/%s/%s", c.baseURL, repoPath, suffix)

	// Add routingId query parameter
	if c.config.AccountID != "" {
		path = fmt.Sprintf("%s?routingId=%s", path, url.QueryEscape(c.config.AccountID))
	}

	c.log.WithFields(logrus.Fields{
		"repo":      repo,
		"repo_path": repoPath,
		"api_url":   path,
	}).Info("API request URL")

	return path
}

func (c *Client) buildRepoPath(repo string) string {
	var repoPath string

	// If org/project are provided, build the full path
	if c.config.OrgID != "" && c.config.ProjectID != "" {
		parts := strings.Split(repo, "/")
		switch len(parts) {
		case 1:
			repoPath = fmt.Sprintf("%s/%s/+/repos/%s", c.config.OrgID, c.config.ProjectID, repo)
		case 2:
			repoPath = fmt.Sprintf("%s/%s/+/repos/%s", c.config.OrgID, parts[0], parts[1])
		default:
			repoPath = fmt.Sprintf("%s/%s/+/repos/%s", parts[0], parts[1], strings.Join(parts[2:], "/"))
		}
	} else {
		// For account-level repos, just use the repo name directly
		repoPath = repo
	}

	c.log.WithFields(logrus.Fields{
		"input_repo": repo,
		"org_id":     c.config.OrgID,
		"project_id": c.config.ProjectID,
		"repo_path":  repoPath,
	}).Debug("built repo path")

	return repoPath
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
		c.log.WithField("payload", string(jsonBody)).Debug("request payload")
	}

	c.log.WithFields(logrus.Fields{
		"method": method,
		"url":    path,
	}).Debug("making API request")

	req, err := http.NewRequestWithContext(ctx, method, path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.Token)
	if c.config.AccountID != "" {
		req.Header.Set("Harness-Account", c.config.AccountID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.WithError(err).Error("API request failed")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	}).Debug("received API response")

	return resp, nil
}

func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)

	c.log.WithFields(logrus.Fields{
		"status_code":   resp.StatusCode,
		"response_body": string(body),
	}).Error("API request returned error")

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
