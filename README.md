# 💬 Comment Plugin

A Drone/Harness CI plugin to post comments on pull requests across multiple SCM providers using [go-scm](https://github.com/drone/go-scm).

[![Docker Hub](https://img.shields.io/docker/v/abhinavharness/comment-plugin?label=Docker%20Hub)](https://hub.docker.com/r/abhinavharness/comment-plugin)
[![GitHub](https://img.shields.io/github/license/abhinav-harness/comment-plugin)](https://github.com/abhinav-harness/comment-plugin)

## Features

| Feature | GitHub | GitLab | Bitbucket | Gitea | Gogs | Harness Code |
|---------|--------|--------|-----------|-------|------|--------------|
| 💬 PR Comments | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 📝 Inline Code Comments | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 📊 Commit Status | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 📁 Batch Comments from JSON | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

## Quick Start

### Basic Example - PR Comment

```yaml
kind: pipeline
type: docker
name: comment

steps:
  - name: post-comment
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: github
      token:
        from_secret: github_token
      repo: owner/repo
      pr_number: ${DRONE_PULL_REQUEST}
      comment_body: "Build passed! ✅"
```

### Advanced Example - Batch Code Review Comments

```yaml
kind: pipeline
type: docker
name: code-review

steps:
  - name: ai-review
    image: abhinavharness/drone-ai-review:latest
    settings:
      enable_bugs: true
      enable_performance: true
      review_output_file: /drone/src/reviews.json

  - name: post-reviews
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: harness
      token:
        from_secret: harness_token
      harness_account_id: ${HARNESS_ACCOUNT_ID}
      repo: my-repo
      pr_number: ${PR_NUMBER}
      comments_file: /drone/src/reviews.json

trigger:
  event:
    - pull_request
```

## Configuration Parameters

All parameters can be configured through the `settings` block in your `.drone.yml` file or as environment variables.

### Core Settings

| Parameter | Environment Variable | Type | Required | Description |
|-----------|---------------------|------|----------|-------------|
| `scm_provider` | `SCM_PROVIDER` | string | ✅ | SCM provider: `github`, `gitlab`, `bitbucket`, `gitea`, `gogs`, `harness` |
| `token` | `TOKEN` | string | ✅ | Authentication token |
| `repo` | `REPO` | string | ✅ | Repository (`owner/repo` or repo name for Harness) |
| `scm_endpoint` | `SCM_ENDPOINT` | string | | Custom API endpoint for self-hosted instances |

### Comment Settings

| Parameter | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `pr_number` | `PR_NUMBER` | integer | | Pull request number |
| `comment_body` | `COMMENT_BODY` | string | | Comment text |
| `file_path` | `FILE_PATH` | string | | File path for inline comments |
| `line` | `LINE` | integer | | Line number for inline comments |
| `comments_file` | `COMMENTS_FILE` | string | | Path to JSON file with batch comments |

### Status Settings

| Parameter | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `commit_sha` | `COMMIT_SHA` | string | | Commit SHA for status |
| `status_state` | `STATUS_STATE` | string | | Status: `pending`, `success`, `failure`, `error` |
| `status_context` | `STATUS_CONTEXT` | string | | Status check name |
| `status_desc` | `STATUS_DESC` | string | | Status description |
| `status_url` | `STATUS_URL` | string | | Link URL for status |

### Harness Code Settings

| Parameter | Environment Variable | Type | Description |
|-----------|---------------------|------|-------------|
| `harness_account_id` | `HARNESS_ACCOUNT_ID` | string | Harness account ID (required for Harness) |
| `harness_org_id` | `HARNESS_ORG_ID` | string | Harness organization ID (optional) |
| `harness_project_id` | `HARNESS_PROJECT_ID` | string | Harness project ID (optional) |

### Debug Settings

| Parameter | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `debug` | `DEBUG` | boolean | false | Enable debug logging |
| `dry_run` | `DRY_RUN` | boolean | false | Log actions without posting |

## Comment Types

### 💬 PR Comment

Post a general comment on a pull request:

```yaml
settings:
  scm_provider: github
  token:
    from_secret: github_token
  repo: owner/repo
  pr_number: ${DRONE_PULL_REQUEST}
  comment_body: "Build completed successfully! 🎉"
```

### 📝 Inline Code Comment

Post a comment on a specific line of code:

```yaml
settings:
  scm_provider: github
  token:
    from_secret: github_token
  repo: owner/repo
  pr_number: ${DRONE_PULL_REQUEST}
  file_path: src/main.go
  line: 42
  comment_body: "Consider using a constant here"
```

### 📊 Commit Status

Set a commit status check:

```yaml
settings:
  scm_provider: github
  token:
    from_secret: github_token
  repo: owner/repo
  commit_sha: ${DRONE_COMMIT_SHA}
  status_state: success
  status_context: ci/lint
  status_desc: "Linting passed"
  status_url: ${DRONE_BUILD_LINK}
```

### 📁 Batch Comments from JSON File

Post multiple inline comments from a JSON file (perfect for AI code reviews):

```yaml
settings:
  scm_provider: harness
  token:
    from_secret: harness_token
  harness_account_id: ACCOUNT_ID
  repo: my-repo
  pr_number: ${PR_NUMBER}
  comments_file: /path/to/reviews.json
```

## JSON File Format

The `comments_file` should contain a JSON object with a `reviews` array:

```json
{
  "reviews": [
    {
      "file_path": "src/main.go",
      "line_number_start": 42,
      "line_number_end": 42,
      "type": "performance",
      "review": "Consider using a more efficient algorithm here"
    },
    {
      "file_path": "src/utils.go",
      "line_number_start": 100,
      "line_number_end": 105,
      "type": "bug",
      "review": "Potential null pointer dereference"
    }
  ]
}
```

### Review Object Fields

| Field | Type | Description |
|-------|------|-------------|
| `file_path` | string | Path to the file |
| `line_number_start` | integer | Starting line number |
| `line_number_end` | integer | Ending line number |
| `type` | string | Review type (e.g., `bug`, `performance`, `scalability`, `code_smell`) |
| `review` | string | The review comment text |

### Supported Review Types

| Type | Emoji | Description |
|------|-------|-------------|
| `bug` | 🐛 | Critical bugs and errors |
| `performance` | ⚡ | Performance issues |
| `scalability` | 📈 | Scalability concerns |
| `code_smell` | 🔍 | Code quality issues |
| *custom* | 📝 | Any custom category |

Comments are formatted with the type as a prefix: **`**performance:** Your review text`**

## Integration with AI Review Plugin

This plugin works seamlessly with [ai-review-prompt-plugin](https://github.com/abhinav-harness/ai-review-prompt-plugin):

```yaml
kind: pipeline
type: docker
name: ai-code-review

steps:
  # Step 1: Generate AI review prompt
  - name: generate-prompt
    image: abhinavharness/drone-ai-review:latest
    settings:
      enable_bugs: true
      enable_performance: true
      enable_scalability: true
      enable_code_smell: true
      review_output_file: /drone/src/reviews.json

  # Step 2: Run AI model (example with OpenAI)
  - name: run-ai-review
    image: your-ai-runner:latest
    commands:
      - cat /drone/src/output/task.txt | openai-api > /drone/src/reviews.json

  # Step 3: Post comments to PR
  - name: post-comments
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: harness
      token:
        from_secret: harness_token
      harness_account_id: ${HARNESS_ACCOUNT_ID}
      repo: ${DRONE_REPO_NAME}
      pr_number: ${DRONE_PULL_REQUEST}
      comments_file: /drone/src/reviews.json

trigger:
  event:
    - pull_request
```

## Harness CI Pipeline

### Basic Harness CI Example

```yaml
- step:
    type: Plugin
    name: Post Comment
    spec:
      connectorRef: dockerhub
      image: abhinavharness/comment-plugin
      settings:
        scm_provider: harness
        token: <+secrets.getValue("harness_token")>
        harness_account_id: <+account.identifier>
        repo: <+pipeline.properties.ci.codebase.repoName>
        pr_number: <+codebase.prNumber>
        comment_body: "Pipeline completed! ✅"
```

### Harness CI with AI Review

```yaml
- step:
    type: Plugin
    name: Post AI Reviews
    spec:
      connectorRef: dockerhub
      image: abhinavharness/comment-plugin
      settings:
        scm_provider: harness
        token: <+secrets.getValue("harness_token")>
        harness_account_id: <+account.identifier>
        repo: <+pipeline.properties.ci.codebase.repoName>
        pr_number: <+codebase.prNumber>
        comments_file: /harness/reviews.json
```

## Development

### Prerequisites

- Go 1.21 or higher
- Docker (with buildx for multi-arch builds)
- Git

### Local Development

1. Clone the repository:

```bash
git clone https://github.com/abhinav-harness/comment-plugin.git
cd comment-plugin
```

2. Install dependencies:

```bash
go mod download
```

3. Build the plugin:

```bash
go build -o comment-plugin ./cmd/plugin
```

4. Run tests:

```bash
go test -v ./...
```

5. Run locally:

```bash
export SCM_PROVIDER="github"
export TOKEN="your-token"
export REPO="owner/repo"
export PR_NUMBER="1"
export COMMENT_BODY="Test comment"
./comment-plugin
```

### Docker Build

```bash
# Build local image
docker build -t comment-plugin .

# Build multi-arch and push
docker buildx build --platform linux/amd64,linux/arm64 \
  -t abhinavharness/comment-plugin:latest \
  --push .
```

## Troubleshooting

### No comment posted

- Verify the token has write permissions to the repository
- Check that `pr_number` is set correctly
- Enable `debug: true` to see detailed logs

### Comments file not found

- The plugin handles missing files gracefully (logs warning and continues)
- Verify the file path is correct and accessible in the container
- Check that the previous step wrote the file successfully

### Harness Code 404 errors

- Ensure `harness_account_id` is set correctly
- Verify the token has access to the repository
- Check that the repository name matches exactly

### Inline comments appear as regular comments

- Ensure `file_path` and `line` are both provided
- For batch comments, verify `line_number_start` and `line_number_end` are set
- The file path must match a file changed in the PR

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Apache 2.0 License - see [LICENSE](LICENSE) file for details.

## Support

- **GitHub Issues**: https://github.com/abhinav-harness/comment-plugin/issues
- **Docker Hub**: https://hub.docker.com/r/abhinavharness/comment-plugin

## Related Projects

- [ai-review-prompt-plugin](https://github.com/abhinav-harness/ai-review-prompt-plugin) - Generate AI review prompts for code review
- [go-scm](https://github.com/drone/go-scm) - Unified SCM client library
