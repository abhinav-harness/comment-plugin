package plugin

// Config holds the plugin configuration parsed from environment variables
type Config struct {
	// SCM Provider
	SCMProvider string `envconfig:"SCM_PROVIDER" required:"true"` // github, gitlab, bitbucket, gitea, gogs, harness
	SCMEndpoint string `envconfig:"SCM_ENDPOINT"`                 // Custom endpoint for self-hosted
	Token       string `envconfig:"TOKEN" required:"true"`

	// Repository
	Repo      string `envconfig:"REPO" required:"true"` // owner/repo
	PRNumber  int    `envconfig:"PR_NUMBER"`
	CommitSHA string `envconfig:"COMMIT_SHA"`

	// Comment
	CommentBody string `envconfig:"COMMENT_BODY"`
	
	// Inline Comment
	FilePath string `envconfig:"FILE_PATH"`
	Line     int    `envconfig:"LINE"`

	// Status
	StatusState   string `envconfig:"STATUS_STATE"`
	StatusContext string `envconfig:"STATUS_CONTEXT"`
	StatusDesc    string `envconfig:"STATUS_DESC"`
	StatusURL     string `envconfig:"STATUS_URL"`

	// Harness Code
	HarnessAccountID string `envconfig:"HARNESS_ACCOUNT_ID"`
	HarnessOrgID     string `envconfig:"HARNESS_ORG_ID"`
	HarnessProjectID string `envconfig:"HARNESS_PROJECT_ID"`

	// Debug
	Debug  bool `envconfig:"DEBUG"`
	DryRun bool `envconfig:"DRY_RUN"`
}
