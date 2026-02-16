package github

import "time"

// Release represents a GitHub release.
type Release struct {
	TagName      string    `json:"tag_name"`
	Name         string    `json:"name"`
	Body         string    `json:"body"`
	Draft        bool      `json:"draft"`
	Prerelease   bool      `json:"prerelease"`
	PublishedAt  time.Time `json:"published_at"`
	HTMLURL      string    `json:"html_url"`
	TargetCommit string    `json:"target_commitish"`
}

// Commit represents a GitHub commit (simplified).
type Commit struct {
	SHA     string       `json:"sha"`
	Commit  CommitDetail `json:"commit"`
	HTMLURL string       `json:"html_url"`
}

// CommitDetail holds the commit message and author info.
type CommitDetail struct {
	Message string       `json:"message"`
	Author  CommitAuthor `json:"author"`
}

// CommitAuthor holds commit author metadata.
type CommitAuthor struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

// CompareResult represents the result of comparing two commits.
type CompareResult struct {
	Status       string   `json:"status"` // "ahead", "behind", "diverged", "identical"
	AheadBy      int      `json:"ahead_by"`
	BehindBy     int      `json:"behind_by"`
	TotalCommits int      `json:"total_commits"`
	Commits      []Commit `json:"commits"`
	HTMLURL      string   `json:"html_url"`
}

// RepoInfo holds basic repository metadata.
type RepoInfo struct {
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
	Description   string `json:"description"`
}
