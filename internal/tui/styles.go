// internal/tui/styles.go

package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Base colors
	primaryColor   = lipgloss.Color("#7B2FBE")
	secondaryColor = lipgloss.Color("#874BFD")
	textColor      = lipgloss.Color("#FAFAFA")
	errorColor     = lipgloss.Color("#FF0000")
	highlightColor = lipgloss.Color("#74B2FF")
	mutedColor     = lipgloss.Color("#626262")

	// Title styling for the main application header
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 1)

	// List styling for the model list container
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1). // Use minimal vertical padding, small horizontal padding
			MarginRight(1) // Add margin for spacing between panels

	// Detail panel styling for model information
	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2)

	// Error message styling
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Background(lipgloss.Color("#2D2D2D")).
			Padding(0, 1).
			MarginBottom(1)

	// Help overlay styles
	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2).
			Background(lipgloss.Color("#1a1a1a")).
			Align(lipgloss.Center)

	// Selected item styling
	selectedStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	// Help text styling
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Tab styling
	tabStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Active tab styling
	activeTabStyle = tabStyle.Copy().
			Foreground(highlightColor).
			Bold(true)

	// Value styling for numbers and measurements
	valueStyle = lipgloss.NewStyle().
			Foreground(highlightColor)

	// Table header styling
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor)
)

// spinnerStyle returns a new spinner style
func spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))
}

// getListWidth returns the appropriate width for the list panel
func getListWidth(totalWidth int) int {
	// Use 30-50% of total width for list, scaling with screen size
	minWidth := 30

	// For small screens (<100 chars), use 50%
	if totalWidth < 100 {
		return max(minWidth, totalWidth/2)
	}

	// For medium screens, scale from 50% down to 30%
	if totalWidth < 160 {
		ratio := 0.5 - (0.2 * float64(totalWidth-100) / 60)
		return max(minWidth, int(float64(totalWidth)*ratio))
	}

	// For large screens, use 30%
	return max(minWidth, totalWidth*3/10)
}

// getDetailWidth returns the appropriate width for the detail panel
func getDetailWidth(totalWidth int) int {
	// Use 60% of total width for details, minus 1 for spacing
	return max(30, totalWidth*3/5-1)
}
