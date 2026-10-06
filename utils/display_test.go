package utils

import (
	"strings"
	"testing"
)

// TestBanner checks the banner stays well-formed after the rename. Toolname is
// figlet ASCII art, so the tool name is not a literal substring; the version
// tag is the only literal marker that can be asserted here.
func TestBanner(t *testing.T) {
	if strings.TrimSpace(Toolname) == "" {
		t.Fatal("Toolname must not be empty")
	}
	if !strings.Contains(Toolname, "v0.3.0") {
		t.Errorf("Toolname banner missing version tag %q", "v0.3.0")
	}
	if strings.TrimSpace(Author) == "" {
		t.Error("Author must not be empty")
	}
	if strings.TrimSpace(Split) == "" {
		t.Error("Split must not be empty")
	}
}
