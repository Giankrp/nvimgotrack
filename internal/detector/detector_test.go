package detector

import (
	"testing"

	"github.com/Giankrp/nvimgotrack/internal/github"
)

func TestBreakingKeywords(t *testing.T) {
	tests := []struct {
		msg      string
		breaking bool
		deprec   bool
	}{
		{"feat: add new feature", false, false},
		{"fix: resolve issue", false, false},
		{"feat!: breaking change in API", true, false},
		{"BREAKING CHANGE: removed old function", true, false},
		{"refactor: this is incompatible with old version", true, false},
		{"chore: deprecated old module", false, true},
		{"fix: removed deprecated functionality, will be removed soon", true, true},
		{"docs: update migration required guide", true, false},
	}

	for _, tt := range tests {
		gotBreaking := breakingRe.MatchString(tt.msg) || featBangRe.MatchString(tt.msg)
		gotDeprec := deprecRe.MatchString(tt.msg)

		if gotBreaking != tt.breaking {
			t.Errorf("breaking(%q) = %v, want %v", tt.msg, gotBreaking, tt.breaking)
		}
		if gotDeprec != tt.deprec {
			t.Errorf("deprecated(%q) = %v, want %v", tt.msg, gotDeprec, tt.deprec)
		}
	}
}

func TestAnalyzeReleasesSeverity(t *testing.T) {
	releases := []github.Release{
		{TagName: "v2.0.0", Name: "BREAKING: Major rewrite", Body: "This is a breaking release"},
		{TagName: "v1.5.0", Name: "Feature release", Body: "Added new features"},
		{TagName: "v1.4.0", Name: "Deprecation notice", Body: "This API is deprecated and will change soon"},
	}

	infos := analyzeReleases(releases, "abc123")

	var foundBreaking, foundDeprecated bool
	for _, info := range infos {
		if info.Tag == "v2.0.0" && info.Severity == SeverityBreaking {
			foundBreaking = true
		}
		if info.Tag == "v1.4.0" && info.Severity == SeverityDeprecation {
			foundDeprecated = true
		}
	}

	if !foundBreaking {
		t.Error("expected v2.0.0 to be classified as breaking")
	}
	if !foundDeprecated {
		t.Error("expected v1.4.0 to be classified as deprecation")
	}
}

func TestSeverityString(t *testing.T) {
	if SeverityBreaking.Icon() != "🔴" {
		t.Errorf("breaking icon: got %s", SeverityBreaking.Icon())
	}
	if SeverityDeprecation.Icon() != "🟡" {
		t.Errorf("deprecation icon: got %s", SeverityDeprecation.Icon())
	}
	if SeverityOK.Icon() != "✅" {
		t.Errorf("ok icon: got %s", SeverityOK.Icon())
	}
}
