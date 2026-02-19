package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	sinceTag := flag.String("since-tag", "", "Previous version tag (e.g., v0.2.14)")
	repo := flag.String("repo", "Azure/kubelogin", "Repository in format owner/repo")
	output := flag.String("output", "changelog-entry.md", "Output file path")
	flag.Parse()

	if *version == "" || *sinceTag == "" {
		log.Fatal("Both --version and --since-tag are required")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	ctx := context.Background()

	// Get the date of the previous tag
	tagDate, err := getTagDate(ctx, *repo, *sinceTag, token)
	if err != nil {
		log.Fatalf("Failed to get tag date: %v", err)
	}

	// Fetch merged PRs since the tag
	prs, err := getMergedPRsSince(ctx, *repo, tagDate, token)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) == 0 {
		log.Println("No merged PRs found since", *sinceTag)
	}

	// Get all contributors before this tag to identify new ones
	allContributorsBefore, err := getAllContributorsBefore(ctx, *repo, tagDate, token)
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

func getTagDate(ctx context.Context, repo, tag, token string) (time.Time, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags/%s", repo, strings.TrimPrefix(tag, "refs/tags/"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return time.Time{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return time.Time{}, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var refData struct {
		Object struct {
			SHA string `json:"sha"`
			URL string `json:"url"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refData); err != nil {
		return time.Time{}, err
	}

	// Get the commit date
	req, err = http.NewRequestWithContext(ctx, "GET", refData.Object.URL, nil)
	if err != nil {
		return time.Time{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	var commitData struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
		Author struct {
			Date time.Time `json:"date"`
		} `json:"author"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commitData); err != nil {
		return time.Time{}, err
	}

	// Use author date as it's more reliable
	if !commitData.Author.Date.IsZero() {
		return commitData.Author.Date, nil
	}
	return commitData.Committer.Date, nil
}

func getMergedPRsSince(ctx context.Context, repo string, since time.Time, token string) ([]GitHubPR, error) {
	var allPRs []GitHubPR
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=closed&sort=updated&direction=desc&per_page=%d&page=%d",
			repo, perPage, page)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
		}

		var prs []GitHubPR
		if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
			return nil, err
		}

		if len(prs) == 0 {
			break
		}

		for _, pr := range prs {
			// Only include merged PRs after the since date
			if !pr.MergedAt.IsZero() && pr.MergedAt.After(since) {
				allPRs = append(allPRs, pr)
			}
		}

		// If we've gone past the since date, we can stop
		if len(prs) > 0 && prs[len(prs)-1].MergedAt.Before(since) {
			break
		}

		page++
		if page > 10 { // Safety limit
			break
		}
	}

	return allPRs, nil
}

func getAllContributorsBefore(ctx context.Context, repo string, before time.Time, token string) (map[string]bool, error) {
	contributors := make(map[string]bool)
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=closed&sort=created&direction=asc&per_page=%d&page=%d",
			repo, perPage, page)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return contributors, nil // Return what we have so far
		}

		var prs []GitHubPR
		if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
			return contributors, nil
		}

		if len(prs) == 0 {
			break
		}

		for _, pr := range prs {
			if !pr.MergedAt.IsZero() && pr.MergedAt.Before(before) {
				contributors[pr.User.Login] = true
			}
			// Stop if we've reached PRs created after the before date
			if pr.CreatedAt.After(before) {
				return contributors, nil
			}
		}

		page++
		if page > 50 { // Safety limit - adjust based on repo size
			break
		}
	}

	return contributors, nil
}

// PRCategory represents a categorized group of PRs
type PRCategory struct {
	Name string
	PRs  []GitHubPR
}

// Categories holds all categorized PRs
type Categories struct {
	Changes         []GitHubPR
	Maintenance     []GitHubPR
	Enhancements    []GitHubPR
	DocUpdates      []GitHubPR
	NewContributors []Contributor
}

func categorizePRs(prs []GitHubPR, existingContributors map[string]bool) Categories {
	cats := Categories{
		Changes:         make([]GitHubPR, 0),
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
	if strings.HasPrefix(title, "bump ") ||
		strings.HasPrefix(title, "update ") ||
		strings.Contains(title, "cve-") ||
		strings.Contains(title, "fix cve") ||
		strings.Contains(title, "dependencies") ||
		strings.HasPrefix(title, "chore") {
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
