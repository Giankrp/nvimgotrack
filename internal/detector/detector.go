package detector

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Giankrp/nvimgotrack/internal/github"
	"github.com/Giankrp/nvimgotrack/internal/parser"
)

type Severity int

const (
	SeverityOK          Severity = iota 
	SeverityFeature                     
	SeverityDeprecation                
	SeverityBreaking                    
)

func (s Severity) String() string {
	switch s {
	case SeverityBreaking:
		return "🔴 BREAKING"
	case SeverityDeprecation:
		return "🟡 DEPRECATED"
	case SeverityFeature:
		return "🟢 Feature"
	default:
		return "✅ OK"
	}
}

func (s Severity) Icon() string {
	switch s {
	case SeverityBreaking:
		return "🔴"
	case SeverityDeprecation:
		return "🟡"
	case SeverityFeature:
		return "🟢"
	default:
		return "✅"
	}
}

type PluginReport struct {
	Plugin       parser.Plugin
	Severity     Severity
	BehindBy     int
	Releases     []ReleaseInfo
	BreakingMsgs []string
	DeprecMsgs   []string
	Error        string
	CompareURL   string
}

type ReleaseInfo struct {
	Tag      string
	Name     string
	Body     string
	URL      string
	Severity Severity
}

var (
	breakingRe = regexp.MustCompile(`(?i)\b(breaking|BREAKING CHANGE|incompatible|removed|migration required)\b`)
	deprecRe   = regexp.MustCompile(`(?i)\b(deprecated|deprecation|will be removed|no longer supported)\b`)
	semverRe   = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)
	featBangRe = regexp.MustCompile(`(?m)^(feat|fix|refactor|chore)!:`)
)

func Analyze(client *github.Client, plugin parser.Plugin) PluginReport {
	report := PluginReport{Plugin: plugin}

	// 1. Compare commits
	compare, err := client.CompareCommits(plugin.Owner, plugin.Repo, plugin.Commit, plugin.Branch)
	if err != nil {
		report.Error = fmt.Sprintf("compare failed: %v", err)
		return report
	}

	report.BehindBy = compare.TotalCommits
	report.CompareURL = compare.HTMLURL

	if compare.TotalCommits == 0 {
		report.Severity = SeverityOK
		return report
	}

	for _, c := range compare.Commits {
		msg := c.Commit.Message
		firstLine := strings.SplitN(msg, "\n", 2)[0]

		if breakingRe.MatchString(msg) || featBangRe.MatchString(msg) {
			report.BreakingMsgs = append(report.BreakingMsgs, firstLine)
		} else if deprecRe.MatchString(msg) {
			report.DeprecMsgs = append(report.DeprecMsgs, firstLine)
		}
	}
	releases, err := client.GetReleases(plugin.Owner, plugin.Repo)
	if err == nil {
		report.Releases = analyzeReleases(releases, plugin.Commit)
	}

	report.Severity = SeverityOK
	if report.BehindBy > 0 {
		report.Severity = SeverityFeature
	}

	for _, r := range report.Releases {
		if r.Severity > report.Severity {
			report.Severity = r.Severity
		}
	}

	if len(report.BreakingMsgs) > 0 {
		report.Severity = SeverityBreaking
	} else if len(report.DeprecMsgs) > 0 && report.Severity < SeverityDeprecation {
		report.Severity = SeverityDeprecation
	}

	return report
}

func analyzeReleases(releases []github.Release, _ string) []ReleaseInfo {
	infos := make([]ReleaseInfo, 0, len(releases))

	// Sort releases by published date, newest first
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].PublishedAt.After(releases[j].PublishedAt)
	})

	var prevMajor int = -1
	for i, r := range releases {
		if r.Draft {
			continue
		}

		info := ReleaseInfo{
			Tag:      r.TagName,
			Name:     r.Name,
			Body:     r.Body,
			URL:      r.HTMLURL,
			Severity: SeverityFeature,
		}

		// Check for semver major bumps
		if matches := semverRe.FindStringSubmatch(r.TagName); matches != nil {
			major, _ := strconv.Atoi(matches[1])
			if i > 0 && prevMajor >= 0 && major > prevMajor {
				info.Severity = SeverityBreaking
			}
			prevMajor = major
		}

		// Check release notes for breaking keywords
		fullText := r.Name + " " + r.Body
		if breakingRe.MatchString(fullText) {
			info.Severity = SeverityBreaking
		} else if deprecRe.MatchString(fullText) && info.Severity < SeverityDeprecation {
			info.Severity = SeverityDeprecation
		}

		infos = append(infos, info)
	}

	return infos
}

func SortReports(reports []PluginReport) {
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Severity != reports[j].Severity {
			return reports[i].Severity > reports[j].Severity
		}
		return reports[i].Plugin.Name < reports[j].Plugin.Name
	})
}
