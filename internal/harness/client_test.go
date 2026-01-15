package harness

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewClient(t *testing.T) {
	// Suppress logging during tests
	logrus.SetLevel(logrus.PanicLevel)

	_, err := NewClient(Config{Token: "test"})
	if err != nil {
		t.Errorf("NewClient failed: %v", err)
	}

	_, err = NewClient(Config{})
	if err == nil {
		t.Error("NewClient should fail without token")
	}
}

func TestApiPathWithAllParams(t *testing.T) {
	// Suppress logging during tests
	logrus.SetLevel(logrus.PanicLevel)

	c, _ := NewClient(Config{Token: "test", OrgID: "org", ProjectID: "proj", AccountID: "acc"})
	path := c.apiPath("repo", "pullreq/1/comments")

	if !strings.Contains(path, "routingId=acc") {
		t.Errorf("apiPath should contain routingId, got: %s", path)
	}
	if !strings.Contains(path, "accountIdentifier=acc") {
		t.Errorf("apiPath should contain accountIdentifier, got: %s", path)
	}
	if !strings.Contains(path, "orgIdentifier=org") {
		t.Errorf("apiPath should contain orgIdentifier, got: %s", path)
	}
	if !strings.Contains(path, "projectIdentifier=proj") {
		t.Errorf("apiPath should contain projectIdentifier, got: %s", path)
	}
}

func TestApiPathWithoutAccountId(t *testing.T) {
	// Suppress logging during tests
	logrus.SetLevel(logrus.PanicLevel)

	c, _ := NewClient(Config{Token: "test", OrgID: "org", ProjectID: "proj"})
	path := c.apiPath("repo", "pullreq/1/comments")

	if strings.Contains(path, "routingId") {
		t.Errorf("apiPath should not contain routingId when AccountID is empty, got: %s", path)
	}
	if strings.Contains(path, "accountIdentifier") {
		t.Errorf("apiPath should not contain accountIdentifier when AccountID is empty, got: %s", path)
	}
	if !strings.Contains(path, "orgIdentifier=org") {
		t.Errorf("apiPath should contain orgIdentifier, got: %s", path)
	}
	if !strings.Contains(path, "projectIdentifier=proj") {
		t.Errorf("apiPath should contain projectIdentifier, got: %s", path)
	}
}

func TestApiPathRepoInUrl(t *testing.T) {
	// Suppress logging during tests
	logrus.SetLevel(logrus.PanicLevel)

	c, _ := NewClient(Config{Token: "test", AccountID: "acc"})
	path := c.apiPath("my-repo", "pullreq/1/comments")

	if !strings.Contains(path, "/repos/my-repo/pullreq") {
		t.Errorf("apiPath should contain repo name in path, got: %s", path)
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
