package plugin

import (
	"testing"

	"github.com/drone/go-scm/scm"
)

func TestMapStatusState(t *testing.T) {
	tests := map[string]scm.State{
		"success": scm.StateSuccess,
		"failure": scm.StateFailure,
		"error":   scm.StateError,
		"pending": scm.StatePending,
		"running": scm.StateRunning,
		"unknown": scm.StateUnknown,
	}

	for input, expected := range tests {
		if got := mapStatusState(input); got != expected {
			t.Errorf("mapStatusState(%q) = %v, want %v", input, got, expected)
		}
	}
}
