package harness

import (
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient(Config{Token: "test"})
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	_, err = NewClient(Config{})
	if err == nil {
		t.Error("NewClient should fail without token")
	}
}

func TestBuildRepoPath(t *testing.T) {
	c := &Client{config: Config{OrgID: "org", ProjectID: "proj"}}

	tests := []struct {
		repo, expected string
	}{
		{"repo", "org/proj/+/repos/repo"},
		{"p/repo", "org/p/+/repos/repo"},
		{"o/p/repo", "o/p/+/repos/repo"},
	}

	for _, tt := range tests {
		if got := c.buildRepoPath(tt.repo); got != tt.expected {
			t.Errorf("buildRepoPath(%q) = %q, want %q", tt.repo, got, tt.expected)
		}
	}
}

func TestApiPathWithRoutingId(t *testing.T) {
	c := &Client{baseURL: "https://app.harness.io", config: Config{OrgID: "o", ProjectID: "p", AccountID: "acc"}}
	path := c.apiPath("repo", "pullreq/1/comments")

	if !strings.Contains(path, "routingId=acc") {
		t.Errorf("apiPath should contain routingId, got: %s", path)
	}
}

func TestMapState(t *testing.T) {
	tests := map[string]string{
		"success": "success",
		"failure": "failure",
		"failed":  "failure",
		"error":   "error",
		"pending": "pending",
		"unknown": "pending",
	}

	for input, expected := range tests {
		if got := mapState(input); got != expected {
			t.Errorf("mapState(%q) = %q, want %q", input, got, expected)
		}
	}
}
