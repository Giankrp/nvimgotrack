package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorBreaking    = lipgloss.Color("#FF4444")
	colorDeprecation = lipgloss.Color("#FFB020")
	colorFeature     = lipgloss.Color("#44DD88")
	colorOK          = lipgloss.Color("#88AACC")
	colorMuted       = lipgloss.Color("#666677")
	colorAccent      = lipgloss.Color("#7C6FFF")
	colorBg          = lipgloss.Color("#1A1B2E")
	colorBgSelected  = lipgloss.Color("#2A2B4E")
	colorWhite       = lipgloss.Color("#E4E4EF")
	colorDim         = lipgloss.Color("#8888AA")

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Background(lipgloss.Color("#12132A")).
			Padding(0, 2).
			MarginBottom(1)

	// Status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	// Plugin list item
	itemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Background(colorBgSelected).
				Bold(true)

	// Severity styles
	breakingStyle = lipgloss.NewStyle().
			Foreground(colorBreaking).
			Bold(true)

	deprecStyle = lipgloss.NewStyle().
			Foreground(colorDeprecation)

	featureStyle = lipgloss.NewStyle().
			Foreground(colorFeature)

	okStyle = lipgloss.NewStyle().
		Foreground(colorOK)

	// Detail view
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				Padding(0, 1).
				MarginBottom(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(colorMuted)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Width(16)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite)

	detailSectionStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				MarginTop(1).
				MarginBottom(0)

	releaseTagStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	bodySnippetStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				PaddingLeft(4)

	// Help
	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	// Loading
	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	// Error message
	errorStyle = lipgloss.NewStyle().
			Foreground(colorBreaking).
			Italic(true)

	// Filter tabs
	filterActiveStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorAccent).
				Padding(0, 2).
				Bold(true)

	filterInactiveStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Padding(0, 2)
)
