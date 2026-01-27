package plugin

// Config holds the plugin configuration parsed from environment variables
type Config struct {
	// SCM Provider
	SCMProvider string `envconfig:"SCM_PROVIDER"` // github, gitlab, bitbucket, gitea, gogs, harness
	SCMEndpoint string `envconfig:"SCM_ENDPOINT"` // Custom endpoint for self-hosted
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

	// Batch Comments from JSON file
	CommentsFile string `envconfig:"COMMENTS_FILE"` // Path to JSON file with array of comments

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

// ReviewsFile represents the top-level structure of the reviews JSON file
type ReviewsFile struct {
	Reviews []ReviewComment `json:"reviews"`
}

// ReviewComment represents a single review comment from the JSON file
type ReviewComment struct {
	FilePath        string `json:"file_path"`
	LineNumberStart int    `json:"line_number_start"`
	LineNumberEnd   int    `json:"line_number_end"`
	Type            string `json:"type"` // issue|performance|scalability|code_smell|etc
	Review          string `json:"review"`
}
