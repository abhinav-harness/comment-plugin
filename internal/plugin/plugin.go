package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/abhinav-harness/comment-plugin/internal/harness"
	scmclient "github.com/abhinav-harness/comment-plugin/internal/scm"
	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
)

// Plugin represents the comment plugin
type Plugin struct {
	config  Config
	client  *scm.Client
	harness *harness.Client
	log     *logrus.Entry
}

// New creates a new Plugin instance
func New(cfg Config) (*Plugin, error) {
	log := logrus.WithField("component", "plugin")

	provider := scmclient.Provider(strings.ToLower(cfg.SCMProvider))

	client, err := scmclient.NewClient(scmclient.ClientOptions{
		Provider: provider,
		Endpoint: cfg.SCMEndpoint,
		Token:    cfg.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SCM client: %w", err)
	}

	p := &Plugin{
		config: cfg,
		client: client,
		log:    log,
	}

	// Initialize Harness wrapper if using Harness Code
	if provider == scmclient.ProviderHarness {
		harnessClient, err := harness.NewClient(harness.Config{
			Endpoint:  cfg.SCMEndpoint,
			Token:     cfg.Token,
			AccountID: cfg.HarnessAccountID,
			OrgID:     cfg.HarnessOrgID,
			ProjectID: cfg.HarnessProjectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Harness client: %w", err)
		}
		p.harness = harnessClient
	}

	return p, nil
}

// Execute runs the plugin
func (p *Plugin) Execute(ctx context.Context) error {
	// Log all input configuration for debugging
	p.log.WithFields(logrus.Fields{
		"scm_provider":       p.config.SCMProvider,
		"scm_endpoint":       p.config.SCMEndpoint,
		"repo":               p.config.Repo,
		"pr_number":          p.config.PRNumber,
		"commit_sha":         p.config.CommitSHA,
		"harness_account_id": p.config.HarnessAccountID,
		"harness_org_id":     p.config.HarnessOrgID,
		"harness_project_id": p.config.HarnessProjectID,
		"comments_file":      p.config.CommentsFile,
		"file_path":          p.config.FilePath,
		"line":               p.config.Line,
		"debug":              p.config.Debug,
		"dry_run":            p.config.DryRun,
	}).Info("executing comment plugin with configuration")

	if p.config.DryRun {
		p.log.WithField("body", p.config.CommentBody).Info("dry run - would post comment")
		return nil
	}

	// Determine what action to take
	if p.config.CommentsFile != "" {
		return p.createCommentsFromFile(ctx)
	}

	if p.config.StatusState != "" {
		return p.createStatus(ctx)
	}

	if p.config.FilePath != "" && p.config.Line > 0 {
		return p.createInlineComment(ctx)
	}

	if p.config.CommentBody != "" {
		return p.createComment(ctx)
	}

	return fmt.Errorf("no action: provide COMMENT_BODY, FILE_PATH+LINE, COMMENTS_FILE, or STATUS_STATE")
}

func (p *Plugin) createCommentsFromFile(ctx context.Context) error {
	if p.config.PRNumber == 0 {
		return fmt.Errorf("PR_NUMBER is required")
	}

	// Check if file exists
	if _, err := os.Stat(p.config.CommentsFile); os.IsNotExist(err) {
		p.log.WithField("file", p.config.CommentsFile).Warn("comments file not found, skipping")
		return nil
	}

	// Read the JSON file
	data, err := os.ReadFile(p.config.CommentsFile)
	if err != nil {
		return fmt.Errorf("failed to read comments file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		p.log.WithField("file", p.config.CommentsFile).Warn("comments file is empty, skipping")
		return nil
	}

	// Parse JSON
	var reviewsFile ReviewsFile
	if err := json.Unmarshal(data, &reviewsFile); err != nil {
		return fmt.Errorf("failed to parse comments file: %w", err)
	}

	reviews := reviewsFile.Reviews

	// Handle empty reviews array
	if len(reviews) == 0 {
		p.log.WithField("file", p.config.CommentsFile).Info("no reviews in file, nothing to post")
		return nil
	}

	p.log.WithFields(logrus.Fields{
		"file":  p.config.CommentsFile,
		"count": len(reviews),
	}).Info("loaded reviews from file")

	// Harness Code
	if p.harness != nil {
		// Get PR details to fetch commit SHAs
		pr, err := p.getPRDetails(ctx)
		if err != nil {
			return fmt.Errorf("failed to get PR details: %w", err)
		}

		// Create each review comment
		for i, review := range reviews {
			err := p.harness.CreateReviewComment(
				ctx,
				p.config.Repo,
				p.config.PRNumber,
				review.FilePath,
				review.LineNumberStart,
				review.LineNumberEnd,
				review.Type,
				review.Review,
				pr.SourceSHA,
				pr.TargetSHA,
			)
			if err != nil {
				p.log.WithError(err).WithField("index", i).Warn("failed to create review comment")
				// Continue with remaining comments
			}
		}

		p.log.WithField("count", len(reviews)).Info("finished creating review comments")
		return nil
	}

	// For other SCM providers, use go-scm Reviews API
	for i, review := range reviews {
		commentText := review.Review
		if review.Type != "" {
			commentText = fmt.Sprintf("**%s:** %s", review.Type, review.Review)
		}

		input := &scm.ReviewInput{
			Body: commentText,
			Path: review.FilePath,
			Line: review.LineNumberEnd,
			Sha:  p.config.CommitSHA,
		}

		_, _, err := p.client.Reviews.Create(ctx, p.config.Repo, p.config.PRNumber, input)
		if err != nil {
			p.log.WithError(err).WithField("index", i).WithField("path", review.FilePath).Warn("failed to create review comment")
		}
	}

	p.log.WithField("count", len(reviews)).Info("finished creating review comments from file")
	return nil
}

func (p *Plugin) getPRDetails(ctx context.Context) (*harness.PRDetails, error) {
	// For Harness, we need to use the Harness client to get PR details
	if p.harness != nil {
		return p.harness.GetPRDetails(ctx, p.config.Repo, p.config.PRNumber)
	}

	// For other providers, we can use go-scm
	pr, _, err := p.client.PullRequests.Find(ctx, p.config.Repo, p.config.PRNumber)
	if err != nil {
		return nil, err
	}

	return &harness.PRDetails{
		SourceSHA: pr.Sha,
		TargetSHA: pr.Base.Sha,
	}, nil
}

func (p *Plugin) createComment(ctx context.Context) error {
	if p.config.PRNumber == 0 {
		return fmt.Errorf("PR_NUMBER is required")
	}

	// Harness Code
	if p.harness != nil {
		return p.harness.CreateComment(ctx, p.config.Repo, p.config.PRNumber, p.config.CommentBody)
	}

	// go-scm
	input := &scm.CommentInput{Body: p.config.CommentBody}
	comment, _, err := p.client.PullRequests.CreateComment(ctx, p.config.Repo, p.config.PRNumber, input)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	p.log.WithField("comment_id", comment.ID).Info("created comment")
	return nil
}

func (p *Plugin) createInlineComment(ctx context.Context) error {
	if p.config.PRNumber == 0 {
		return fmt.Errorf("PR_NUMBER is required")
	}
	if p.config.CommentBody == "" {
		return fmt.Errorf("COMMENT_BODY is required")
	}

	// Harness Code
	if p.harness != nil {
		return p.harness.CreateInlineComment(ctx, p.config.Repo, p.config.PRNumber, p.config.FilePath, p.config.Line, p.config.CommentBody)
	}

	// go-scm uses Reviews for inline comments
	input := &scm.ReviewInput{
		Body: p.config.CommentBody,
		Path: p.config.FilePath,
		Line: p.config.Line,
		Sha:  p.config.CommitSHA,
	}

	_, _, err := p.client.Reviews.Create(ctx, p.config.Repo, p.config.PRNumber, input)
	if err != nil {
		return fmt.Errorf("failed to create inline comment: %w", err)
	}

	p.log.WithFields(logrus.Fields{"file": p.config.FilePath, "line": p.config.Line}).Info("created inline comment")
	return nil
}

func (p *Plugin) createStatus(ctx context.Context) error {
	if p.config.CommitSHA == "" {
		return fmt.Errorf("COMMIT_SHA is required")
	}

	// Harness Code
	if p.harness != nil {
		return p.harness.CreateStatus(ctx, p.config.Repo, p.config.CommitSHA, p.config.StatusState, p.config.StatusContext, p.config.StatusDesc, p.config.StatusURL)
	}

	// go-scm
	input := &scm.StatusInput{
		State:  mapStatusState(p.config.StatusState),
		Label:  p.config.StatusContext,
		Desc:   p.config.StatusDesc,
		Target: p.config.StatusURL,
	}

	_, _, err := p.client.Repositories.CreateStatus(ctx, p.config.Repo, p.config.CommitSHA, input)
	if err != nil {
		return fmt.Errorf("failed to create status: %w", err)
	}

	p.log.WithFields(logrus.Fields{"state": p.config.StatusState, "context": p.config.StatusContext}).Info("created status")
	return nil
}

func mapStatusState(state string) scm.State {
	switch strings.ToLower(state) {
	case "success":
		return scm.StateSuccess
	case "failure", "failed":
		return scm.StateFailure
	case "error":
		return scm.StateError
	case "pending":
		return scm.StatePending
	case "running":
		return scm.StateRunning
	default:
		return scm.StateUnknown
	}
}
