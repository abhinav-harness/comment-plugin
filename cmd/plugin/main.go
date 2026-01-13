package main

import (
	"context"
	"os"

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

	// Set log level
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithFields(logrus.Fields{
		"scm_provider": cfg.SCMProvider,
		"repo":         cfg.Repo,
		"pr_number":    cfg.PRNumber,
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
