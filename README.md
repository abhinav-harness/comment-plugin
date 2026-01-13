# Comment Plugin

A Drone/Harness CI plugin to post comments on pull requests using [go-scm](https://github.com/drone/go-scm).

## Docker Image

```
docker pull abhinavharness/comment-plugin
```

## Supported Providers

| Provider | PR Comments | Inline Comments | Status |
|----------|-------------|-----------------|--------|
| GitHub | ✅ | ✅ | ✅ |
| GitLab | ✅ | ✅ | ✅ |
| Bitbucket | ✅ | ✅ | ✅ |
| Gitea | ✅ | ✅ | ✅ |
| Gogs | ✅ | ✅ | ✅ |
| **Harness Code** | ✅ | ✅ | ✅ |

## Usage

### PR Comment

```yaml
steps:
  - name: comment
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: github
      token:
        from_secret: github_token
      repo: owner/repo
      pr_number: ${DRONE_PULL_REQUEST}
      comment_body: "Build passed! ✅"
```

### Inline Code Comment

```yaml
steps:
  - name: comment
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: github
      token:
        from_secret: github_token
      repo: owner/repo
      pr_number: ${DRONE_PULL_REQUEST}
      file_path: src/main.go
      line: 42
      comment_body: "Consider refactoring this function"
```

### Commit Status

```yaml
steps:
  - name: status
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: github
      token:
        from_secret: github_token
      repo: owner/repo
      commit_sha: ${DRONE_COMMIT_SHA}
      status_state: success
      status_context: ci/build
      status_desc: Build passed
```

### Harness Code

```yaml
steps:
  - name: comment
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: harness
      token:
        from_secret: harness_token
      harness_account_id: ACCOUNT_ID
      repo: my-repo
      pr_number: ${PR_NUMBER}
      comment_body: "Pipeline complete! 🚀"
```

### Batch Comments from JSON File

Post multiple inline comments from a JSON file (useful for code review bots, linters, etc.):

```yaml
steps:
  - name: batch-comments
    image: abhinavharness/comment-plugin
    settings:
      scm_provider: harness
      token:
        from_secret: harness_token
      harness_account_id: ACCOUNT_ID
      repo: my-repo
      pr_number: ${PR_NUMBER}
      comments_file: /path/to/comments.json
```

**JSON file format:**

```json
[
  {
    "text": "Consider refactoring this function for better performance",
    "line_start": 42,
    "line_end": 42,
    "path": "src/main.go",
    "source_commit_sha": "abc123...",
    "target_commit_sha": "def456..."
  },
  {
    "text": "This could cause a memory leak",
    "line_start": 100,
    "line_end": 105,
    "path": "src/utils.go",
    "source_commit_sha": "abc123...",
    "target_commit_sha": "def456..."
  }
]
```

### Harness CI Pipeline

```yaml
- step:
    type: Plugin
    name: Comment
    spec:
      connectorRef: dockerhub
      image: abhinavharness/comment-plugin
      settings:
        scm_provider: harness
        token: <+secrets.getValue("harness_token")>
        harness_account_id: <+account.identifier>
        repo: <+pipeline.properties.ci.codebase.repoName>
        pr_number: <+codebase.prNumber>
        comment_body: "Build complete! ✅"
```

## Configuration

| Setting | Required | Description |
|---------|----------|-------------|
| `scm_provider` | ✅ | `github`, `gitlab`, `bitbucket`, `gitea`, `gogs`, `harness` |
| `token` | ✅ | Authentication token |
| `repo` | ✅ | Repository (`owner/repo` or just `repo` for Harness) |
| `pr_number` | | Pull request number (for comments) |
| `commit_sha` | | Commit SHA (for status) |
| `comment_body` | | Comment text |
| `file_path` | | File path (for inline comments) |
| `line` | | Line number (for inline comments) |
| `comments_file` | | Path to JSON file with array of comments |
| `status_state` | | `pending`, `success`, `failure`, `error` |
| `status_context` | | Status check name |
| `status_desc` | | Status description |
| `status_url` | | Link URL for status |
| `scm_endpoint` | | Custom API endpoint (for self-hosted) |

### Harness Code Settings

| Setting | Description |
|---------|-------------|
| `harness_account_id` | Harness account ID |
| `harness_org_id` | Harness org ID (optional) |
| `harness_project_id` | Harness project ID (optional) |

## Build

```bash
# Build binary
go build -o comment-plugin ./cmd/plugin

# Build Docker image
docker build -t comment-plugin .

# Run tests
go test ./...
```

## License

Apache 2.0
