package scm

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/bitbucket"
	"github.com/drone/go-scm/scm/driver/gitea"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/driver/gogs"
	"github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport"
	"github.com/drone/go-scm/scm/transport/oauth2"
)

// Provider represents supported SCM providers
type Provider string

const (
	ProviderGitHub          Provider = "github"
	ProviderGitHubEnterprise Provider = "github-enterprise"
	ProviderGitLab          Provider = "gitlab"
	ProviderBitbucket       Provider = "bitbucket"
	ProviderBitbucketServer Provider = "bitbucket-server"
	ProviderGitea           Provider = "gitea"
	ProviderGogs            Provider = "gogs"
	ProviderHarness         Provider = "harness"
	ProviderAzureDevOps     Provider = "azure-devops"
)

// ClientOptions holds options for creating an SCM client
type ClientOptions struct {
	Provider Provider
	Endpoint string // Custom endpoint for self-hosted instances
	Token    string // Authentication token

	// Harness-specific options
	HarnessAccountID string
	HarnessOrgID     string
	HarnessProjectID string
}

// NewClient creates a new SCM client based on the provider
func NewClient(opts ClientOptions) (*scm.Client, error) {
	var client *scm.Client
	var err error

	switch opts.Provider {
	case ProviderGitHub:
		client = github.NewDefault()
	case ProviderGitHubEnterprise:
		if opts.Endpoint == "" {
			return nil, fmt.Errorf("endpoint required for GitHub Enterprise")
		}
		client, err = github.New(opts.Endpoint)
	case ProviderGitLab:
		if opts.Endpoint != "" {
			client, err = gitlab.New(opts.Endpoint)
		} else {
			client = gitlab.NewDefault()
		}
	case ProviderBitbucket:
		client = bitbucket.NewDefault()
	case ProviderBitbucketServer:
		if opts.Endpoint == "" {
			return nil, fmt.Errorf("endpoint required for Bitbucket Server")
		}
		client, err = stash.New(opts.Endpoint)
	case ProviderGitea:
		if opts.Endpoint == "" {
			return nil, fmt.Errorf("endpoint required for Gitea")
		}
		client, err = gitea.New(opts.Endpoint)
	case ProviderGogs:
		if opts.Endpoint == "" {
			return nil, fmt.Errorf("endpoint required for Gogs")
		}
		client, err = gogs.New(opts.Endpoint)
	case ProviderHarness:
		// Harness Code uses a custom client - we'll wrap it
		client, err = newHarnessClient(opts)
	case ProviderAzureDevOps:
		return nil, fmt.Errorf("Azure DevOps is not yet supported by go-scm, use harness wrapper")
	default:
		return nil, fmt.Errorf("unsupported SCM provider: %s", opts.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create SCM client: %w", err)
	}

	// Configure authentication
	if opts.Token != "" {
		client.Client = configureAuth(client.Client, opts)
	}

	return client, nil
}

// configureAuth configures authentication for the HTTP client
func configureAuth(httpClient *http.Client, opts ClientOptions) *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	switch opts.Provider {
	case ProviderBitbucket:
		// Bitbucket Cloud uses OAuth2
		httpClient.Transport = &oauth2.Transport{
			Source: oauth2.StaticTokenSource(&scm.Token{Token: opts.Token}),
		}
	case ProviderBitbucketServer:
		// Bitbucket Server uses Bearer token
		httpClient.Transport = &transport.BearerToken{Token: opts.Token}
	default:
		// Most providers use Bearer token
		httpClient.Transport = &transport.BearerToken{Token: opts.Token}
	}

	return httpClient
}

// newHarnessClient creates a client configured for Harness Code
func newHarnessClient(opts ClientOptions) (*scm.Client, error) {
	endpoint := opts.Endpoint
	if endpoint == "" {
		endpoint = "https://app.harness.io/gateway/code/api/v1"
	}

	// Harness Code API is compatible with a subset of GitHub API
	// but requires special headers and path modifications
	client, err := github.New(endpoint)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// ParseRepo parses a repository string into owner and repo name
func ParseRepo(repo string) (owner, name string, err error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}
	return parts[0], parts[1], nil
}

// SupportedProviders returns a list of supported SCM providers
func SupportedProviders() []Provider {
	return []Provider{
		ProviderGitHub,
		ProviderGitHubEnterprise,
		ProviderGitLab,
		ProviderBitbucket,
		ProviderBitbucketServer,
		ProviderGitea,
		ProviderGogs,
		ProviderHarness,
	}
}

