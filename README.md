# Comment Plugin

[![Build](https://github.com/abhinav-harness/comment-plugin/actions/workflows/build.yml/badge.svg)](https://github.com/abhinav-harness/comment-plugin)

Drone plugin to post comments on PRs using [go-scm](https://github.com/drone/go-scm).

## Supported Providers

- GitHub / GitHub Enterprise
- GitLab
- Bitbucket Cloud / Server
- Gitea / Gogs
- **Harness Code**

## Usage

### PR Comment

```yaml
- name: comment
  image: plugins/comment
  settings:
    scm_provider: github
    token:
      from_secret: github_token
    repo: owner/repo
    pr_number: ${DRONE_PULL_REQUEST}
    comment_body: "Build passed! ✅"
```

### Inline Comment

```yaml
- name: comment
  image: plugins/comment
  settings:
    scm_provider: github
    token:
      from_secret: github_token
    repo: owner/repo
    pr_number: ${DRONE_PULL_REQUEST}
    file_path: src/main.go
    line: 42
    comment_body: "Consider refactoring"
```

### Commit Status

```yaml
- name: status
  image: plugins/comment
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
- name: comment
  image: plugins/comment
  settings:
    scm_provider: harness
    token:
      from_secret: harness_token
    harness_account_id: ${ACCOUNT_ID}
    harness_org_id: ${ORG_ID}
    harness_project_id: ${PROJECT_ID}
    repo: my-repo
    pr_number: ${PR_NUMBER}
    comment_body: "Pipeline complete! 🚀"
```

## Configuration

| Setting | Required | Description |
|---------|----------|-------------|
| `scm_provider` | ✅ | `github`, `gitlab`, `bitbucket`, `gitea`, `gogs`, `harness` |
| `token` | ✅ | Auth token |
| `repo` | ✅ | `owner/repo` |
| `pr_number` | | PR number |
| `commit_sha` | | Commit SHA (for status) |
| `comment_body` | | Comment text |
| `file_path` | | File path (inline) |
| `line` | | Line number (inline) |
| `status_state` | | `pending`, `success`, `failure`, `error` |
| `status_context` | | Status name |
| `status_desc` | | Status description |
| `scm_endpoint` | | Custom API endpoint |

### Harness Settings

| Setting | Description |
|---------|-------------|
| `harness_account_id` | Account ID |
| `harness_org_id` | Org ID |
| `harness_project_id` | Project ID |

## Build

```bash
go build -o comment-plugin ./cmd/plugin
docker build -t plugins/comment .
```
