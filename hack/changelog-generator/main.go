package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

// GitHubPR represents a GitHub pull request
type GitHubPR struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	MergedAt  time.Time `json:"merged_at"`
	Labels    []Label   `json:"labels"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// Contributor tracks first-time contributors
type Contributor struct {
	Login  string
	PRURL  string
	Number int
}

func main() {
	version := flag.String("version", "", "Version number (e.g., 0.2.15)")
	sinceTag := flag.String("since-tag", "", "Previous version tag (e.g., v0.2.14); defaults to the latest tag")
	repo := flag.String("repo", "Azure/kubelogin", "Repository in format owner/repo")
	output := flag.String("output", "changelog-entry.md", "Output file path")
	flag.Parse()

	if *version == "" {
		log.Fatal("--version is required")
	}

	// Resolve the previous tag if not provided
	if *sinceTag == "" {
		resolved, err := getLatestTag(*repo)
		if err != nil {
			log.Fatalf("Failed to resolve previous tag: %v", err)
		}
		log.Printf("No --since-tag provided; using latest tag: %s", resolved)
		*sinceTag = resolved
	}

	// Get the date of the previous tag
	tagDate, err := getTagDate(*repo, *sinceTag)
	if err != nil {
		log.Fatalf("Failed to get tag date: %v", err)
	}

	// Fetch merged PRs since the tag
	prs, err := getMergedPRsSince(*repo, tagDate)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) == 0 {
		log.Println("No merged PRs found since", *sinceTag)
	}

	// Get all contributors before this tag to identify new ones
	allContributorsBefore, err := getAllContributorsBefore(*repo, tagDate)
	if err != nil {
		log.Printf("Warning: Failed to get historical contributors: %v", err)
		allContributorsBefore = make(map[string]bool)
	}

	// Categorize PRs and identify new contributors
	categories := categorizePRs(prs, allContributorsBefore)

	// Generate the changelog entry
	entry := generateChangelogEntry(*version, *sinceTag, categories, *repo)

	// Write to output file
	if err := os.WriteFile(*output, []byte(entry), 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	log.Printf("Successfully generated changelog entry for version %s", *version)
}

// ghAPI runs "gh api <args>" and returns the output.
// Authentication is handled automatically by the gh CLI
// (GITHUB_TOKEN env var or the credential stored by "gh auth login").
func ghAPI(args ...string) ([]byte, error) {
	out, err := exec.Command("gh", append([]string{"api"}, args...)...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh api %v: %w\n%s", args, err, exitErr.Stderr)
		}
		return nil, fmt.Errorf("gh api %v: %w", args, err)
	}
	return out, nil
}

// getLatestTag returns the tag of the most recent stable release.
func getLatestTag(repo string) (string, error) {
	out, err := ghAPI(fmt.Sprintf("repos/%s/releases/latest", repo), "--jq", ".tag_name")
	if err != nil {
		return "", err
	}
	tag := strings.TrimSpace(string(out))
	if tag == "" {
		return "", fmt.Errorf("no releases found in repository %s", repo)
	}
	return tag, nil
}

// getTagDate returns the author date of the commit the tag points to.
func getTagDate(repo, tag string) (time.Time, error) {
	out, err := ghAPI(
		fmt.Sprintf("repos/%s/commits/%s", repo, tag),
		"--jq", ".commit.author.date",
	)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
}

// decodePRStream decodes newline-delimited JSON objects produced by
// "gh api --paginate ... --jq '.[]'".
func decodePRStream(data []byte) ([]GitHubPR, error) {
	var prs []GitHubPR
	dec := json.NewDecoder(strings.NewReader(string(data)))
	for dec.More() {
		var pr GitHubPR
		if err := dec.Decode(&pr); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// releasePRTitle matches PR titles that represent a version release
// (e.g. "v0.2.14 release") so they can be excluded from the changelog.
var releasePRTitle = regexp.MustCompile(`(?i)^v?\d+\.\d+\.\d+`)

// isReleasePR returns true when the PR title looks like a release commit
// (e.g. "v0.2.14 release", "0.2.14 release").
func isReleasePR(title string) bool {
	return releasePRTitle.MatchString(strings.TrimSpace(title))
}

// getMergedPRsSince returns all merged PRs after the given time.
func getMergedPRsSince(repo string, since time.Time) ([]GitHubPR, error) {
	out, err := ghAPI(
		"--paginate",
		fmt.Sprintf("repos/%s/pulls?state=closed&sort=created&direction=desc&per_page=100", repo),
		"--jq", ".[]",
	)
	if err != nil {
		return nil, err
	}
	all, err := decodePRStream(out)
	if err != nil {
		return nil, err
	}
	var prs []GitHubPR
	for _, pr := range all {
		if !pr.MergedAt.IsZero() && pr.MergedAt.After(since) && !isReleasePR(pr.Title) {
			prs = append(prs, pr)
		}
	}
	return prs, nil
}

// getAllContributorsBefore returns the set of logins that contributed before the given time.
func getAllContributorsBefore(repo string, before time.Time) (map[string]bool, error) {
	out, err := ghAPI(
		"--paginate",
		fmt.Sprintf("repos/%s/pulls?state=closed&sort=created&direction=asc&per_page=100", repo),
		"--jq", ".[]",
	)
	if err != nil {
		return nil, err
	}
	all, err := decodePRStream(out)
	if err != nil {
		return nil, err
	}
	contributors := make(map[string]bool)
	for _, pr := range all {
		if !pr.MergedAt.IsZero() && pr.MergedAt.Before(before) {
			contributors[pr.User.Login] = true
		}
	}
	return contributors, nil
}

// Categories holds all categorized PRs
type Categories struct {
	Changes         []GitHubPR
	BugFixes        []GitHubPR
	Maintenance     []GitHubPR
	Enhancements    []GitHubPR
	DocUpdates      []GitHubPR
	NewContributors []Contributor
}

func categorizePRs(prs []GitHubPR, existingContributors map[string]bool) Categories {
	cats := Categories{
		Changes:         make([]GitHubPR, 0),
		BugFixes:        make([]GitHubPR, 0),
		Maintenance:     make([]GitHubPR, 0),
		Enhancements:    make([]GitHubPR, 0),
		DocUpdates:      make([]GitHubPR, 0),
		NewContributors: make([]Contributor, 0),
	}

	seenNewContributors := make(map[string]bool)

	for _, pr := range prs {
		// Check for new contributors
		if !existingContributors[pr.User.Login] && !seenNewContributors[pr.User.Login] {
			cats.NewContributors = append(cats.NewContributors, Contributor{
				Login:  pr.User.Login,
				PRURL:  pr.HTMLURL,
				Number: pr.Number,
			})
			seenNewContributors[pr.User.Login] = true
		}

		// Categorize based on labels and title
		category := categorizeByLabelsAndTitle(pr)
		switch category {
		case "bugfix":
			cats.BugFixes = append(cats.BugFixes, pr)
		case "maintenance":
			cats.Maintenance = append(cats.Maintenance, pr)
		case "enhancement":
			cats.Enhancements = append(cats.Enhancements, pr)
		case "documentation":
			cats.DocUpdates = append(cats.DocUpdates, pr)
		default:
			cats.Changes = append(cats.Changes, pr)
		}
	}

	return cats
}

func categorizeByLabelsAndTitle(pr GitHubPR) string {
	title := strings.ToLower(pr.Title)

	// Check labels first
	for _, label := range pr.Labels {
		labelName := strings.ToLower(label.Name)
		if strings.Contains(labelName, "bug") ||
			strings.Contains(labelName, "fix") {
			return "bugfix"
		}
		if strings.Contains(labelName, "maintenance") ||
			strings.Contains(labelName, "dependencies") ||
			strings.Contains(labelName, "chore") {
			return "maintenance"
		}
		if strings.Contains(labelName, "enhancement") ||
			strings.Contains(labelName, "feature") {
			return "enhancement"
		}
		if strings.Contains(labelName, "documentation") ||
			strings.Contains(labelName, "docs") {
			return "documentation"
		}
	}

	// Check title patterns
	if strings.HasPrefix(title, "fix:") ||
		strings.HasPrefix(title, "bugfix:") ||
		strings.HasPrefix(title, "bug fix:") ||
		strings.HasPrefix(title, "hotfix:") {
		return "bugfix"
	}

	if strings.HasPrefix(title, "bump ") ||
		strings.HasPrefix(title, "update ") ||
		strings.Contains(title, "cve-") ||
		strings.Contains(title, "fix cve") ||
		strings.Contains(title, "dependencies") ||
		strings.HasPrefix(title, "chore:") ||
		strings.HasPrefix(title, "chore ") {
		return "maintenance"
	}

	if strings.HasPrefix(title, "docs:") ||
		strings.HasPrefix(title, "doc:") ||
		strings.Contains(title, "documentation") ||
		strings.Contains(title, "install doc") {
		return "documentation"
	}

	if strings.HasPrefix(title, "feat:") ||
		strings.HasPrefix(title, "feature:") ||
		strings.Contains(title, "add support") ||
		strings.Contains(title, "new feature") {
		return "enhancement"
	}

	return "change"
}

func generateChangelogEntry(version, sinceTag string, cats Categories, repo string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## [%s]\n", version))

	// What's Changed
	if len(cats.Changes) > 0 {
		sb.WriteString("\n### What's Changed\n\n")
		for _, pr := range cats.Changes {
			sb.WriteString(fmt.Sprintf("* %s by @%s in %s\n", pr.Title, pr.User.Login, pr.HTMLURL))
		}
	}

	// Enhancements
	if len(cats.Enhancements) > 0 {
		sb.WriteString("\n### Enhancements\n\n")
		for _, pr := range cats.Enhancements {
			sb.WriteString(fmt.Sprintf("* %s by @%s in %s\n", pr.Title, pr.User.Login, pr.HTMLURL))
		}
	}

	// Bug Fixes
	if len(cats.BugFixes) > 0 {
		sb.WriteString("\n### Bug Fixes\n\n")
		for _, pr := range cats.BugFixes {
			sb.WriteString(fmt.Sprintf("* %s by @%s in %s\n", pr.Title, pr.User.Login, pr.HTMLURL))
		}
	}

	// Maintenance
	if len(cats.Maintenance) > 0 {
		sb.WriteString("\n### Maintenance\n\n")
		for _, pr := range cats.Maintenance {
			sb.WriteString(fmt.Sprintf("* %s by @%s in %s\n", pr.Title, pr.User.Login, pr.HTMLURL))
		}
	}

	// Doc Updates
	if len(cats.DocUpdates) > 0 {
		sb.WriteString("\n### Doc Update\n\n")
		for _, pr := range cats.DocUpdates {
			sb.WriteString(fmt.Sprintf("* %s by @%s in %s\n", pr.Title, pr.User.Login, pr.HTMLURL))
		}
	}

	// New Contributors
	if len(cats.NewContributors) > 0 {
		sb.WriteString("\n### New Contributors\n\n")
		// Sort by username for consistency
		sort.Slice(cats.NewContributors, func(i, j int) bool {
			return cats.NewContributors[i].Login < cats.NewContributors[j].Login
		})
		for _, c := range cats.NewContributors {
			sb.WriteString(fmt.Sprintf("* @%s made their first contribution in %s\n", c.Login, c.PRURL))
		}
	}

	// Full Changelog link
	currentTag := "v" + version
	sb.WriteString(fmt.Sprintf("\n**Full Changelog**: https://github.com/%s/compare/%s...%s\n",
		repo, sinceTag, currentTag))

	return sb.String()
}
