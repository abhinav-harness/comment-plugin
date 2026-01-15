package main

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/abhinav-harness/comment-plugin/internal/plugin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load .env file if exists (for local development)
	_ = godotenv.Load()

	// Configure logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Parse configuration from environment variables
	var cfg plugin.Config
	if err := envconfig.Process("PLUGIN", &cfg); err != nil {
		logrus.WithError(err).Fatal("failed to parse configuration")
	}

	// Also check for HARNESS_* env vars directly (without PLUGIN_ prefix)
	if cfg.HarnessAccountID == "" {
		cfg.HarnessAccountID = os.Getenv("HARNESS_ACCOUNT_ID")
	}
	if cfg.HarnessOrgID == "" {
		cfg.HarnessOrgID = os.Getenv("HARNESS_ORG_ID")
	}
	if cfg.HarnessProjectID == "" {
		cfg.HarnessProjectID = os.Getenv("HARNESS_PROJECT_ID")
	}

	// Extract base URL from HARNESS_STO_SERVICE_ENDPOINT if SCM_ENDPOINT is not set
	// e.g., "https://qa.harness.io/prod1/sto/" -> "https://qa.harness.io"
	if cfg.SCMEndpoint == "" {
		if stoEndpoint := os.Getenv("HARNESS_STO_SERVICE_ENDPOINT"); stoEndpoint != "" {
			if parsed, err := url.Parse(stoEndpoint); err == nil {
				cfg.SCMEndpoint = strings.TrimSuffix(parsed.Scheme+"://"+parsed.Host, "/")
				logrus.WithFields(logrus.Fields{
					"sto_endpoint": stoEndpoint,
					"base_url":     cfg.SCMEndpoint,
				}).Debug("extracted base URL from HARNESS_STO_SERVICE_ENDPOINT")
			}
		}
	}

	// Set log level
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug logging enabled")
	}

	// Mask token for logging
	tokenPreview := ""
	if len(cfg.Token) > 8 {
		tokenPreview = cfg.Token[:8] + "..."
	} else if len(cfg.Token) > 0 {
		tokenPreview = "***"
	}

	logrus.WithFields(logrus.Fields{
		"scm_provider":       cfg.SCMProvider,
		"scm_endpoint":       cfg.SCMEndpoint,
		"repo":               cfg.Repo,
		"pr_number":          cfg.PRNumber,
		"commit_sha":         cfg.CommitSHA,
		"harness_account_id": cfg.HarnessAccountID,
		"harness_org_id":     cfg.HarnessOrgID,
		"harness_project_id": cfg.HarnessProjectID,
		"comments_file":      cfg.CommentsFile,
		"file_path":          cfg.FilePath,
		"line":               cfg.Line,
		"token":              tokenPreview,
	}).Info("starting comment plugin")

	// Create and execute the plugin
	p, err := plugin.New(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create plugin")
	}

	ctx := context.Background()
	if err := p.Execute(ctx); err != nil {
		logrus.WithError(err).Fatal("plugin execution failed")
		os.Exit(1)
	}

	logrus.Info("plugin executed successfully")
}
