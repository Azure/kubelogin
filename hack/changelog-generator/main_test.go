package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCategorizeByLabelsAndTitle(t *testing.T) {
	tests := []struct {
		name     string
		pr       GitHubPR
		expected string
	}{
		// Label-based categorization takes precedence
		{name: "bug label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "bug"}}}, expected: "bugfix"},
		{name: "fix label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "fix"}}}, expected: "bugfix"},
		{name: "enhancement label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "enhancement"}}}, expected: "enhancement"},
		{name: "feature label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "feature"}}}, expected: "enhancement"},
		{name: "dependencies label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "dependencies"}}}, expected: "maintenance"},
		{name: "documentation label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "documentation"}}}, expected: "documentation"},
		{name: "docs label", pr: GitHubPR{Title: "something", Labels: []Label{{Name: "docs"}}}, expected: "documentation"},

		// Title prefix — bug fixes
		{name: "fix: prefix", pr: GitHubPR{Title: "fix: nil pointer"}, expected: "bugfix"},
		{name: "bugfix: prefix", pr: GitHubPR{Title: "bugfix: something"}, expected: "bugfix"},
		{name: "bug fix: prefix", pr: GitHubPR{Title: "bug fix: something"}, expected: "bugfix"},
		{name: "hotfix: prefix", pr: GitHubPR{Title: "hotfix: something"}, expected: "bugfix"},

		// Title prefix — maintenance
		{name: "bump prefix", pr: GitHubPR{Title: "Bump Go to 1.24"}, expected: "maintenance"},
		{name: "update prefix", pr: GitHubPR{Title: "Update dependency"}, expected: "maintenance"},
		{name: "cve in title", pr: GitHubPR{Title: "address cve-2024-1234"}, expected: "maintenance"},
		{name: "fix cve", pr: GitHubPR{Title: "fix cve issues"}, expected: "maintenance"},
		{name: "chore: prefix", pr: GitHubPR{Title: "chore: tidy modules"}, expected: "maintenance"},
		{name: "chore space prefix", pr: GitHubPR{Title: "chore bump version"}, expected: "maintenance"},
		{name: "choreography not matched", pr: GitHubPR{Title: "choreography work"}, expected: "change"},

		// Title prefix — documentation
		{name: "docs: prefix", pr: GitHubPR{Title: "docs: update readme"}, expected: "documentation"},
		{name: "doc: prefix", pr: GitHubPR{Title: "doc: fix typo"}, expected: "documentation"},

		// Title prefix — enhancements
		{name: "feat: prefix", pr: GitHubPR{Title: "feat: add new auth"}, expected: "enhancement"},
		{name: "feature: prefix", pr: GitHubPR{Title: "feature: new login"}, expected: "enhancement"},
		{name: "add support", pr: GitHubPR{Title: "add support for X"}, expected: "enhancement"},

		// Default
		{name: "unrecognized", pr: GitHubPR{Title: "Refactor internal logic"}, expected: "change"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := categorizeByLabelsAndTitle(tc.pr)
			if got != tc.expected {
				t.Errorf("categorizeByLabelsAndTitle(%q, labels=%v) = %q; want %q",
					tc.pr.Title, tc.pr.Labels, got, tc.expected)
			}
		})
	}
}

func prURL(n int) string {
	return fmt.Sprintf("https://github.com/Azure/kubelogin/pull/%d", n)
}

func TestCategorizePRs(t *testing.T) {
	now := time.Now()
	prs := []GitHubPR{
		{Number: 1, Title: "feat: add thing", User: User{Login: "alice"}, HTMLURL: prURL(1), MergedAt: now},
		{Number: 2, Title: "fix: nil pointer", User: User{Login: "bob"}, HTMLURL: prURL(2), MergedAt: now},
		{Number: 3, Title: "Bump Go to 1.24", User: User{Login: "dependabot"}, HTMLURL: prURL(3), MergedAt: now},
		{Number: 4, Title: "docs: update readme", User: User{Login: "carol"}, HTMLURL: prURL(4), MergedAt: now},
		{Number: 5, Title: "General change", User: User{Login: "alice"}, HTMLURL: prURL(5), MergedAt: now},
	}
	// bob and carol are new; alice and dependabot are existing
	existing := map[string]bool{"alice": true, "dependabot": true}

	cats := categorizePRs(prs, existing)

	if len(cats.Enhancements) != 1 || cats.Enhancements[0].Number != 1 {
		t.Errorf("expected 1 enhancement (PR#1), got %d", len(cats.Enhancements))
	}
	if len(cats.BugFixes) != 1 || cats.BugFixes[0].Number != 2 {
		t.Errorf("expected 1 bug fix (PR#2), got %d", len(cats.BugFixes))
	}
	if len(cats.Maintenance) != 1 || cats.Maintenance[0].Number != 3 {
		t.Errorf("expected 1 maintenance (PR#3), got %d", len(cats.Maintenance))
	}
	if len(cats.DocUpdates) != 1 || cats.DocUpdates[0].Number != 4 {
		t.Errorf("expected 1 doc update (PR#4), got %d", len(cats.DocUpdates))
	}
	if len(cats.Changes) != 1 || cats.Changes[0].Number != 5 {
		t.Errorf("expected 1 general change (PR#5), got %d", len(cats.Changes))
	}

	// New contributors: bob (PR#2) and carol (PR#4)
	if len(cats.NewContributors) != 2 {
		t.Errorf("expected 2 new contributors, got %d", len(cats.NewContributors))
	}
	newLogins := map[string]bool{}
	for _, c := range cats.NewContributors {
		newLogins[c.Login] = true
	}
	if !newLogins["bob"] || !newLogins["carol"] {
		t.Errorf("expected bob and carol as new contributors, got %v", newLogins)
	}
}

func TestGenerateChangelogEntry(t *testing.T) {
	now := time.Now()
	cats := Categories{
		Enhancements:    []GitHubPR{{Number: 1, Title: "feat: new thing", User: User{Login: "alice"}, HTMLURL: "https://github.com/Azure/kubelogin/pull/1", MergedAt: now}},
		BugFixes:        []GitHubPR{{Number: 2, Title: "fix: crash", User: User{Login: "bob"}, HTMLURL: "https://github.com/Azure/kubelogin/pull/2", MergedAt: now}},
		NewContributors: []Contributor{{Login: "bob", PRURL: "https://github.com/Azure/kubelogin/pull/2"}},
	}

	entry := generateChangelogEntry("0.2.15", "v0.2.14", cats, "Azure/kubelogin")

	for _, want := range []string{
		"## [0.2.15]",
		"### Enhancements",
		"feat: new thing",
		"@alice",
		"### Bug Fixes",
		"fix: crash",
		"@bob",
		"### New Contributors",
		"@bob made their first contribution",
		"**Full Changelog**: https://github.com/Azure/kubelogin/compare/v0.2.14...v0.2.15",
	} {
		if !strings.Contains(entry, want) {
			t.Errorf("expected changelog entry to contain %q\nGot:\n%s", want, entry)
		}
	}
}

func TestIsReleasePR(t *testing.T) {
	cases := []struct {
		title    string
		expected bool
	}{
		{"v0.2.14 release", true},
		{"0.2.14 release", true},
		{"v1.0.0", true},
		{"v0.2.14", true},
		// Regular PRs — must NOT be filtered
		{"fix: nil pointer", false},
		{"feat: add new login flow", false},
		{"Bump Go to 1.24.11", false},
		{"docs: update readme", false},
		{"[Bug Fix] - PoP token crash", false},
		{"chore: update CHANGELOG.md for v0.2.15", false},
	}
	for _, tc := range cases {
		got := isReleasePR(tc.title)
		if got != tc.expected {
			t.Errorf("isReleasePR(%q) = %v; want %v", tc.title, got, tc.expected)
		}
	}
}

func TestHasLabel(t *testing.T) {
	pr := GitHubPR{Labels: []Label{{Name: "release"}, {Name: "chore"}}}
	if !hasLabel(pr, "release") {
		t.Error("expected hasLabel to return true for 'release'")
	}
	if !hasLabel(pr, "Release") {
		t.Error("expected hasLabel to be case-insensitive")
	}
	if hasLabel(pr, "bug") {
		t.Error("expected hasLabel to return false for 'bug'")
	}
	empty := GitHubPR{}
	if hasLabel(empty, "release") {
		t.Error("expected hasLabel to return false for PR with no labels")
	}
}
