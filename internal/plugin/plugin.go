package plugin

import (
	"context"
	"fmt"
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
	p.log.Info("executing comment plugin")

	if p.config.DryRun {
		p.log.WithField("body", p.config.CommentBody).Info("dry run - would post comment")
		return nil
	}

	// Determine what action to take
	if p.config.StatusState != "" {
		return p.createStatus(ctx)
	}

	if p.config.FilePath != "" && p.config.Line > 0 {
		return p.createInlineComment(ctx)
	}

	if p.config.CommentBody != "" {
		return p.createComment(ctx)
	}

	return fmt.Errorf("no action: provide COMMENT_BODY, FILE_PATH+LINE, or STATUS_STATE")
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
