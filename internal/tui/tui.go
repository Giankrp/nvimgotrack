package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Giankrp/nvimgotrack/internal/detector"
	"github.com/Giankrp/nvimgotrack/internal/github"
	"github.com/Giankrp/nvimgotrack/internal/parser"
)

// view is the current screen.
type view int

const (
	viewList view = iota
	viewDetail
)

// filter determines which plugins to show.
type filter int

const (
	filterAll filter = iota
	filterBreaking
	filterDeprecated
	filterBehind
)

// Model is the Bubble Tea model for the TUI.
type Model struct {
	// Data
	plugins  []parser.Plugin
	reports  []detector.PluginReport
	filtered []int // indices into reports
	client   *github.Client

	// UI state
	cursor      int
	view        view
	filter      filter
	width       int
	height      int
	scrollTop   int // for detail view scrolling
	detailLines int

	// Loading
	loading     bool
	loadingIdx  int
	loadingName string
	spinner     spinner.Model
	done        bool
}

type pluginAnalyzed struct {
	index  int
	report detector.PluginReport
}

type allDone struct{}

// NewModel creates a new TUI model.
func NewModel(plugins []parser.Plugin, client *github.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		plugins: plugins,
		reports: make([]detector.PluginReport, len(plugins)),
		client:  client,
		loading: true,
		spinner: s,
		filter:  filterAll,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.analyzeNext(0),
	)
}

func (m Model) analyzeNext(i int) tea.Cmd {
	if i >= len(m.plugins) {
		return func() tea.Msg { return allDone{} }
	}
	plugin := m.plugins[i]
	client := m.client
	return func() tea.Msg {
		report := detector.Analyze(client, plugin)
		return pluginAnalyzed{index: i, report: report}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case pluginAnalyzed:
		m.reports[msg.index] = msg.report
		m.loadingIdx = msg.index + 1
		if msg.index+1 < len(m.plugins) {
			m.loadingName = m.plugins[msg.index+1].Name
		}
		m.applyFilter()
		return m, m.analyzeNext(msg.index + 1)

	case allDone:
		m.loading = false
		m.done = true
		detector.SortReports(m.reports)
		m.applyFilter()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.view == viewDetail {
			m.view = viewList
			m.scrollTop = 0
			return m, nil
		}
		return m, tea.Quit

	case "esc":
		if m.view == viewDetail {
			m.view = viewList
			m.scrollTop = 0
			return m, nil
		}
		return m, nil

	case "j", "down":
		if m.view == viewList {
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		} else {
			m.scrollTop++
		}
		return m, nil

	case "k", "up":
		if m.view == viewList {
			if m.cursor > 0 {
				m.cursor--
			}
		} else {
			if m.scrollTop > 0 {
				m.scrollTop--
			}
		}
		return m, nil

	case "g":
		if m.view == viewList {
			m.cursor = 0
		} else {
			m.scrollTop = 0
		}
		return m, nil

	case "G":
		if m.view == viewList {
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
		}
		return m, nil

	case "enter":
		if m.view == viewList && len(m.filtered) > 0 {
			m.view = viewDetail
			m.scrollTop = 0
			return m, nil
		}
		return m, nil

	case "tab":
		m.filter = (m.filter + 1) % 4
		m.applyFilter()
		m.cursor = 0
		return m, nil

	case "shift+tab":
		m.filter = (m.filter + 3) % 4 // wrap backwards
		m.applyFilter()
		m.cursor = 0
		return m, nil
	}

	return m, nil
}

// applyFilter rebuilds the filtered index list.
func (m *Model) applyFilter() {
	m.filtered = m.filtered[:0]
	for i, r := range m.reports {
		switch m.filter {
		case filterAll:
			m.filtered = append(m.filtered, i)
		case filterBreaking:
			if r.Severity == detector.SeverityBreaking {
				m.filtered = append(m.filtered, i)
			}
		case filterDeprecated:
			if r.Severity >= detector.SeverityDeprecation {
				m.filtered = append(m.filtered, i)
			}
		case filterBehind:
			if r.BehindBy > 0 {
				m.filtered = append(m.filtered, i)
			}
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

// View renders the UI.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title bar
	title := titleStyle.Width(m.width).Render("  ⚡ NvimGoTrack — Plugin Breaking-Change Tracker")
	b.WriteString(title)
	b.WriteString("\n")

	if m.loading {
		b.WriteString(m.viewLoading())
	} else if m.view == viewDetail {
		b.WriteString(m.viewDetailView())
	} else {
		b.WriteString(m.viewListView())
	}

	return b.String()
}

func (m Model) viewLoading() string {
	var b strings.Builder
	progress := fmt.Sprintf("%d/%d", m.loadingIdx, len(m.plugins))
	name := m.loadingName
	if m.loadingIdx < len(m.plugins) {
		name = m.plugins[m.loadingIdx].Name
	}
	b.WriteString(fmt.Sprintf("\n  %s Analyzing plugins... %s\n", m.spinner.View(), progress))
	b.WriteString(fmt.Sprintf("    → %s\n", name))

	// Show already-completed plugins
	b.WriteString("\n")
	for i := 0; i < m.loadingIdx && i < len(m.reports); i++ {
		r := m.reports[i]
		icon := r.Severity.Icon()
		b.WriteString(fmt.Sprintf("  %s %s", icon, r.Plugin.Name))
		if r.Error != "" {
			b.WriteString(errorStyle.Render(" ✗"))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// viewListView renders the plugin list.
func (m Model) viewListView() string {
	var b strings.Builder

	// Summary counts
	var breaking, deprecated, behind, total int
	for _, r := range m.reports {
		total++
		switch r.Severity {
		case detector.SeverityBreaking:
			breaking++
		case detector.SeverityDeprecation:
			deprecated++
		}
		if r.BehindBy > 0 {
			behind++
		}
	}
	summary := fmt.Sprintf("  %s %d breaking  %s %d deprecated  %s %d behind  │  %d plugins total",
		breakingStyle.Render("●"), breaking,
		deprecStyle.Render("●"), deprecated,
		featureStyle.Render("●"), behind,
		total,
	)
	b.WriteString(statusStyle.Render(summary))
	b.WriteString("\n")

	// Filter tabs
	filters := []string{"All", "🔴 Breaking", "🟡 Deprecated", "📦 Behind"}
	var tabs []string
	for i, f := range filters {
		if filter(i) == m.filter {
			tabs = append(tabs, filterActiveStyle.Render(f))
		} else {
			tabs = append(tabs, filterInactiveStyle.Render(f))
		}
	}
	b.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n\n")

	header := fmt.Sprintf("  %-3s %-32s %-12s %-10s %s",
		"", "Plugin", "Commit", "Behind", "Status")
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Bold(true).Render(header))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("  " + strings.Repeat("─", min(m.width-4, 90))))
	b.WriteString("\n")

	listHeight := m.height - 9
	if listHeight < 1 {
		listHeight = 10
	}

	scrollStart := 0
	if m.cursor >= scrollStart+listHeight {
		scrollStart = m.cursor - listHeight + 1
	}
	if m.cursor < scrollStart {
		scrollStart = m.cursor
	}

	for idx := scrollStart; idx < len(m.filtered) && idx < scrollStart+listHeight; idx++ {
		ri := m.filtered[idx]
		r := m.reports[ri]

		icon := r.Severity.Icon()
		name := r.Plugin.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		commit := r.Plugin.Commit
		if len(commit) > 10 {
			commit = commit[:10]
		}

		behindStr := ""
		if r.BehindBy > 0 {
			behindStr = fmt.Sprintf("+%d", r.BehindBy)
		}

		statusStr := ""
		if r.Error != "" {
			statusStr = errorStyle.Render("error")
		} else {
			statusStr = severityLabel(r.Severity)
		}

		line := fmt.Sprintf("  %s %-32s %-12s %-10s %s",
			icon, name, commit, behindStr, statusStr)

		if idx == m.cursor {
			b.WriteString(selectedItemStyle.Width(m.width).Render(line))
		} else {
			b.WriteString(itemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Help bar
	b.WriteString("\n")
	help := "  j/k navigate  •  enter detail  •  tab filter  •  q quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// viewDetailView renders the detail screen for the selected plugin.
func (m Model) viewDetailView() string {
	if len(m.filtered) == 0 {
		return "  No plugin selected"
	}

	ri := m.filtered[m.cursor]
	r := m.reports[ri]
	var b strings.Builder

	// Title
	b.WriteString(detailTitleStyle.Width(m.width - 4).Render(
		fmt.Sprintf("  %s  %s", r.Severity.Icon(), r.Plugin.Name)))
	b.WriteString("\n")

	// Info
	addField := func(label, value string) {
		b.WriteString("  ")
		b.WriteString(detailLabelStyle.Render(label))
		b.WriteString(detailValueStyle.Render(value))
		b.WriteString("\n")
	}

	addField("Repository:", fmt.Sprintf("%s/%s", r.Plugin.Owner, r.Plugin.Repo))
	addField("Branch:", r.Plugin.Branch)
	addField("Current Commit:", r.Plugin.Commit[:min(12, len(r.Plugin.Commit))])
	addField("Behind by:", fmt.Sprintf("%d commits", r.BehindBy))
	addField("Severity:", r.Severity.String())

	if r.CompareURL != "" {
		addField("Compare URL:", r.CompareURL)
	}

	if r.Error != "" {
		b.WriteString("\n")
		b.WriteString("  " + errorStyle.Render("Error: "+r.Error))
		b.WriteString("\n")
	}

	// Breaking changes
	if len(r.BreakingMsgs) > 0 {
		b.WriteString("\n")
		b.WriteString("  " + detailSectionStyle.Render("🔴 Breaking Changes"))
		b.WriteString("\n")
		for _, msg := range r.BreakingMsgs {
			b.WriteString(breakingStyle.Render("    • " + truncate(msg, m.width-8)))
			b.WriteString("\n")
		}
	}

	// Deprecations
	if len(r.DeprecMsgs) > 0 {
		b.WriteString("\n")
		b.WriteString("  " + detailSectionStyle.Render("🟡 Deprecation Warnings"))
		b.WriteString("\n")
		for _, msg := range r.DeprecMsgs {
			b.WriteString(deprecStyle.Render("    • " + truncate(msg, m.width-8)))
			b.WriteString("\n")
		}
	}

	// Releases
	if len(r.Releases) > 0 {
		b.WriteString("\n")
		b.WriteString("  " + detailSectionStyle.Render("📦 Recent Releases"))
		b.WriteString("\n")
		limit := min(10, len(r.Releases))
		for _, rel := range r.Releases[:limit] {
			icon := rel.Severity.Icon()
			tag := releaseTagStyle.Render(rel.Tag)
			name := ""
			if rel.Name != "" && rel.Name != rel.Tag {
				name = " — " + rel.Name
			}
			b.WriteString(fmt.Sprintf("    %s %s%s\n", icon, tag, name))

			// Show first 3 lines of body
			if rel.Body != "" {
				lines := strings.Split(rel.Body, "\n")
				for i, line := range lines {
					if i >= 3 {
						b.WriteString(bodySnippetStyle.Render("..."))
						b.WriteString("\n")
						break
					}
					trimmed := strings.TrimSpace(line)
					if trimmed != "" {
						b.WriteString(bodySnippetStyle.Render(truncate(trimmed, m.width-10)))
						b.WriteString("\n")
					}
				}
			}
		}
	}

	// Help
	b.WriteString("\n")
	help := "  esc/q back  •  j/k scroll"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// severityLabel returns a styled severity label.
func severityLabel(s detector.Severity) string {
	switch s {
	case detector.SeverityBreaking:
		return breakingStyle.Render("BREAKING")
	case detector.SeverityDeprecation:
		return deprecStyle.Render("deprecated")
	case detector.SeverityFeature:
		return featureStyle.Render("updates")
	default:
		return okStyle.Render("up to date")
	}
}

// truncate shortens a string to maxLen.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 80
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
